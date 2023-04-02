package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRole() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Role in Formal.",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "Role ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"db_username": {
				// This description is used by the documentation generator and the language server.
				Description: "The username that the user will use to access the sidecar.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"first_name": {
				// This description is used by the documentation generator and the language server.
				Description: "For human users, their first name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"last_name": {
				// This description is used by the documentation generator and the language server.
				Description: "For human users, their last name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Either 'human' or 'machine'.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"email": {
				// This description is used by the documentation generator and the language server.
				Description: "For human users, their email.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"admin": {
				Description: "For human users, specify if their admin.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "For machine users, the name of the role.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"app_type": {
				// This description is used by the documentation generator and the language server.
				Description: "If the role is of type `machine`, this is an optional designation for the app that this role will be used for. Supported values are `metabase`, `tableau`, and `popsql`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"machine_role_access_token": {
				// This description is used by the documentation generator and the language server.
				Description: "If the role is of type `machine`, this is the accesss token (database password) of this role.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"expire_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When the Role should be deleted and access revoked. Value should be provided in Unix epoch time, in seconds since midnight UTC of January 1, 1970.",
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newRole := api.Role{
		Type:      d.Get("type").(string),
		FirstName: d.Get("first_name").(string),
		LastName:  d.Get("last_name").(string),
		Email:     d.Get("email").(string),
		Admin:     d.Get("admin").(bool),
		Name:      d.Get("name").(string),
		AppType:   d.Get("app_type").(string),
		ExpireAt:  d.Get("expire_at").(int),
	}

	role, err := client.CreateRole(newRole)
	if err != nil {
		return diag.FromErr(err)
	}
	if role == nil {
		return diag.FromErr(err)
	}

	d.SetId(role.ID)

	resourceRoleRead(ctx, d, meta)

	return diags
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)
	var diags diag.Diagnostics

	roleId := d.Id()

	role, err := client.GetRole(roleId)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Policy was deleted
			tflog.Warn(ctx, "The Role with ID "+roleId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if role == nil {
		return diags
	}

	// Should map to all fields of RoleOrgItem
	d.Set("id", role.ID)
	d.Set("type", role.Type)
	d.Set("db_username", role.DBUsername)
	d.Set("name", role.Name)
	d.Set("first_name", role.FirstName)
	d.Set("last_name", role.LastName)
	d.Set("email", role.Email)
	d.Set("admin", role.Admin)
	d.Set("app_type", role.AppType)
	d.Set("machine_role_access_token", role.MachineRoleAccessToken)
	d.Set("expire_at", role.ExpireAt)
	d.SetId(roleId)

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	var diags diag.Diagnostics

	roleId := d.Id()
	name := d.Get("name").(string)
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)

	err := client.UpdateRole(roleId, api.Role{Name: name, FirstName: firstName, LastName: lastName})
	if err != nil {
		return diag.FromErr(err)
	}

	resourceRoleRead(ctx, d, meta)

	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)

	var diags diag.Diagnostics

	roleId := d.Id()

	err := client.DeleteRole(roleId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
