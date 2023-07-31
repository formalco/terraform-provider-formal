package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func ResourceSatellite() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Satellite",
		CreateContext: resourceSatelliteCreate,
		ReadContext:   resourceSatelliteRead,
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
		},
	}
}

func resourceSatelliteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	name := d.Get("name").(string)

	res, err := c.Grpc.Sdk.SatelliteServiceClient.CreateSatellite(ctx, connect.NewRequest(&adminv1.CreateSatelliteRequest{
		Name: name,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Id)

	resourceIntegrationIncidentRead(ctx, d, meta)

	return diags
}

func resourceSatelliteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.SatelliteServiceClient.GetSatelliteById(ctx, connect.NewRequest(&adminv1.GetSatelliteByIdRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", res.Msg.Satellite.Name)

	if c.Grpc.ReturnSensitiveValue {
		d.Set("tls_cert", res.Msg.Satellite.TlsCert)
	}

	d.SetId(res.Msg.Satellite.Id)

	return diags
}

func resourceSatelliteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.SatelliteServiceClient.DeleteSatellite(ctx, connect.NewRequest(&adminv1.DeleteSatelliteRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
