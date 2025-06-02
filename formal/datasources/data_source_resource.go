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
		Description: "Data source for looking up a Resource by name.",
		ReadContext: resourceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Resource.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"hostname": {
				Description: "Hostname of the Resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"technology": {
				Description: "Technology of the Resource.",
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

	// List resources with name filter
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

	d.SetId(res.Msg.Resources[0].Id)
	d.Set("name", res.Msg.Resources[0].Name)
	d.Set("hostname", res.Msg.Resources[0].Hostname)
	d.Set("port", res.Msg.Resources[0].Port)
	d.Set("technology", res.Msg.Resources[0].Technology)
	d.Set("environment", res.Msg.Resources[0].Environment)
	d.Set("termination_protection", res.Msg.Resources[0].TerminationProtection)
	if res.Msg.Resources[0].Space != nil {
		d.Set("space_id", res.Msg.Resources[0].Space.Id)
	}
	d.Set("created_at", res.Msg.Resources[0].CreatedAt.AsTime().Unix())

	return diags
}
