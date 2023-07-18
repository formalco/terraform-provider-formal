package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func ResourceRole() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "Use formal_user resource instead.",
		// This description is used by the documentation generator and the language server.
		Description: "User in Formal.",

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
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.UserServiceClient.CreateUser(ctx, connect.NewRequest(&adminv1.CreateUserRequest{
		FirstName: d.Get("first_name").(string),
		LastName:  d.Get("last_name").(string),
		Type:      d.Get("type").(string),
		AppType:   d.Get("app_type").(string),
		Name:      d.Get("name").(string),
		Email:     d.Get("email").(string),
		Admin:     d.Get("admin").(bool),
		ExpireAt:  timestamppb.New(time.Unix(int64(d.Get("expire_at").(int)), 0)),
	}))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.User.Id)

	resourceRoleRead(ctx, d, meta)

	return diags
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	roleId := d.Id()

	res, err := c.Grpc.Sdk.UserServiceClient.GetUserById(ctx, connect.NewRequest(&adminv1.GetUserByIdRequest{Id: roleId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Policy was deleted
			tflog.Warn(ctx, "The Role with ID "+roleId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of RoleOrgItem
	d.Set("id", res.Msg.User.Id)
	d.Set("type", res.Msg.User.Type)
	d.Set("db_username", res.Msg.User.DbUsername)
	d.Set("name", res.Msg.User.Name)
	d.Set("first_name", res.Msg.User.FirstName)
	d.Set("last_name", res.Msg.User.LastName)
	d.Set("email", res.Msg.User.Email)
	d.Set("admin", res.Msg.User.Admin)
	d.Set("app_type", res.Msg.User.AppType)
	d.Set("expire_at", res.Msg.User.ExpireAt.AsTime().Unix())

	if c.Grpc.ReturnSensitiveValue {
		d.Set("machine_role_access_token", res.Msg.User.MachineRoleAccessToken)
	}

	d.SetId(roleId)

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	roleId := d.Id()
	name := d.Get("name").(string)
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)

	_, err := c.Grpc.Sdk.UserServiceClient.UpdateUser(ctx, connect.NewRequest(&adminv1.UpdateUserRequest{
		Id:        roleId,
		Name:      name,
		FirstName: firstName,
		LastName:  lastName,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceRoleRead(ctx, d, meta)

	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	roleId := d.Id()

	_, err := c.Grpc.Sdk.UserServiceClient.DeleteUser(ctx, connect.NewRequest(&adminv1.DeleteUserRequest{Id: roleId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
