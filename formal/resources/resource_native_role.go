package resource

import (
	"context"
	"errors"
	"fmt"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"strings"

	"github.com/formalco/terraform-provider-formal/formal/api"
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
		},
	}
}

const nativeRoleDelimiter = "#_#"

func resourceNativeRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newRole := api.NativeRole{
		DatastoreId:      d.Get("datastore_id").(string),
		NativeRoleId:     d.Get("native_role_id").(string),
		NativeRoleSecret: d.Get("native_role_secret").(string),
		UseAsDefault:     d.Get("use_as_default").(bool),
	}

	role, err := c.Http.CreateNativeRole(newRole)
	if err != nil {
		return diag.FromErr(err)
	}
	if role == nil {
		return diag.FromErr(err)
	}

	d.SetId(newRole.DatastoreId + nativeRoleDelimiter + newRole.NativeRoleId)

	resourceNativeRoleRead(ctx, d, meta)

	return diags
}

func resourceNativeRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	roleId := d.Id()

	splitId := strings.Split(roleId, nativeRoleDelimiter)
	if len(splitId) != 2 {
		return diag.FromErr(errors.New("Resource ID for Native Role is Malformatted: " + roleId))
	}

	role, err := c.Http.GetNativeRole(splitId[0], splitId[1])
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Policy was deleted
			tflog.Warn(ctx, "The Native Role for Datastore ID "+splitId[0]+" and Native Role ID"+splitId[1]+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	fmt.Println("role")
	fmt.Println(role)
	if role == nil {
		return diags
	}

	// Should map to all fields of RoleOrgItem
	d.Set("datastore_id", role.DatastoreId)
	d.Set("native_role_id", role.NativeRoleId)
	d.Set("native_role_secret", role.NativeRoleSecret)
	d.Set("use_as_default", role.UseAsDefault)
	d.SetId(roleId)

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
			err := c.Http.UpdateNativeRole(datastoreId, nativeRoleId, "", true)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if d.HasChange("native_role_secret") {
		nativeRoleSecret := d.Get("native_role_secret").(string)
		err := c.Http.UpdateNativeRole(datastoreId, nativeRoleId, nativeRoleSecret, false)
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

	roleId := d.Id()
	splitId := strings.Split(roleId, nativeRoleDelimiter)
	if len(splitId) != 2 {
		return diag.FromErr(errors.New("Resource ID for Native Role is Malformatted: " + roleId))
	}

	err := c.Http.DeleteNativeRole(splitId[0], splitId[1])
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
