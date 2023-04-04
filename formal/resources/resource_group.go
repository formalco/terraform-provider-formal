package resource

import (
	"context"
	"fmt"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"strings"

	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newGroup := api.GroupStruct{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	group, err := c.Http.CreateGroup(newGroup)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(group.ID)

	resourceGroupRead(ctx, d, meta)
	return diags
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	groupId := d.Id()

	group, err := c.Http.GetGroup(groupId)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Group was deleted
			tflog.Warn(ctx, "The Group was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
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
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	groupId := d.Id()
	groupName := d.Get("name").(string)
	groupDesc := d.Get("description").(string)

	err := c.Http.UpdateGroup(groupId, api.GroupStruct{Name: groupName, Description: groupDesc})
	if err != nil {
		return diag.FromErr(err)
	}

	resourceGroupRead(ctx, d, meta)

	return diags
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	groupId := d.Id()

	err := c.Http.DeleteGroup(groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
