package resource

import (
	"context"
	"errors"
	"fmt"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceGroupLinkRole() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Linking a Role to a Group in Formal.",

		CreateContext: resourceGroupLinkRoleCreate,
		ReadContext:   resourceGroupLinkRoleRead,
		// UpdateContext: resourceGroupLinkRoleUpdate,
		DeleteContext: resourceGroupLinkRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID of this link.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"role_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID of the role to be linked.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for the group to be linked.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

const roleLinkGroupTerraformIdDelimiter = "#_#"

func resourceGroupLinkRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	roleId := d.Get("role_id").(string)
	groupId := d.Get("group_id").(string)

	err := c.Http.CreateGroupLinkRole(roleId, groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	terraformResourceId := groupId + roleLinkGroupTerraformIdDelimiter + roleId
	d.SetId(terraformResourceId)

	resourceGroupLinkRoleRead(ctx, d, meta)
	return diags
}

func resourceGroupLinkRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	roleLinkGroupTerraformId := d.Id()
	// Split
	roleLinkGroupTerraformIdSplit := strings.Split(roleLinkGroupTerraformId, roleLinkGroupTerraformIdDelimiter)
	if len(roleLinkGroupTerraformIdSplit) != 2 {
		return diag.FromErr(errors.New("formal Terraform resource id for role_link_group is malformatted. Please contact Formal support"))
	}
	groupId := roleLinkGroupTerraformIdSplit[0]
	roleId := roleLinkGroupTerraformIdSplit[1]

	roleLinkGroup, err := c.Http.GetGroupLinkRole(roleId, groupId)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Link was deleted
			tflog.Warn(ctx, "The Group-User link was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if roleLinkGroup == "" {
		// Not found
		return diags
	}

	// Should map to all fields of
	d.Set("group_id", groupId)
	d.Set("role_id", roleId)

	d.SetId(roleLinkGroupTerraformId)

	return diags
}

// func resourceGroupLinkRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	return diag.Errorf("Group User Links are immutable. Please create a new roleLinkGroup.")
// }

func resourceGroupLinkRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	roleLinkGroupTerraformId := d.Id()
	// Split
	roleLinkGroupTerraformIdSplit := strings.Split(roleLinkGroupTerraformId, roleLinkGroupTerraformIdDelimiter)
	if len(roleLinkGroupTerraformIdSplit) != 2 {
		return diag.FromErr(errors.New("formal Terraform resource id for role_link_group is malformatted. Please contact Formal support"))
	}
	groupId := roleLinkGroupTerraformIdSplit[0]
	roleId := roleLinkGroupTerraformIdSplit[1]

	err := c.Http.DeleteGroupLinkRole(roleId, groupId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
