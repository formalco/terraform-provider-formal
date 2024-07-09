package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceSatellite() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Satellite",
		CreateContext: resourceSatelliteCreate,
		ReadContext:   resourceSatelliteRead,
		UpdateContext: resourceSatelliteUpdate,
		DeleteContext: resourceSatelliteDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Satellite.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for the Satellite.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"tls_cert": {
				// This description is used by the documentation generator and the language server.
				Description: "TLS certificate of the Satellite.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"api_key": {
				// This description is used by the documentation generator and the language server.
				Description: "Api key of the Satellite.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Satellite cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceSatelliteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	terminationProtection := d.Get("termination_protection").(bool)

	res, err := c.Grpc.Sdk.SatelliteServiceClient.CreateSatellite(ctx, connect.NewRequest(&corev1.CreateSatelliteRequest{
		Name:                  name,
		TerminationProtection: terminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Satellite.Id)

	resourceSatelliteRead(ctx, d, meta)

	return diags
}

func resourceSatelliteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.SatelliteServiceClient.GetSatellite(ctx, connect.NewRequest(&corev1.GetSatelliteRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", res.Msg.Satellite.Name)
	d.Set("termination_protection", res.Msg.Satellite.TerminationProtection)

	if c.Grpc.ReturnSensitiveValue {
		res, err := c.Grpc.Sdk.SatelliteServiceClient.GetSatelliteApiKey(ctx, connect.NewRequest(&corev1.GetSatelliteApiKeyRequest{Id: d.Id()}))
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("api_key", res.Msg.ApiKey)
		d.Set("tls_cert", res.Msg.ApiKey)
	}

	d.SetId(res.Msg.Satellite.Id)

	return diags
}

func resourceSatelliteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	satelliteId := d.Id()

	if d.HasChange("termination_protection") {
		terminationProtection := d.Get("termination_protection").(bool)
		_, err := c.Grpc.Sdk.SatelliteServiceClient.UpdateSatellite(ctx, connect.NewRequest(&corev1.UpdateSatelliteRequest{
			Id:                    satelliteId,
			TerminationProtection: &terminationProtection,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, err := c.Grpc.Sdk.SatelliteServiceClient.UpdateSatellite(ctx, connect.NewRequest(&corev1.UpdateSatelliteRequest{
			Id:   satelliteId,
			Name: &name,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceSatelliteRead(ctx, d, meta)

	return diags
}

func resourceSatelliteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Satellite cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.SatelliteServiceClient.DeleteSatellite(ctx, connect.NewRequest(&corev1.DeleteSatelliteRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
