package resource

import (
	"context"
	"errors"
	"strings"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"

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
		UpdateContext: resourceGroupLinkRoleUpdate,
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
			"user_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID of the user to be linked.",
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
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Link cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
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
	userId := d.Get("user_id").(string)
	groupId := d.Get("group_id").(string)

	res, err := c.Grpc.Sdk.GroupServiceClient.CreateUserGroupLink(ctx, connect.NewRequest(&corev1.CreateUserGroupLinkRequest{GroupId: groupId, UserId: userId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.UserGroupLink.Id)

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
	userId := roleLinkGroupTerraformIdSplit[1]
	res, err := c.Grpc.Sdk.GroupServiceClient.ListUserGroupLinks(ctx, connect.NewRequest(&corev1.ListUserGroupLinksRequest{GroupId: groupId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Link was deleted
			tflog.Warn(ctx, "The Group-User link was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	found := false
	for _, user := range res.Msg.UserGroupLinks {
		if user.User.Id == userId {
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
	d.Set("user_id", userId)

	d.SetId(roleLinkGroupTerraformId)

	return diags
}

func resourceGroupLinkRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Group User Links are immutable. Please create a new roleLinkGroup.")
}

func resourceGroupLinkRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	linkId := d.Id()

	_, err := c.Grpc.Sdk.GroupServiceClient.DeleteUserGroupLink(ctx, connect.NewRequest(&corev1.DeleteUserGroupLinkRequest{Id: linkId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
