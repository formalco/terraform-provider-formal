package datasources

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func User() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for looking up a User by ID or by identity. Use either `id` or `db_username`, but not both. A human user's identity is `idp:formal:human:<email>`, so this can resolve an email to a user ID, for example to add a user to a Group with `formal_group_link_user`.",
		ReadContext: userRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The ID of the User to look up.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "db_username"},
			},
			"db_username": {
				Description:  "The identity of the User to look up, for example `idp:formal:human:jane@example.com`.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "db_username"},
			},
			"type": {
				Description: "The type of this User, either `human` or `machine`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"first_name": {
				Description: "The first name of this User. Only set for human users.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_name": {
				Description: "The last name of this User. Only set for human users.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "The email of this User. Only set for human users.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"full_name": {
				Description: "The full name of this User.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"group_ids": {
				Description: "The IDs of the Groups this User belongs to.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"termination_protection": {
				Description: "If set to true, this User cannot be deleted.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func userRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	var user *corev1.User

	if userID, ok := d.GetOk("id"); ok {
		res, err := c.Grpc.Sdk.UserServiceClient.GetUser(ctx, connect.NewRequest(&corev1.GetUserRequest{Id: userID.(string)}))
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				return diag.Errorf("no user found with id %s", userID)
			}
			return diag.FromErr(err)
		}
		user = res.Msg.User
	} else {
		dbUsername := d.Get("db_username").(string)
		filterValue, err := anypb.New(&wrapperspb.StringValue{Value: dbUsername})
		if err != nil {
			return diag.FromErr(err)
		}
		res, err := c.Grpc.Sdk.UserServiceClient.ListUsers(ctx, connect.NewRequest(&corev1.ListUsersRequest{
			Filter: &corev1.Filter{
				Field: &corev1.Field{
					Key:      "db_username",
					Operator: "equals",
					Value:    filterValue,
				},
			},
			Limit: 1,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Msg.Users) == 0 {
			return diag.Errorf("no user found with db_username %s", dbUsername)
		}
		user = res.Msg.Users[0]
	}

	d.SetId(user.Id)
	d.Set("db_username", user.DbUsername)
	d.Set("full_name", user.FullName)
	d.Set("type", user.Type)
	d.Set("group_ids", user.GroupIds)
	d.Set("termination_protection", user.TerminationProtection)
	if human := user.GetHuman(); human != nil {
		d.Set("first_name", human.FirstName)
		d.Set("last_name", human.LastName)
		d.Set("email", human.Email)
	}

	return diags
}
