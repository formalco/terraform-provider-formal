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

func ResourceNativeRole() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "This resource creates a Native Role.",

		CreateContext: resourceNativeRoleCreate,
		ReadContext:   resourceNativeRoleRead,
		UpdateContext: resourceNativeRoleUpdate,
		DeleteContext: resourceNativeRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Sidecar ID for the datastore this Native Role is for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"native_role_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The username of the Native Role.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"native_role_secret": {
				// This description is used by the documentation generator and the language server.
				Description: "The password of the Native Role.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"use_as_default": {
				// This description is used by the documentation generator and the language server.
				Description: "The password of the Native Role.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Native Role cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceNativeRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	DatastoreId := d.Get("datastore_id").(string)
	NativeRoleId := d.Get("native_role_id").(string)
	NativeRoleSecret := d.Get("native_role_secret").(string)
	UseAsDefault := d.Get("use_as_default").(bool)
	TerminationProtection := d.Get("termination_protection").(bool)

	res, err := c.Grpc.Sdk.NativeUserServiceClient.CreateNativeUser(ctx, connect.NewRequest(&adminv1.CreateNativeUserRequest{
		DataStoreId:           DatastoreId,
		NativeUserId:          NativeRoleId,
		NativeUserSecret:      NativeRoleSecret,
		UseAsDefault:          UseAsDefault,
		TerminationProtection: TerminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.NativeUser.Id)

	resourceNativeRoleRead(ctx, d, meta)

	return diags
}

func resourceNativeRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	datastoreId := d.Get("datastore_id").(string)
	nativeRoleId := d.Get("native_role_id").(string)

	res, err := c.Grpc.Sdk.NativeUserServiceClient.GetNativeUser(ctx, connect.NewRequest(&adminv1.GetNativeUserRequest{DataStoreId: datastoreId, NativeUserId: nativeRoleId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Policy was deleted
			tflog.Warn(ctx, "The Native Role for Datastore ID "+datastoreId+" and Native Role ID"+nativeRoleId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of RoleOrgItem
	d.Set("datastore_id", res.Msg.NativeUser.DatastoreId)
	d.Set("native_role_id", res.Msg.NativeUser.NativeUserId)
	d.Set("native_role_secret", res.Msg.NativeUser.NativeUserSecret)
	d.Set("use_as_default", res.Msg.NativeUser.UseAsDefault)
	d.Set("termination_protection", res.Msg.NativeUser.TerminationProtection)

	d.SetId(res.Msg.NativeUser.Id)

	return diags
}

func resourceNativeRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	// if d.HasChangesExcept("use_as_default", "native_role_secret") {
	// 	return diag.Errorf("Native Roles can only be updated for use_as_default and native_role_secret. Please create a new Native Role.")
	// }

	datastoreId := d.Get("datastore_id").(string)
	nativeRoleId := d.Get("native_role_id").(string)

	if d.HasChange("use_as_default") {
		useAsDefault := d.Get("use_as_default").(bool)
		if useAsDefault {
			_, err := c.Grpc.Sdk.NativeUserServiceClient.SetNativeUserAsDefault(ctx, connect.NewRequest(&adminv1.SetNativeUserAsDefaultRequest{
				DataStoreId:  datastoreId,
				NativeUserId: nativeRoleId,
			}))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if d.HasChange("native_role_secret") {
		nativeRoleSecret := d.Get("native_role_secret").(string)
		_, err := c.Grpc.Sdk.NativeUserServiceClient.UpdateNativeUserSecret(ctx, connect.NewRequest(&adminv1.UpdateNativeUserSecretRequest{
			DataStoreId:      datastoreId,
			NativeUserId:     nativeRoleId,
			NativeUserSecret: nativeRoleSecret,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		_, err := c.Grpc.Sdk.NativeUserServiceClient.SetNativeUserTerminationProtection(ctx, connect.NewRequest(&adminv1.SetNativeUserTerminationProtectionRequest{
			Id:                    d.Id(),
			TerminationProtection: terminationProtection,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceNativeRoleRead(ctx, d, meta)

	return diags
}

func resourceNativeRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	datastoreId := d.Get("datastore_id").(string)
	nativeRoleId := d.Get("native_role_id").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Native Role cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.NativeUserServiceClient.DeleteNativeUser(ctx, connect.NewRequest(&adminv1.DeleteNativeUserRequest{DataStoreId: datastoreId, NativeUserId: nativeRoleId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
