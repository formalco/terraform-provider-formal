package resource

import (
	"context"
	"reflect"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

const (
	ValidUserTypes = "'password', 'iam', 'k8s'"
	ValidIAMTypes  = "'aws', 'gcp'"
)

func ResourceNativeUser() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "This resource creates a Native User.",

		CreateContext: resourceNativeUserCreate,
		ReadContext:   resourceNativeUserRead,
		UpdateContext: resourceNativeUserUpdate,
		DeleteContext: resourceNativeUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Native User.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Sidecar ID for the resource this Native User is for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"native_user_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The username of the Native User.",
				Deprecated:  "This field will be removed in a future version. Please use the 'type' field along with other associated fields based on the type.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"native_user_secret": {
				// This description is used by the documentation generator and the language server.
				Description: "The password of the Native User.",
				Deprecated:  "This field will be removed in a future version. Please use the 'type' field along with other associated fields based on the type.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"type": {
				Description: "The type of the Native User. (one of " + ValidUserTypes + ")",
				Type:        schema.TypeString,
				Required:    true,
				// The type of the native user can't be changed after creation in the current API implementation
				// Just recreate the resource if the type changed
				ForceNew: true,
			},
			// Password user fields
			"username": {
				Description: "For password users, the username.",
				Type:        schema.TypeString,
				Required:    false,
			},
			"username_is_env": {
				Description: "For password users, whether the username is the name of an environment variable where the real username is stored.",
				Type:        schema.TypeBool,
				Required:    false,
			},
			"password": {
				Description: "For password users, the password.",
				Type:        schema.TypeString,
				Required:    false,
				Sensitive:   true,
			},
			"password_is_env": {
				Description: "For password users, whether the password is the name of an environment variable where the real password is stored.",
				Type:        schema.TypeBool,
				Required:    false,
			},
			// IAM user fields
			"iam_type": {
				Description: "For IAM users, the type of IAM user. (one of " + ValidIAMTypes + ")",
				Type:        schema.TypeString,
				Required:    false,
			},
			"iam_role": {
				Description: "For IAM users, the IAM role to assume. Currently only takes effect for AWS IAM.",
				Type:        schema.TypeString,
				Required:    false,
			},
			// K8s user fields
			"kubeconfig_env": {
				Description: "For kubernetes users, the name of the environment variable where the path to a kubeconfig file is stored.",
				Type:        schema.TypeString,
				Required:    false,
			},
			"use_as_default": {
				// This description is used by the documentation generator and the language server.
				Description: "The password of the Native User.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Native User cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceNativeUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	ResourceId := d.Get("resource_id").(string)
	Type := d.Get("type").(string)
	UseAsDefault := d.Get("use_as_default").(bool)
	TerminationProtection := d.Get("termination_protection").(bool)

	req := &corev1.CreateNativeUserV2Request{
		ResourceId:            ResourceId,
		UseAsDefault:          UseAsDefault,
		TerminationProtection: TerminationProtection,
	}

	switch Type {
	case "password":
		req.Password = &corev1.CreateNativeUserV2Request_Password{
			Username:      d.Get("username").(string),
			UsernameIsEnv: d.Get("username_is_env").(bool),
			Password:      d.Get("password").(string),
			PasswordIsEnv: d.Get("password_is_env").(bool),
		}
	case "iam":
		req.Iam = &corev1.CreateNativeUserV2Request_Iam{
			IamType: d.Get("iam_type").(string),
			IamRole: d.Get("iam_role").(string),
		}
	case "k8s":
		req.K8S = &corev1.CreateNativeUserV2Request_K8S{
			KubeconfigEnv: d.Get("kubeconfig_env").(string),
		}
	default:
		return diag.Errorf("invalid native user type: %s (expected one of %s)", Type, ValidUserTypes)
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateNativeUserV2(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.NativeUser.Id)

	resourceNativeUserRead(ctx, d, meta)

	return diags
}

func resourceNativeUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Get("id").(string)

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetNativeUser(ctx, connect.NewRequest(&corev1.GetNativeUserRequest{Id: id}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Policy was deleted
			tflog.Warn(ctx, "The Native User "+id+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of UserOrgItem
	d.Set("resource_id", res.Msg.NativeUser.ResourceId)
	d.Set("use_as_default", res.Msg.NativeUser.UseAsDefault)
	d.Set("termination_protection", res.Msg.NativeUser.TerminationProtection)

	switch res.Msg.NativeUser.Type.(type) {
	case *corev1.NativeUser_Password:
		d.Set("username", res.Msg.NativeUser.Password.Username)
		d.Set("username_is_env", res.Msg.NativeUser.Password.UsernameIsEnv)
		d.Set("password", res.Msg.NativeUser.Password.Password)
		d.Set("password_is_env", res.Msg.NativeUser.Password.PasswordIsEnv)
		d.Set("type", "password")
	case *corev1.NativeUser_Iam:
		d.Set("iam_type", res.Msg.NativeUser.Iam.IamType)
		d.Set("iam_role", res.Msg.NativeUser.Iam.IamRole)
		d.Set("type", "iam")
	case *corev1.NativeUser_K8S:
		d.Set("kubeconfig_env", res.Msg.NativeUser.K8S.KubeconfigEnv)
		d.Set("type", "k8s")
	default:
		return diag.Errorf("invalid native user type: %s (expected one of %s)", reflect.TypeOf(res.Msg.NativeUser.Type), ValidUserTypes)
	}

	d.SetId(res.Msg.NativeUser.Id)

	return diags
}

func resourceNativeUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	// if d.HasChangesExcept("use_as_default", "native_User_secret") {
	// 	return diag.Errorf("Native Users can only be updated for use_as_default and native_User_secret. Please create a new Native User.")
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

	userType := d.Get("type").(string)
	switch userType {
	case "password":
		if d.HasChange("username") || d.HasChange("username_is_env") || d.HasChange("password") || d.HasChange("password_is_env") {
			_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateNativeUserV2(ctx, connect.NewRequest(&corev1.UpdateNativeUserV2Request{
				Id: id,
				Password: &corev1.UpdateNativeUserV2Request_Password{
					Username:      d.Get("username").(string),
					UsernameIsEnv: d.Get("username_is_env").(bool),
					Password:      d.Get("password").(string),
					PasswordIsEnv: d.Get("password_is_env").(bool),
				},
			}))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	case "iam":
		if d.HasChange("iam_type") || d.HasChange("iam_role") {
			_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateNativeUserV2(ctx, connect.NewRequest(&corev1.UpdateNativeUserV2Request{
				Id: id,
				Iam: &corev1.UpdateNativeUserV2Request_Iam{
					IamType: d.Get("iam_type").(string),
					IamRole: d.Get("iam_role").(string),
				},
			}))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	case "k8s":
		if d.HasChange("kubeconfig_env") {
			_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateNativeUserV2(ctx, connect.NewRequest(&corev1.UpdateNativeUserV2Request{
				Id: id,
				K8S: &corev1.UpdateNativeUserV2Request_K8S{
					KubeconfigEnv: d.Get("kubeconfig_env").(string),
				},
			}))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	default:
		return diag.Errorf("invalid native user type: %s (expected one of %s)", userType, ValidUserTypes)
	}

	// Deprecated but this still works
	if d.HasChange("native_user_secret") {
		secret := d.Get("native_user_secret").(string)
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

	resourceNativeUserRead(ctx, d, meta)

	return diags
}

func resourceNativeUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Native User cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteNativeUser(ctx, connect.NewRequest(&corev1.DeleteNativeUserRequest{Id: d.Id()}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
