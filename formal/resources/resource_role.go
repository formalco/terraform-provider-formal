package resource

import (
	"context"

	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRole() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a Role in formal.",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,

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
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "For machine users, the name of the role.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"app_type": {
				// This description is used by the documentation generator and the language server.
				Description: "If machine, app that this role will be used for. Supported values are `metabase`, `tableau`, and `popsql`.",
				Type:        schema.TypeString,
				Optional:    true,
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
		//
		Name:    d.Get("name").(string),
		AppType: d.Get("app_type").(string),
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
		return diag.FromErr(err)
	}
	if role == nil {
		return diags
	}


	// Should map to all fields of RoleOrgItem
	d.Set("id", role.ID)
	d.Set("type", role.Type)
	d.Set("db_username", role.DBUsername)
	d.Set("first_name", role.FirstName)
	d.Set("last_name", role.LastName)
	d.Set("email", role.Email)
	d.Set("name", role.Name)
	d.Set("app_type", role.AppType)

	d.SetId(roleId)

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Roles are immutable at the moment, but will be updateable soon. Please create a new role. Thank you!")

	// client := meta.(*Client)

	// roleId := d.Id()

	// roleUpdate := RoleOrgItem{
	// 	Name:        d.Get("name").(string),
	// 	Description: d.Get("description").(string),
	// 	Module:      d.Get("module").(string),
	// }

	// client.UpdateRole(roleId, roleUpdate)
	// return resourceRoleRead(ctx, d, meta)
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
