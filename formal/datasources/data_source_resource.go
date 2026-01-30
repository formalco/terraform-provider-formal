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

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for looking up a Resource by ID or by name. Use either `id` or `name`, but not both.",
		ReadContext: resourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The ID of this Resource.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Description:  "The name of the Resource to look up. Use this to fetch a resource by name.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"technology": {
				Description: "Technology of the Resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hostname": {
				Description: "Hostname of the Resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"port": {
				Description: "The port your Resource is listening on.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"created_at": {
				Description: "Creation time of the Resource.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"environment": {
				Description: "Environment for the Resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"termination_protection": {
				Description: "If set to true, the Resource cannot be deleted.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"space_id": {
				Description: "The ID of the Space the Resource is in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	var resource *corev1.Resource

	if resourceID, ok := d.GetOk("id"); ok {
		res, err := c.Grpc.Sdk.ResourceServiceClient.GetResource(ctx, connect.NewRequest(&corev1.GetResourceRequest{Id: resourceID.(string)}))
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				return diag.Errorf("no resource found with id %s", resourceID)
			}
			return diag.FromErr(err)
		}
		resource = res.Msg.Resource
	} else {
		name, ok := d.GetOk("name")
		if !ok {
			return diag.Errorf("name or id is required")
		}
		filterValue, err := anypb.New(&wrapperspb.StringValue{
			Value: name.(string),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		res, err := c.Grpc.Sdk.ResourceServiceClient.ListResources(ctx, connect.NewRequest(&corev1.ListResourcesRequest{
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
		if len(res.Msg.Resources) == 0 {
			return diag.Errorf("no resource found with name %s", name)
		}
		resource = res.Msg.Resources[0]
	}

	d.SetId(resource.Id)
	d.Set("name", resource.Name)
	d.Set("hostname", resource.Hostname)
	d.Set("port", resource.Port)
	d.Set("technology", resource.Technology)
	d.Set("environment", resource.Environment)
	d.Set("termination_protection", resource.TerminationProtection)
	if resource.Space != nil {
		d.Set("space_id", resource.Space.Id)
	}
	if resource.CreatedAt != nil {
		d.Set("created_at", resource.CreatedAt.AsTime().Unix())
	}

	return diags
}
