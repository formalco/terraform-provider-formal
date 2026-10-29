package resource

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.yaml.in/yaml/v3"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

const skillListPageSize = 500

func ResourceSkill() *schema.Resource {
	return &schema.Resource{
		Description: "Registering a Skill in the Formal Catalog from its `SKILL.md`. Skill names are unique, and changing `skill_md` replaces the skill.",

		CreateContext: resourceSkillCreate,
		ReadContext:   resourceSkillRead,
		DeleteContext: resourceSkillDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSkillImport,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this Skill.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"skill_md": {
				Description:      "Contents of the `SKILL.md` file, including its `name` and `description` frontmatter. Formatting-only changes are ignored.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentSkillMD,
			},
			"name": {
				Description: "Name of the Skill, from the `SKILL.md` frontmatter.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description of the Skill, from the `SKILL.md` frontmatter.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSkillCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	res, err := c.Grpc.Sdk.SkillServiceClient.CreateFormalSkill(ctx, &corev1.CreateFormalSkillRequest{
		Source: &corev1.CreateFormalSkillRequest_SkillMd{SkillMd: d.Get("skill_md").(string)},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Skill.Id)
	d.Set("name", res.Skill.Name)

	return resourceSkillRead(ctx, d, meta)
}

func resourceSkillRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	name := d.Get("name").(string)
	skill, found, err := findSkillByName(ctx, c, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if !found {
		tflog.Warn(ctx, "The Skill "+name+" was not found, which means it may have been deleted without using this Terraform config.")
		d.SetId("")
		return nil
	}

	d.SetId(skill.Id)
	d.Set("name", skill.Name)
	d.Set("description", skill.Description)

	// The API doesn't return the original file. Keep the configured one unless it drifted.
	remote := parsedSkillMD{Name: skill.Name, Description: skill.Description, Contents: skill.Contents}
	local, err := parseSkillMD(d.Get("skill_md").(string))
	if err != nil || local != remote {
		d.Set("skill_md", renderSkillMD(remote))
	}

	return nil
}

func resourceSkillDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.SkillServiceClient.DeleteFormalSkill(ctx, &corev1.DeleteFormalSkillRequest{Name: d.Get("name").(string)})
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// The API addresses skills by name, so skills are imported by name.
func resourceSkillImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	name := d.Id()
	d.Set("name", name)

	if diags := resourceSkillRead(ctx, d, meta); diags.HasError() {
		return nil, errors.New(diags[0].Summary)
	}
	if d.Id() == "" {
		return nil, errors.New("skill " + name + " not found")
	}
	return []*schema.ResourceData{d}, nil
}

// The name filter matches substrings, so look for the exact name.
func findSkillByName(ctx context.Context, c *clients.Clients, name string) (*corev1.SkillVariant, bool, error) {
	cursor := ""
	for {
		res, err := c.Grpc.Sdk.SkillServiceClient.ListSkills(ctx, &corev1.ListSkillsRequest{
			Limit:      skillListPageSize,
			Cursor:     cursor,
			NameFilter: name,
		})
		if err != nil {
			return nil, false, err
		}
		for _, skill := range res.Skills {
			if skill.Name == name {
				return skill, true, nil
			}
		}
		if res.NextCursor == "" {
			return nil, false, nil
		}
		cursor = res.NextCursor
	}
}

type parsedSkillMD struct {
	Name        string
	Description string
	Contents    string
}

// Must match how the control plane parses SKILL.md.
func parseSkillMD(raw string) (parsedSkillMD, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return parsedSkillMD{}, errors.New("SKILL.md must start with YAML frontmatter")
	}
	end := strings.Index(normalized[4:], "\n---\n")
	if end < 0 {
		return parsedSkillMD{}, errors.New("SKILL.md frontmatter is not closed")
	}
	var frontmatter struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(normalized[4:4+end]), &frontmatter); err != nil {
		return parsedSkillMD{}, err
	}
	return parsedSkillMD{
		Name:        strings.TrimSpace(frontmatter.Name),
		Description: strings.TrimSpace(frontmatter.Description),
		Contents:    normalized[4+end+5:],
	}, nil
}

func renderSkillMD(skill parsedSkillMD) string {
	frontmatter, err := yaml.Marshal(struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}{Name: skill.Name, Description: skill.Description})
	if err != nil {
		return ""
	}
	return "---\n" + string(frontmatter) + "---\n" + skill.Contents
}

func suppressEquivalentSkillMD(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	oldSkill, err := parseSkillMD(oldValue)
	if err != nil {
		return false
	}
	newSkill, err := parseSkillMD(newValue)
	if err != nil {
		return false
	}
	return oldSkill == newSkill
}
