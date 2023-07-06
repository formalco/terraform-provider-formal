package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	Name := d.Get("name").(string)
	Description := d.Get("description").(string)

	res, err := c.Grpc.Sdk.GroupServiceClient.CreateGroup(ctx, connect.NewRequest(&adminv1.CreateGroupRequest{
		Name:        Name,
		Description: Description,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Group.Id)

	resourceGroupRead(ctx, d, meta)
	return diags
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	groupId := d.Id()

	res, err := c.Grpc.Sdk.GroupServiceClient.GetGroupById(ctx, connect.NewRequest(&adminv1.GetGroupByIdRequest{Id: groupId}))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Group was deleted
			tflog.Warn(ctx, "The Group was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of
	d.Set("id", res.Msg.Group.Id)
	d.Set("name", res.Msg.Group.Name)
	d.Set("description", res.Msg.Group.Description)

	d.SetId(groupId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	groupId := d.Id()
	groupName := d.Get("name").(string)
	//groupDesc := d.Get("description").(string)

	_, err := c.Grpc.Sdk.GroupServiceClient.UpdateGroup(ctx, connect.NewRequest(&adminv1.UpdateGroupRequest{Name: groupName, Id: groupId}))
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
	_, err := c.Grpc.Sdk.GroupServiceClient.DeleteGroup(ctx, connect.NewRequest(&adminv1.DeleteGroupRequest{Id: groupId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
