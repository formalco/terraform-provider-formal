package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Policy in Formal.",

		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Policy Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				// This description is used by the documentation generator and the language server.
				Description: "Policy Description.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"module": {
				// This description is used by the documentation generator and the language server.
				Description: "The module describing how the policy works. Create one in the Formal Console.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of this Policy.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				// This description is used by the documentation generator and the language server.
				Description: "Who the policy was created by.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When the policy was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Last update time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"active": {
				// This description is used by the documentation generator and the language server.
				Description: "Active status of this policy.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"org_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for your organisation.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expire_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When this policy is set to expire.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				// This description is used by the documentation generator and the language server.
				Description: "Additional descriptor for active status of this policy.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newPolicy := api.CreatePolicyPayload{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Module:      d.Get("module").(string),
		SourceType:  "terraform",
	}

	policy, err := client.CreatePolicy(ctx, newPolicy)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policy.ID)

	resourcePolicyRead(ctx, d, meta)
	return diags
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)
	var diags diag.Diagnostics

	policyId := d.Id()

	policy, err := client.GetPolicy(policyId)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Policy was deleted
			tflog.Warn(ctx, "The Policy with ID "+policyId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if policy == nil {
		return diags
	}

	// Should map to all fields of PolicyOrgItem
	d.Set("id", policy.ID)
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("created_by", policy.CreatedBy)
	d.Set("created_at", policy.CreatedBy)
	d.Set("updated_at", policy.UpdatedAt)
	d.Set("module", policy.Module)
	d.Set("active", policy.Active)
	d.Set("org_id", policy.OrganisationID)
	d.Set("expire_at", policy.ExpireAt)
	d.Set("status", policy.Status)

	d.SetId(policyId)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	policyId := d.Id()

	policyUpdate := api.PolicyOrgItem{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Module:      d.Get("module").(string),
		SourceType:  "terraform",
	}

	err := client.UpdatePolicy(policyId, policyUpdate)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	var diags diag.Diagnostics

	policyId := d.Id()

	err := client.DeletePolicy(policyId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
