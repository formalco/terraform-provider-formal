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

func Group() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for looking up a Group by ID or by name. Use either `id` or `name`, but not both.",
		ReadContext: groupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The ID of this Group.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Description:  "The name of the Group to look up. Use this to fetch a group by name.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"description": {
				Description: "Description for this Group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"termination_protection": {
				Description: "If set to true, this Group cannot be deleted.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func groupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	var group *corev1.Group

	if groupID, ok := d.GetOk("id"); ok {
		res, err := c.Grpc.Sdk.GroupServiceClient.GetGroup(ctx, connect.NewRequest(&corev1.GetGroupRequest{Id: groupID.(string)}))
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				return diag.Errorf("no group found with id %s", groupID)
			}
			return diag.FromErr(err)
		}
		group = res.Msg.Group
	} else {
		name := d.Get("name").(string)
		filterValue, err := anypb.New(&wrapperspb.StringValue{
			Value: name,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		res, err := c.Grpc.Sdk.GroupServiceClient.ListGroups(ctx, connect.NewRequest(&corev1.ListGroupsRequest{
			Filter: &corev1.Filter{
				Field: &corev1.Field{
					Key:      "name",
					Operator: "equals",
					Value:    filterValue,
				},
			},
			Limit: 1,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Msg.Groups) == 0 {
			return diag.Errorf("no group found with name %s", name)
		}
		group = res.Msg.Groups[0]
	}

	d.SetId(group.Id)
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("termination_protection", group.TerminationProtection)

	return diags
}
