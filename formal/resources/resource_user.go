package resource

import (
	"context"
	"errors"
	"strconv"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/formalco/terraform-provider-formal/formal/clients"
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
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "For machine users, the name of the user.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"app_type": {
				// This description is used by the documentation generator and the language server.
				Description: "If the user is of type `machine`, this is an optional designation for the app that this user will be used for. Supported values are `metabase`, `tableau`, and `popsql`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"machine_user_access_token": {
				// This description is used by the documentation generator and the language server.
				Description: "If the user is of type `machine`, this is the access token (database password) of this user.",
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

	userType := d.Get("type").(string)
	var res *connect.Response[corev1.CreateUserResponse]
	var err error

	switch userType {
	case "human":
		user := &corev1.CreateUserRequest_Human{
			Human: &corev1.User_Human{
				FirstName: d.Get("first_name").(string),
				LastName:  d.Get("last_name").(string),
				Email:     d.Get("email").(string),
			},
		}
		res, err = c.Grpc.Sdk.UserServiceClient.CreateUser(ctx, connect.NewRequest(&corev1.CreateUserRequest{
			Type:                  d.Get("type").(string),
			Info:                  user,
			ExpireAt:              timestamppb.New(time.Unix(int64(d.Get("expire_at").(int)), 0)),
			TerminationProtection: d.Get("termination_protection").(bool),
		}))
		if err != nil {
			return diag.FromErr(err)
		}

	case "machine":
		user := &corev1.CreateUserRequest_Machine{
			Machine: &corev1.User_Machine{
				Name: d.Get("name").(string),
			},
		}
		res, err = c.Grpc.Sdk.UserServiceClient.CreateUser(ctx, connect.NewRequest(&corev1.CreateUserRequest{
			Type:                  d.Get("type").(string),
			Info:                  user,
			ExpireAt:              timestamppb.New(time.Unix(int64(d.Get("expire_at").(int)), 0)),
			TerminationProtection: d.Get("termination_protection").(bool),
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("Unsupported user type: %s", userType)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	for {
		// Retrieve status
		createdUser, err := c.Grpc.Sdk.UserServiceClient.GetUser(ctx, connect.NewRequest(&corev1.GetUserRequest{Id: res.Msg.User.Id}))
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

	d.SetId(res.Msg.User.Id)

	resourceUserRead(ctx, d, meta)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	userId := d.Id()

	res, err := c.Grpc.Sdk.UserServiceClient.GetUser(ctx, connect.NewRequest(&corev1.GetUserRequest{Id: userId}))
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
	d.Set("expire_at", res.Msg.User.ExpireAt.AsTime().Unix())
	d.Set("termination_protection", res.Msg.User.TerminationProtection)

	if res.Msg.User.Type == "machine" {
		res, err := c.Grpc.Sdk.UserServiceClient.GetMachineUserCredentials(ctx, connect.NewRequest(&corev1.GetMachineUserCredentialsRequest{Id: userId}))
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("machine_user_access_token", res.Msg.Password)
	}

	switch info := res.Msg.User.Info.(type) {
	case *corev1.User_Human_:
		d.Set("first_name", info.Human.FirstName)
		d.Set("last_name", info.Human.LastName)
		d.Set("email", info.Human.Email)
	case *corev1.User_Machine_:
		d.Set("name", info.Machine.Name)
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

	_, err := c.Grpc.Sdk.UserServiceClient.UpdateUser(ctx, connect.NewRequest(&corev1.UpdateUserRequest{
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

	_, err := c.Grpc.Sdk.UserServiceClient.DeleteUser(ctx, connect.NewRequest(&corev1.DeleteUserRequest{Id: userId}))
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	deleteTimeStart := time.Now()
	for {
		// Retrieve status
		_, err = c.Grpc.Sdk.UserServiceClient.GetUser(ctx, connect.NewRequest(&corev1.GetUserRequest{Id: userId}))
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
