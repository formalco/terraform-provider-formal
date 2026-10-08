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

const groupsPageSize int32 = 100

type groupLister interface {
	ListGroups(context.Context, *corev1.ListGroupsRequest) (*corev1.ListGroupsResponse, error)
}

func Groups() *schema.Resource {
	groupSchema := commonGroupSchema()
	groupSchema["id"] = &schema.Schema{
		Description: "The ID of the Group.",
		Type:        schema.TypeString,
		Computed:    true,
	}
	groupSchema["name"] = &schema.Schema{
		Description: "The name of the Group.",
		Type:        schema.TypeString,
		Computed:    true,
	}

	return &schema.Resource{
		Description: "Data source for listing all Groups in Formal.",
		ReadContext: groupsRead,
		Schema: map[string]*schema.Schema{
			"groups": {
				Description: "All Groups in Formal.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: groupSchema,
				},
			},
		},
	}
}

func groupsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	groups, err := listAllGroups(ctx, c.Grpc.Sdk.GroupServiceClient)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("groups", flattenGroups(groups)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("groups")

	return nil
}

func listAllGroups(ctx context.Context, client groupLister) ([]*corev1.Group, error) {
	var groups []*corev1.Group
	var cursor string

	for {
		res, err := client.ListGroups(ctx, &corev1.ListGroupsRequest{
			Limit:  groupsPageSize,
			Cursor: cursor,
		})
		if err != nil {
			return nil, err
		}
		groups = append(groups, res.Groups...)

		if res.ListMetadata == nil || res.ListMetadata.NextCursor == "" {
			return groups, nil
		}
		if res.ListMetadata.NextCursor == cursor {
			return nil, fmt.Errorf("list groups returned the same cursor %q twice", cursor)
		}
		cursor = res.ListMetadata.NextCursor
	}
}

func flattenGroups(groups []*corev1.Group) []map[string]any {
	return lo.Map(groups, func(group *corev1.Group, _ int) map[string]any {
		return map[string]any{
			"id":                     group.Id,
			"name":                   group.Name,
			"description":            group.Description,
			"termination_protection": group.TerminationProtection,
		}
	})
}
