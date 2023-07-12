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

func ResourceIntegrationIncident() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Integration Incident app.",
		CreateContext: resourceIntegrationIncidentCreate,
		ReadContext:   resourceIntegrationIncidentRead,
		DeleteContext: resourceIntegrationIncidentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the App.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for the Incident app.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of the Incident app: pagerduty or custom",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"logo": {
				// This description is used by the documentation generator and the language server.
				Description: "Logo of the Incident app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"api_key": {
				// This description is used by the documentation generator and the language server.
				Description: "API Key of the Incident app.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceIntegrationIncidentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	incidentType := d.Get("type").(string)
	logo := d.Get("logo").(string)
	apiKey := d.Get("api_key").(string)

	res, err := c.Grpc.Sdk.IncidentServiceClient.ConnectIncidentAccount(ctx, connect.NewRequest(&adminv1.ConnectIncidentAccountRequest{
		Name:   name,
		Type:   incidentType,
		Logo:   logo,
		ApiKey: apiKey,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Id)

	resourceIntegrationIncidentRead(ctx, d, meta)

	return diags
}

func resourceIntegrationIncidentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.IncidentServiceClient.GetIncidentAccountById(ctx, connect.NewRequest(&adminv1.GetIncidentAccountByIdRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", res.Msg.Account.Name)
	d.Set("type", res.Msg.Account.Type)
	d.Set("logo", res.Msg.Account.Logo)

	d.SetId(res.Msg.Account.Id)

	return diags
}

func resourceIntegrationIncidentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.IncidentServiceClient.DeleteIncidentAccountById(ctx, connect.NewRequest(&adminv1.DeleteIncidentAccountByIdRequest{
		AccountId: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
