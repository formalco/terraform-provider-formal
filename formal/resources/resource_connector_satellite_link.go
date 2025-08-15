package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceConnectorSatelliteLink() *schema.Resource {
	return &schema.Resource{
		Description:   "Link a Connector to a Satellite.",
		CreateContext: resourceConnectorSatelliteLinkCreate,
		ReadContext:   resourceConnectorSatelliteLinkRead,
		DeleteContext: resourceConnectorSatelliteLinkDelete,
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this connector satellite link.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector_id": {
				Description: "The ID of the Connector to link to the satellite.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"satellite_id": {
				Description: "The ID of the Satellite to link to the connector.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"satellite_type": {
				Description: "The type of satellite being linked.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"data_classifier",
					"policy_data_loader",
				}, false),
			},
			"created_at": {
				Description: "The timestamp when the connector satellite link was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "The timestamp when the connector satellite link was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceConnectorSatelliteLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	req := &corev1.CreateConnectorSatelliteLinkRequest{
		ConnectorId:   d.Get("connector_id").(string),
		SatelliteId:   d.Get("satellite_id").(string),
		SatelliteType: d.Get("satellite_type").(string),
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorSatelliteLink(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ConnectorSatelliteLink.Id)

	resourceConnectorSatelliteLinkRead(ctx, d, meta)

	return diags
}

func resourceConnectorSatelliteLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorSatelliteLinkId := d.Id()

	req := connect.NewRequest(&corev1.GetConnectorSatelliteLinkRequest{
		Id: connectorSatelliteLinkId,
	})

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorSatelliteLink(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector satellite link was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.ConnectorSatelliteLink.Id)
	d.Set("connector_id", res.Msg.ConnectorSatelliteLink.ConnectorId)
	d.Set("satellite_id", res.Msg.ConnectorSatelliteLink.SatelliteId)
	d.Set("satellite_type", res.Msg.ConnectorSatelliteLink.SatelliteType)
	d.Set("created_at", res.Msg.ConnectorSatelliteLink.CreatedAt.AsTime().Format(time.RFC3339))
	d.Set("updated_at", res.Msg.ConnectorSatelliteLink.UpdatedAt.AsTime().Format(time.RFC3339))

	d.SetId(res.Msg.ConnectorSatelliteLink.Id)

	return diags
}


func resourceConnectorSatelliteLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorSatelliteLinkId := d.Id()

	req := connect.NewRequest(&corev1.DeleteConnectorSatelliteLinkRequest{
		Id: connectorSatelliteLinkId,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorSatelliteLink(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
