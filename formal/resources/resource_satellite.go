package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	res, err := c.Grpc.Sdk.SatelliteServiceClient.CreateSatellite(ctx, connect.NewRequest(&adminv1.CreateSatelliteRequest{
		Name: name,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	const ErrorTolerance = 5
	currentErrors := 0
	for {
		// Retrieve status
		createdSatellite, err := c.Grpc.Sdk.SatelliteServiceClient.GetSatelliteById(ctx, connect.NewRequest(&adminv1.GetSatelliteByIdRequest{Id: res.Msg.Id}))
		if err != nil {
			if currentErrors >= ErrorTolerance {
				return diag.FromErr(err)
			} else {
				tflog.Warn(ctx, "Experienced an error #"+strconv.Itoa(currentErrors+1)+" retrieving Satellite: ", map[string]interface{}{"err": err})
				currentErrors += 1
				time.Sleep(15 * time.Second)
				continue
			}
		}

		if createdSatellite.Msg.Satellite == nil {
			err = errors.New("satellite with the given ID not found. It may have been deleted")
			return diag.FromErr(err)
		}

		tflog.Info(ctx, "Satellite status is: "+fmt.Sprint(createdSatellite.Msg.Satellite.Status))
		// Check status
		if createdSatellite.Msg.Satellite.Status == "ready" {
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}

	d.SetId(res.Msg.Id)

	resourceSatelliteRead(ctx, d, meta)

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
	d.Set("termination_protection", res.Msg.Satellite.TerminationProtection)

	if c.Grpc.ReturnSensitiveValue {
		res, err := c.Grpc.Sdk.SatelliteServiceClient.GetSatelliteApiKey(ctx, connect.NewRequest(&adminv1.GetSatelliteApiKeyRequest{Id: d.Id()}))
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
		_, err := c.Grpc.Sdk.SatelliteServiceClient.UpdateSatellite(ctx, connect.NewRequest(&adminv1.UpdateSatelliteRequest{
			Id:                    satelliteId,
			TerminationProtection: d.Get("termination_protection").(bool),
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

	_, err := c.Grpc.Sdk.SatelliteServiceClient.DeleteSatellite(ctx, connect.NewRequest(&adminv1.DeleteSatelliteRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
