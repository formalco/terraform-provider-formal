package resource

import (
	"context"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
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
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Sidecar ID of the Native Role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
	datastoreId := d.Get("datastore_id").(string)
	nativeRoleId := d.Get("native_role_id").(string)
	formalIdentityId := d.Get("formal_identity_id").(string)
	formalIdentityType := d.Get("formal_identity_type").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	res, err := c.Grpc.Sdk.NativeUserServiceClient.CreateNativeUserIdentityLinkV2(ctx, connect.NewRequest(&adminv1.CreateNativeUserIdentityLinkV2Request{
		DataStoreId:           datastoreId,
		NativeUserId:          nativeRoleId,
		IdentityId:            formalIdentityId,
		FormalIdentityType:    formalIdentityType,
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
		_, err := c.Grpc.Sdk.NativeUserServiceClient.UpdateNativeUserIdentityLink(ctx, connect.NewRequest(&adminv1.UpdateNativeUserIdentityLinkRequest{
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

	datastoreId := d.Get("datastore_id").(string)
	formalIdentityId := d.Get("formal_identity_id").(string)

	res, err := c.Grpc.Sdk.NativeUserServiceClient.GetNativeUserIdentityLink(ctx, connect.NewRequest(&adminv1.GetNativeUserIdentityLinkRequest{
		DataStoreId: datastoreId,
		IdentityId:  formalIdentityId,
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

	if res.Msg.Link.FormalIdentityType == "role" {
		res.Msg.Link.FormalIdentityType = "user"
	}

	// Should map to all fields of
	d.Set("datastore_id", res.Msg.Link.DataStoreId)
	d.Set("native_role_id", res.Msg.Link.NativeUserId)
	d.Set("formal_identity_id", res.Msg.Link.FormalIdentityId)
	d.Set("formal_identity_type", res.Msg.Link.FormalIdentityType)
	d.Set("termination_protection", res.Msg.Link.TerminationProtection)

	d.SetId(res.Msg.Link.Id)

	return diags
}

func resourceNativeRoleLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	datastoreId := d.Get("datastore_id").(string)
	formalIdentityId := d.Get("formal_identity_id").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Native Role Link cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.NativeUserServiceClient.DeleteNativeUserIdentityLink(ctx, connect.NewRequest(&adminv1.DeleteNativeUserIdentityLinkRequest{DataStoreId: datastoreId, IdentityId: formalIdentityId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
