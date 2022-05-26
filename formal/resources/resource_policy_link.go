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

func ResourcePolicyLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Linking a Policy to a Role, Group, or Datastore in Formal.",

		CreateContext: resourcePolicyLinkCreate,
		ReadContext:   resourcePolicyLinkRead,
		UpdateContext: resourcePolicyLinkUpdate,
		DeleteContext: resourcePolicyLinkDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "Resource ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"policy_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Policy ID to be linked.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"item_id": {
				// This description is used by the documentation generator and the language server.
				Description: "User, Group, or Datastore ID that should be linked. NOTE: deleting one of these item types will delete all policy links between policies and that item. The policies are not deleted.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of item that should be linked. Possible values are `role`, `group`, and `datastore`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// "active": {
			// 	// This description is used by the documentation generator and the language server.
			// 	Description: "created_at",
			// 	Type:        schema.TypeBool,
			// 	Computed:    true,
			// },
			"expire_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When the policy should expire.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourcePolicyLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newPolicyLink := api.PolicyLinkStruct{
		PolicyID: d.Get("policy_id").(string),
		ItemID:   d.Get("item_id").(string),
		Type:     d.Get("type").(string),
	}

	policyLink, err := client.CreatePolicyLink(newPolicyLink)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policyLink.ID)

	resourcePolicyLinkRead(ctx, d, meta)
	return diags
}

func resourcePolicyLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)
	var diags diag.Diagnostics

	policyLinkId := d.Id()

	policyLink, err := client.GetPolicyLink(policyLinkId)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Link was deleted
			tflog.Warn(ctx, "The Policy-Item link with ID "+policyLink.ID+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if policyLink == nil {
		return diags
	}

	// Should map to all fields of
	d.Set("id", policyLink.ID)
	d.Set("policy_id", policyLink.PolicyID)
	d.Set("item_id", policyLink.ItemID)
	d.Set("type", policyLink.Type)
	d.Set("expire_at", policyLink.ExpireAt)

	d.SetId(policyLinkId)

	return diags
}

func resourcePolicyLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Policy Links are immutable. Please create a new policyLink.")
}

func resourcePolicyLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	var diags diag.Diagnostics

	policyLinkId := d.Id()

	err := client.DeletePolicyLink(ctx, policyLinkId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
