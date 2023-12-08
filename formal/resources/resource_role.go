package resource

import (
	"github.com/formalco/terraform-provider-formal/formal/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRole() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "Use formal_user resource instead.",
		// This description is used by the documentation generator and the language server.
		Description: "User in Formal.",

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "User ID",
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
				Description:  "Either 'human' or 'machine'.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.UserType(),
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
				Description:  "If the role is of type `machine`, this is an optional designation for the app that this role will be used for. Supported values are `metabase`, `tableau`, and `popsql`.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.UserAppType(),
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
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this User cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}
