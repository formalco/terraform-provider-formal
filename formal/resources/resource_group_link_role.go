package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"errors"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	_, err := c.Grpc.Sdk.GroupServiceClient.LinkUsersToGroup(ctx, connect.NewRequest(&adminv1.LinkUsersToGroupRequest{Id: groupId, UserIds: []string{roleId}}))
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
	res, err := c.Grpc.Sdk.GroupServiceClient.GetGroupById(ctx, connect.NewRequest(&adminv1.GetGroupByIdRequest{Id: groupId}))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Link was deleted
			tflog.Warn(ctx, "The Group-User link was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	found := false
	for _, userId := range res.Msg.Group.UserIds {
		if userId == roleId {
			found = true
			break
		}
	}

	if !found {
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

	_, err := c.Grpc.Sdk.GroupServiceClient.UnlinkUsersFromGroup(ctx, connect.NewRequest(&adminv1.UnlinkUsersFromGroupRequest{Id: groupId, UserIds: []string{roleId}}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
