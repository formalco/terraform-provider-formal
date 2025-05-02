package resource

import (
	"context"
	"fmt"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourcePermission() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Permission in Formal.",

		CreateContext: resourcePermissionCreate,
		ReadContext:   resourcePermissionRead,
		UpdateContext: resourcePermissionUpdate,
		DeleteContext: resourcePermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourcePermissionInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourcePermissionStateUpgradeV0,
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Permission Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				// This description is used by the documentation generator and the language server.
				Description: "Permission Description.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"module": {
				// This description is used by the documentation generator and the language server.
				Description: "The module describing how the permission works. Create one in the Formal Console.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of this Permission.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When the permission was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				// This description is used by the documentation generator and the language server.
				Description: "Last update time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				// This description is used by the documentation generator and the language server.
				Description: "Defines the current status of the permission. It can be one of the following: 'draft', 'dry-run', or 'active'.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"draft",
					"dry-run",
					"active",
				}, false),
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Permission cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourcePermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	Name := d.Get("name").(string)
	Description := d.Get("description").(string)
	Module := d.Get("module").(string)
	Status := d.Get("status").(string)
	TerminationProtection := d.Get("termination_protection").(bool)

	newPermission := &corev1.CreatePermissionRequest{
		Name:                  Name,
		Description:           Description,
		Code:                  Module,
		Status:                Status,
		TerminationProtection: TerminationProtection,
	}

	res, err := c.Grpc.Sdk.PermissionsServiceClient.CreatePermission(ctx, connect.NewRequest(newPermission))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Permission.Id)

	resourcePermissionRead(ctx, d, meta)
	return diags
}

func resourcePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	permissionId := d.Id()

	res, err := c.Grpc.Sdk.PermissionsServiceClient.GetPermission(ctx, connect.NewRequest(&corev1.GetPermissionRequest{Id: permissionId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Permission was deleted
			tflog.Warn(ctx, "The Permission with ID "+permissionId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of Permission
	d.Set("id", res.Msg.Permission.Id)
	d.Set("name", res.Msg.Permission.Name)
	d.Set("description", res.Msg.Permission.Description)
	d.Set("module", res.Msg.Permission.Code)
	d.Set("status", res.Msg.Permission.Status)
	d.Set("termination_protection", res.Msg.Permission.TerminationProtection)

	d.SetId(permissionId)

	return diags
}

func resourcePermissionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	permissionId := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("module") || d.HasChange("status") || d.HasChange("termination_protection") {
		Name := d.Get("name").(string)
		Description := d.Get("description").(string)
		Module := d.Get("module").(string)
		Status := d.Get("status").(string)
		TerminationProtection := d.Get("termination_protection").(bool)

		updatedPermission := &corev1.UpdatePermissionRequest{
			Id:                    permissionId,
			Name:                  Name,
			Description:           Description,
			Code:                  Module,
			Status:                Status,
			TerminationProtection: TerminationProtection,
		}

		_, err := c.Grpc.Sdk.PermissionsServiceClient.UpdatePermission(ctx, connect.NewRequest(updatedPermission))
		if err != nil {
			return diag.FromErr(err)
		}

		return resourcePermissionRead(ctx, d, meta)
	}
	return diag.Errorf("At the moment you can only update a permission's name, description, module and status. Please delete and recreate the Permission")
}

func resourcePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	permissionId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("Permission cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.PermissionsServiceClient.DeletePermission(ctx, connect.NewRequest(&corev1.DeletePermissionRequest{Id: permissionId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourcePermissionInstanceResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"status": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourcePermissionStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		return nil, fmt.Errorf("sidecar resource state upgrade failed, state is nil")
	}

	c := meta.(*clients.Clients)

	if val, ok := rawState["id"]; ok {
		res, err := c.Grpc.Sdk.PermissionsServiceClient.GetPermission(ctx, connect.NewRequest(&corev1.GetPermissionRequest{Id: val.(string)}))
		if err != nil {
			return nil, err
		}
		rawState["status"] = res.Msg.Permission.Status
	}

	return rawState, nil
}
