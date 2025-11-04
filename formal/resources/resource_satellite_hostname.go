package resource

import (
	"context"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceSatelliteHostname() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage a Satellite Hostname.",
		CreateContext: resourceSatelliteHostnameCreate,
		ReadContext:   resourceSatelliteHostnameRead,
		UpdateContext: resourceSatelliteHostnameUpdate,
		DeleteContext: resourceSatelliteHostnameDelete,
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this satellite hostname.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"satellite_id": {
				Description: "The ID of the Satellite to create the hostname for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"hostname": {
				Description: "The hostname for the satellite.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"termination_protection": {
				Description: "If set to true, this satellite hostname cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"created_at": {
				Description: "The timestamp when the satellite hostname was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "The timestamp when the satellite hostname was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSatelliteHostnameCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	req := &corev1.CreateSatelliteHostnameRequest{
		SatelliteId:           d.Get("satellite_id").(string),
		Hostname:              d.Get("hostname").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.SatelliteServiceClient.CreateSatelliteHostname(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.SatelliteHostname.Id)

	resourceSatelliteHostnameRead(ctx, d, meta)

	return diags
}

func resourceSatelliteHostnameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	satelliteHostnameId := d.Id()

	req := connect.NewRequest(&corev1.GetSatelliteHostnameRequest{
		Id: satelliteHostnameId,
	})

	res, err := c.Grpc.Sdk.SatelliteServiceClient.GetSatelliteHostname(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Satellite hostname was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.SatelliteHostname.Id)
	d.Set("satellite_id", res.Msg.SatelliteHostname.SatelliteId)
	d.Set("hostname", res.Msg.SatelliteHostname.Hostname)
	d.Set("termination_protection", res.Msg.SatelliteHostname.TerminationProtection)
	d.Set("created_at", res.Msg.SatelliteHostname.CreatedAt.AsTime().Format(time.RFC3339))
	d.Set("updated_at", res.Msg.SatelliteHostname.UpdatedAt.AsTime().Format(time.RFC3339))

	d.SetId(res.Msg.SatelliteHostname.Id)

	return diags
}

func resourceSatelliteHostnameUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	fieldsThatCanChange := []string{"termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	terminationProtection := d.Get("termination_protection").(bool)
	req := connect.NewRequest(&corev1.UpdateSatelliteHostnameRequest{
		Id:                    d.Id(),
		TerminationProtection: &terminationProtection,
	})

	_, err := c.Grpc.Sdk.SatelliteServiceClient.UpdateSatelliteHostname(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceSatelliteHostnameRead(ctx, d, meta)

	return diags
}

func resourceSatelliteHostnameDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	satelliteHostnameId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Satellite hostname cannot be deleted because termination_protection is set to true")
	}

	req := connect.NewRequest(&corev1.DeleteSatelliteHostnameRequest{
		Id: satelliteHostnameId,
	})

	_, err := c.Grpc.Sdk.SatelliteServiceClient.DeleteSatelliteHostname(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
