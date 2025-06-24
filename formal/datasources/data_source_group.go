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
		Description: "Data source for looking up a Group by name.",
		ReadContext: groupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "The Formal ID for this Group.",
				Type:        schema.TypeString,
				Computed:    true,
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

	name, ok := d.GetOk("name")
	if !ok {
		return diag.Errorf("name must be specified")
	}

	filterValue, err := anypb.New(&wrapperspb.StringValue{
		Value: name.(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// List groups with name filter
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

	d.SetId(res.Msg.Groups[0].Id)
	d.Set("name", res.Msg.Groups[0].Name)
	d.Set("description", res.Msg.Groups[0].Description)
	d.Set("termination_protection", res.Msg.Groups[0].TerminationProtection)

	return diags
}
