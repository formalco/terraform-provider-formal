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
		Description: "Data source for looking up a Connector by name or by ID. Use either `name` or `id`, but not both.",
		ReadContext: connectorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description:  "The ID of the Connector to look up.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Description:  "The name of the Connector to look up. Use this to fetch a connector by name.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"space_id": {
				Description: "The ID of the Space the Connector is in.",
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
		},
	}
}

func connectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	var connector *corev1.Connector

	if connectorID, ok := d.GetOk("id"); ok {
		// Fetch by ID using GetConnector (https://docs.joinformal.com/api-reference/corev1connectorservice/getconnector)
		res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnector(ctx, connect.NewRequest(&corev1.GetConnectorRequest{Id: connectorID.(string)}))
		if err != nil {
			if connect.CodeOf(err) == connect.CodeNotFound {
				return diag.Errorf("no connector found with id %s", connectorID)
			}
			return diag.FromErr(err)
		}
		connector = res.Msg.Connector
	} else {
		// Fetch by name using ListConnectors with filter
		name := d.Get("name").(string)
		filterValue, err := anypb.New(&wrapperspb.StringValue{
			Value: name,
		})
		if err != nil {
			return diag.FromErr(err)
		}
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
		connector = res.Msg.Connectors[0]
	}

	// Get API key separately (GetConnectorApiKey)
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
