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

func ResourceNativeRoleLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "This resource creates assigns a Native Role to a Formal Identity.",

		CreateContext: resourceNativeRoleLinkCreate,
		ReadContext:   resourceNativeRoleLinkRead,
		UpdateContext: resourceNativeRoleLinkUpdate,
		DeleteContext: resourceNativeRoleLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Resource ID of the Native Role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"native_role_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Native Role ID of the Native Role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"formal_identity_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for the Role or Group to be linked.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"formal_identity_type": {
				// This description is used by the documentation generator and the language server.
				Description: "The type of Formal Identity to be linked. Accepted values are `role` and `group`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Native Role link cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceNativeRoleLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	nativeRoleId := d.Get("native_role_id").(string)
	formalIdentityId := d.Get("formal_identity_id").(string)
	formalIdentityType := d.Get("formal_identity_type").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateNativeUserIdentityLink(ctx, connect.NewRequest(&corev1.CreateNativeUserIdentityLinkRequest{
		NativeUserId:          nativeRoleId,
		IdentityId:            formalIdentityId,
		IdentityType:          formalIdentityType,
		TerminationProtection: terminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Link.Id)

	resourceNativeRoleLinkRead(ctx, d, meta)
	return diags
}

func resourceNativeRoleLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateNativeUserIdentityLink(ctx, connect.NewRequest(&corev1.UpdateNativeUserIdentityLinkRequest{
			Id:                    d.Id(),
			TerminationProtection: &terminationProtection,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceNativeRoleLinkRead(ctx, d, meta)

	return diags
}

func resourceNativeRoleLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	nativeUserIdentityId := d.Id()

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetNativeUserIdentityLink(ctx, connect.NewRequest(&corev1.GetNativeUserIdentityLinkRequest{
		Id: nativeUserIdentityId,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Link was deleted
			tflog.Warn(ctx, "The Native Role Link was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	switch info := res.Msg.Link.Identity.(type) {
	case *corev1.NativeUserLink_User:
		d.Set("formal_identity_id", info.User.Id)
	case *corev1.NativeUserLink_Group:
		d.Set("formal_identity_id", info.Group.Id)
	}

	// Should map to all fields of
	d.Set("resource_id", res.Msg.Link.NativeUser.ResourceId)
	d.Set("native_role_id", res.Msg.Link.NativeUser.Id)
	d.Set("formal_identity_type", res.Msg.Link.Identity)
	d.Set("termination_protection", res.Msg.Link.TerminationProtection)

	d.SetId(res.Msg.Link.Id)

	return diags
}

func resourceNativeRoleLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()
	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Native Role Link cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteNativeUserIdentityLink(ctx, connect.NewRequest(&corev1.DeleteNativeUserIdentityLinkRequest{Id: id}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
