package resource

import (
	"context"
	"errors"
	"strconv"
	"time"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/formalco/terraform-provider-formal/formal/validation"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
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
				Description: "For machine users, the name of the user.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"app_type": {
				// This description is used by the documentation generator and the language server.
				Description:  "If the user is of type `machine`, this is an optional designation for the app that this user will be used for. Supported values are `metabase`, `tableau`, and `popsql`.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.UserAppType(),
			},
			"machine_user_access_token": {
				// This description is used by the documentation generator and the language server.
				Description: "If the user is of type `machine`, this is the accesss token (database password) of this user.",
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

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.UserServiceClient.CreateUserV2(ctx, connect.NewRequest(&adminv1.CreateUserV2Request{
		FirstName:             d.Get("first_name").(string),
		LastName:              d.Get("last_name").(string),
		Type:                  d.Get("type").(string),
		AppType:               d.Get("app_type").(string),
		Name:                  d.Get("name").(string),
		Email:                 d.Get("email").(string),
		Admin:                 d.Get("admin").(bool),
		ExpireAt:              timestamppb.New(time.Unix(int64(d.Get("expire_at").(int)), 0)),
		TerminationProtection: d.Get("termination_protection").(bool),
	}))

	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	for {
		// Retrieve status
		createdUser, err := c.Grpc.Sdk.UserServiceClient.GetUserById(ctx, connect.NewRequest(&adminv1.GetUserByIdRequest{Id: res.Msg.Id}))
		if err != nil {
			if currentErrors >= ErrorTolerance {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors+1)+" retrieving User: ", map[string]interface{}{"err": err})
				currentErrors += 1
				time.Sleep(15 * time.Second)
				continue
			}
		}

		if createdUser == nil {
			err = errors.New("user with the given ID not found. It may have been deleted")
			return diag.FromErr(err)
		}

		// Check status
		if createdUser.Msg.User != nil {
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}

	d.SetId(res.Msg.Id)

	resourceUserRead(ctx, d, meta)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	userId := d.Id()

	res, err := c.Grpc.Sdk.UserServiceClient.GetUserById(ctx, connect.NewRequest(&adminv1.GetUserByIdRequest{Id: userId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Policy was deleted
			tflog.Warn(ctx, "The Role with ID "+userId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
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
	d.Set("termination_protection", res.Msg.User.TerminationProtection)
	if c.Grpc.ReturnSensitiveValue {
		d.Set("machine_user_access_token", res.Msg.User.MachineRoleAccessToken)
	}

	d.SetId(userId)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	userId := d.Id()
	name := d.Get("name").(string)
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	_, err := c.Grpc.Sdk.UserServiceClient.UpdateUser(ctx, connect.NewRequest(&adminv1.UpdateUserRequest{
		Id:                    userId,
		Name:                  name,
		FirstName:             firstName,
		LastName:              lastName,
		TerminationProtection: &terminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceUserRead(ctx, d, meta)

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	userId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)

	if terminationProtection {
		return diag.Errorf("User cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.UserServiceClient.DeleteUser(ctx, connect.NewRequest(&adminv1.DeleteUserRequest{Id: userId}))
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err = c.Grpc.Sdk.UserServiceClient.GetUserById(ctx, connect.NewRequest(&adminv1.GetUserByIdRequest{Id: userId}))
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				tflog.Info(ctx, "User deleted", map[string]interface{}{"user_id": userId})
				// User was deleted
				break
			}

			// Handle other errors
			if currentErrors >= ErrorTolerance {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors)+" checking on User Status: ", map[string]interface{}{"err": err})
				currentErrors += 1
			}
		}

		if time.Since(deleteTimeStart) > time.Minute*15 {
			newErr := errors.New("deletion of this user has taken more than 15m")
			return diag.FromErr(newErr)
		}

		time.Sleep(15 * time.Second)
	}

	d.SetId("")

	return diags
}
