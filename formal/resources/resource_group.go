package resource

import (
	"context"

	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Group in Formal.",

		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for this Group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly Name for this Group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				// This description is used by the documentation generator and the language server.
				Description: "Description for this Group.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newGroup := api.GroupStruct{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	group, err := client.CreateGroup(newGroup)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(group.ID)

	resourceGroupRead(ctx, d, meta)
	return diags
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)
	var diags diag.Diagnostics

	groupId := d.Id()

	group, err := client.GetGroup(groupId)
	if err != nil {
		return diag.FromErr(err)
	}
	if group == nil {
		return diags
	}

	// Should map to all fields of
	d.Set("id", group.ID)
	d.Set("name", group.Name)
	d.Set("description", group.Description)

	d.SetId(groupId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Policy Links are immutable. Please create a new group.")
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	var diags diag.Diagnostics

	groupId := d.Id()

	err := client.DeleteGroup(groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
