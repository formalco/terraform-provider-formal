package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

const usersPageSize int32 = 100

type userLister interface {
	ListUsers(context.Context, *corev1.ListUsersRequest) (*corev1.ListUsersResponse, error)
}

func Users() *schema.Resource {
	userSchema := commonUserSchema()
	userSchema["id"] = &schema.Schema{
		Description: "The ID of the User.",
		Type:        schema.TypeString,
		Computed:    true,
	}
	userSchema["db_username"] = &schema.Schema{
		Description: "The identity the User uses to access Formal.",
		Type:        schema.TypeString,
		Computed:    true,
	}

	return &schema.Resource{
		Description: "Data source for listing all Users in Formal.",
		ReadContext: usersRead,
		Schema: map[string]*schema.Schema{
			"users": {
				Description: "All Users in Formal.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: userSchema,
				},
			},
		},
	}
}

func usersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	users, err := listAllUsers(ctx, c.Grpc.Sdk.UserServiceClient)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("users", flattenUsers(users)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("users")

	return nil
}

func listAllUsers(ctx context.Context, client userLister) ([]*corev1.User, error) {
	var users []*corev1.User
	var cursor string

	for {
		res, err := client.ListUsers(ctx, &corev1.ListUsersRequest{
			Limit:  usersPageSize,
			Cursor: cursor,
		})
		if err != nil {
			return nil, err
		}
		users = append(users, res.Users...)

		if res.ListMetadata == nil || res.ListMetadata.NextCursor == "" {
			return users, nil
		}
		if res.ListMetadata.NextCursor == cursor {
			return nil, fmt.Errorf("list users returned the same cursor %q twice", cursor)
		}
		cursor = res.ListMetadata.NextCursor
	}
}

func flattenUsers(users []*corev1.User) []map[string]any {
	return lo.Map(users, func(user *corev1.User, _ int) map[string]any {
		flattened := map[string]any{
			"id":                     user.Id,
			"db_username":            user.DbUsername,
			"type":                   user.Type,
			"full_name":              user.FullName,
			"group_ids":              user.GroupIds,
			"termination_protection": user.TerminationProtection,
		}
		if human := user.GetHuman(); human != nil {
			flattened["first_name"] = human.FirstName
			flattened["last_name"] = human.LastName
			flattened["email"] = human.Email
		}
		return flattened
	})
}
