package resource

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceOrgTheme() *schema.Resource {
	logoTypes := orgThemeStringIn("logo_content_type")
	return &schema.Resource{
		Description: "The organization's display name and logo, shown to its users in the Catalog.",

		CreateContext: resourceOrgThemeCreate,
		ReadContext:   resourceOrgThemeRead,
		UpdateContext: resourceOrgThemeUpdate,
		DeleteContext: resourceOrgThemeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the theme.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"display_name": {
				Description:  "The name shown in the Catalog header and browser tab. At most 60 characters.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: orgThemeValidateString("display_name"),
			},
			"logo": {
				Description:  "The base64-encoded logo shown in place of the Formal mark and as the favicon, for example `filebase64(\"logo.svg\")`. Must be square, at most 256 KiB.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateOrgThemeLogo,
			},
			"logo_content_type": {
				Description:  fmt.Sprintf("The image type of `logo`. One of %s.", orgThemeList(logoTypes)),
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: orgThemeValidateString("logo_content_type"),
			},
			"created_at": {
				Description: "When the theme was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "When the theme was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// orgThemeFieldRules reads a field's rules from the API proto, so the provider
// accepts exactly what the API accepts.
func orgThemeFieldRules(field string) *validate.FieldRules {
	descriptor := (&corev1.CreateOrgThemeRequest{}).ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name(field))
	if descriptor == nil {
		panic(fmt.Sprintf("org theme field %s does not exist", field))
	}
	rules, ok := proto.GetExtension(descriptor.Options(), validate.E_Field).(*validate.FieldRules)
	if !ok {
		panic(fmt.Sprintf("org theme field %s has no validation rules", field))
	}
	return rules
}

func orgThemeStringIn(field string) []string {
	return orgThemeFieldRules(field).GetString_().GetIn()
}

func orgThemeList(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = "`" + value + "`"
	}
	return strings.Join(quoted, ", ")
}

// orgThemeValidateString checks the allow-list and length limits the proto
// declares for a field.
func orgThemeValidateString(field string) schema.SchemaValidateFunc {
	rules := orgThemeFieldRules(field).GetString_()
	return func(value any, key string) ([]string, []error) {
		text, ok := value.(string)
		if !ok {
			return nil, []error{fmt.Errorf("expected %s to be a string", key)}
		}
		if in := rules.GetIn(); len(in) > 0 && !slices.Contains(in, text) {
			return nil, []error{fmt.Errorf("%s must be one of %s", key, strings.Join(in, ", "))}
		}
		if length := uint64(len([]rune(text))); length < rules.GetMinLen() || (rules.GetMaxLen() > 0 && length > rules.GetMaxLen()) {
			return nil, []error{fmt.Errorf("%s must be between %d and %d characters", key, rules.GetMinLen(), rules.GetMaxLen())}
		}
		return nil, nil
	}
}

func validateOrgThemeLogo(value any, key string) ([]string, []error) {
	text, ok := value.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected %s to be a string", key)}
	}
	logo, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, []error{fmt.Errorf("%s must be base64-encoded: %w", key, err)}
	}
	rules := orgThemeFieldRules("logo").GetBytes()
	if length := uint64(len(logo)); length < rules.GetMinLen() || length > rules.GetMaxLen() {
		return nil, []error{fmt.Errorf("%s must be between %d and %d bytes", key, rules.GetMinLen(), rules.GetMaxLen())}
	}
	return nil, nil
}

type orgThemeArgs struct {
	displayName     string
	logo            []byte
	logoContentType string
}

func expandOrgTheme(d *schema.ResourceData) (orgThemeArgs, error) {
	logo, err := base64.StdEncoding.DecodeString(d.Get("logo").(string))
	if err != nil {
		return orgThemeArgs{}, fmt.Errorf("logo must be base64-encoded: %w", err)
	}
	return orgThemeArgs{
		displayName:     d.Get("display_name").(string),
		logo:            logo,
		logoContentType: d.Get("logo_content_type").(string),
	}, nil
}

func flattenOrgTheme(d *schema.ResourceData, theme *corev1.OrgTheme) error {
	d.SetId(theme.Id)
	values := map[string]any{
		"display_name":      theme.DisplayName,
		"logo":              base64.StdEncoding.EncodeToString(theme.Logo),
		"logo_content_type": theme.LogoContentType,
		"created_at":        theme.CreatedAt.AsTime().Format(time.RFC3339),
		"updated_at":        theme.UpdatedAt.AsTime().Format(time.RFC3339),
	}
	for key, value := range values {
		if err := d.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func resourceOrgThemeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	args, err := expandOrgTheme(d)
	if err != nil {
		return diag.FromErr(err)
	}
	response, err := c.Grpc.Sdk.OrgThemeServiceClient.CreateOrgTheme(ctx, &corev1.CreateOrgThemeRequest{
		DisplayName:     args.displayName,
		Logo:            args.logo,
		LogoContentType: args.logoContentType,
	})
	if connect.CodeOf(err) == connect.CodeAlreadyExists {
		return diag.FromErr(errors.New("the organization already has a theme: import it with `terraform import formal_org_theme.<name> <theme_id>`"))
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if err := flattenOrgTheme(d, response.Theme); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceOrgThemeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	response, err := c.Grpc.Sdk.OrgThemeServiceClient.GetOrgTheme(ctx, &corev1.GetOrgThemeRequest{})
	if connect.CodeOf(err) == connect.CodeNotFound {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if err := flattenOrgTheme(d, response.Theme); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceOrgThemeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	args, err := expandOrgTheme(d)
	if err != nil {
		return diag.FromErr(err)
	}
	response, err := c.Grpc.Sdk.OrgThemeServiceClient.UpdateOrgTheme(ctx, &corev1.UpdateOrgThemeRequest{
		DisplayName:     args.displayName,
		Logo:            args.logo,
		LogoContentType: args.logoContentType,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if err := flattenOrgTheme(d, response.Theme); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceOrgThemeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	_, err := c.Grpc.Sdk.OrgThemeServiceClient.DeleteOrgTheme(ctx, &corev1.DeleteOrgThemeRequest{})
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
