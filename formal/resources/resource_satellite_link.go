package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceSatelliteLink() *schema.Resource {
	return &schema.Resource{
		Description:   "Link a Satellite to another Satellite. For example, link a Data Discovery Satellite to an AI Satellite for column classification.",
		CreateContext: resourceSatelliteLinkCreate,
		ReadContext:   resourceSatelliteLinkRead,
		DeleteContext: resourceSatelliteLinkDelete,
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this satellite link.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"source_satellite_id": {
				Description: "The ID of the source Satellite (e.g., Data Discovery Satellite).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"target_satellite_id": {
				Description: "The ID of the target Satellite (e.g., AI Satellite).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"created_at": {
				Description: "The timestamp when the satellite link was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "The timestamp when the satellite link was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSatelliteLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	req := &corev1.CreateSatelliteLinkRequest{
		SourceSatelliteId: d.Get("source_satellite_id").(string),
		TargetSatelliteId: d.Get("target_satellite_id").(string),
	}

	res, err := c.Grpc.Sdk.SatelliteServiceClient.CreateSatelliteLink(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.SatelliteLink.Id)

	resourceSatelliteLinkRead(ctx, d, meta)

	return diags
}

func resourceSatelliteLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	satelliteLinkId := d.Id()

	req := connect.NewRequest(&corev1.GetSatelliteLinkRequest{
		Id: satelliteLinkId,
	})

	res, err := c.Grpc.Sdk.SatelliteServiceClient.GetSatelliteLink(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Satellite link was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.SatelliteLink.Id)
	d.Set("source_satellite_id", res.Msg.SatelliteLink.SourceSatelliteId)
	d.Set("target_satellite_id", res.Msg.SatelliteLink.TargetSatelliteId)
	d.Set("created_at", res.Msg.SatelliteLink.CreatedAt.AsTime().Format(time.RFC3339))
	d.Set("updated_at", res.Msg.SatelliteLink.UpdatedAt.AsTime().Format(time.RFC3339))

	d.SetId(res.Msg.SatelliteLink.Id)

	return diags
}

func resourceSatelliteLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	satelliteLinkId := d.Id()

	req := connect.NewRequest(&corev1.DeleteSatelliteLinkRequest{
		Id: satelliteLinkId,
	})

	_, err := c.Grpc.Sdk.SatelliteServiceClient.DeleteSatelliteLink(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
