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

func Connector() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for looking up a Connector by name.",
		ReadContext: connectorRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the Connector.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "The Formal ID for this Connector.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"api_key": {
				Description: "Api key for the deployed Connector.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"termination_protection": {
				Description: "If set to true, this Connector cannot be deleted.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"space_id": {
				Description: "The ID of the Space the Connector is in.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func connectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// List connectors with name filter
	res, err := c.Grpc.Sdk.ConnectorServiceClient.ListConnectors(ctx, connect.NewRequest(&corev1.ListConnectorsRequest{
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
	if len(res.Msg.Connectors) == 0 {
		return diag.Errorf("no connector found with name %s", name)
	}

	connector := res.Msg.Connectors[0]

	// Get API key separately
	resApiKey, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorApiKey(ctx, connect.NewRequest(&corev1.GetConnectorApiKeyRequest{Id: connector.Id}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(connector.Id)
	d.Set("name", connector.Name)
	d.Set("api_key", resApiKey.Msg.Secret)
	d.Set("termination_protection", connector.TerminationProtection)
	if connector.Space != nil {
		d.Set("space_id", connector.Space.Id)
	}

	return diags
}
