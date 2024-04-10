package resource

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
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
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Group cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
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
	TerminationProtection := d.Get("termination_protection").(bool)

	res, err := c.Grpc.Sdk.GroupServiceClient.CreateGroup(ctx, connect.NewRequest(&corev1.CreateGroupRequest{
		Name:                  Name,
		Description:           Description,
		TerminationProtection: TerminationProtection,
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

	res, err := c.Grpc.Sdk.GroupServiceClient.GetGroup(ctx, connect.NewRequest(&corev1.GetGroupRequest{Id: groupId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
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
	d.Set("termination_protection", res.Msg.Group.TerminationProtection)

	d.SetId(groupId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	groupId := d.Id()
	groupName := d.Get("name").(string)
	//groupDesc := d.Get("description").(string)
	groupTermProtection := d.Get("termination_protection").(bool)

	_, err := c.Grpc.Sdk.GroupServiceClient.UpdateGroup(ctx, connect.NewRequest(&corev1.UpdateGroupRequest{Name: &groupName, Id: groupId, TerminationProtection: &groupTermProtection}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceGroupRead(ctx, d, meta)

	return diags
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Group cannot be deleted because termination_protection is set to true")
	}

	groupId := d.Id()
	_, err := c.Grpc.Sdk.GroupServiceClient.DeleteGroup(ctx, connect.NewRequest(&corev1.DeleteGroupRequest{Id: groupId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
