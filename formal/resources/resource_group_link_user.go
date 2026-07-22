package resource

import (
	"context"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceGroupLinkUser() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Linking a User to a Group in Formal.",

		CreateContext: resourceGroupLinkUserCreate,
		ReadContext:   resourceGroupLinkUserRead,
		UpdateContext: resourceGroupLinkUserUpdate,
		DeleteContext: resourceGroupLinkUserDelete,
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

func resourceGroupLinkUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	userId := d.Get("user_id").(string)
	groupId := d.Get("group_id").(string)

	res, err := c.Grpc.Sdk.GroupServiceClient.CreateUserGroupLink(ctx, &corev1.CreateUserGroupLinkRequest{GroupId: groupId, UserId: userId})
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("user_id", res.UserGroupLink.User.Id)
	d.Set("group_id", res.UserGroupLink.Group.Id)

	d.SetId(res.UserGroupLink.Id)

	resourceGroupLinkUserRead(ctx, d, meta)
	return diags
}

func resourceGroupLinkUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	groupLinkId := d.Id()
	// Maps to user-defined fields
	userId := d.Get("user_id").(string)
	groupId := d.Get("group_id").(string)

	res, err := c.Grpc.Sdk.GroupServiceClient.ListUserGroupLinks(ctx, &corev1.ListUserGroupLinksRequest{GroupId: groupId, Limit: 500})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Link was deleted
			tflog.Warn(ctx, "The Group-User link was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	found := false
	for _, user := range res.UserGroupLinks {
		if user.User.Id == groupLinkId {
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

	d.SetId(groupLinkId)

	return diags
}

func resourceGroupLinkUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return diag.Errorf("Group User Links are immutable. Please create a new roleLinkGroup.")
}

func resourceGroupLinkUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	linkId := d.Id()

	_, err := c.Grpc.Sdk.GroupServiceClient.DeleteUserGroupLink(ctx, &corev1.DeleteUserGroupLinkRequest{Id: linkId})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
