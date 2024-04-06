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
	DatastoreId := d.Get("resource_id").(string)
	Username := d.Get("username").(string)
	Secret := d.Get("secret").(string)
	UseAsDefault := d.Get("use_as_default").(bool)
	TerminationProtection := d.Get("termination_protection").(bool)

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateNativeUser(ctx, connect.NewRequest(&corev1.CreateNativeUserRequest{
		ResourceId:            DatastoreId,
		Username:              Username,
		Secret:                Secret,
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

	id := d.Get("id").(string)

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetNativeUser(ctx, connect.NewRequest(&corev1.GetNativeUserRequest{Id: id}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Policy was deleted
			tflog.Warn(ctx, "The Native Role "+id+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of RoleOrgItem
	d.Set("resource_id", res.Msg.NativeUser.ResourceId)
	d.Set("native_role_username", res.Msg.NativeUser.Username)
	d.Set("native_role_secret", res.Msg.NativeUser.Secret)
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

	id := d.Id()

	if d.HasChange("use_as_default") {
		useAsDefault := d.Get("use_as_default").(bool)
		if useAsDefault {
			_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateNativeUser(ctx, connect.NewRequest(&corev1.UpdateNativeUserRequest{
				Id:           id,
				UseAsDefault: &useAsDefault,
			}))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if d.HasChange("secret") {
		secret := d.Get("secret").(string)
		_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateNativeUser(ctx, connect.NewRequest(&corev1.UpdateNativeUserRequest{
			Id:     id,
			Secret: &secret,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateNativeUser(ctx, connect.NewRequest(&corev1.UpdateNativeUserRequest{
			Id:                    id,
			TerminationProtection: &terminationProtection,
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

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Native Role cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteNativeUser(ctx, connect.NewRequest(&corev1.DeleteNativeUserRequest{Id: d.Id()}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
