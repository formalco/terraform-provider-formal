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

func ResourceIntegrationLogs() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Integration Logs app.",
		CreateContext: resourceIntegrationLogsCreate,
		ReadContext:   resourceIntegrationLogsRead,
		DeleteContext: resourceIntegrationLogsDelete,

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
				Description: "Friendly name for the Integration app.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of the Integration app: datadog or splunk",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"dd_site": {
				// This description is used by the documentation generator and the language server.
				Description: "Url of your Datadog app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"dd_api_key": {
				// This description is used by the documentation generator and the language server.
				Description: "API Key of Datadog.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"dd_account_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Account ID of Datadog.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"splunk_url": {
				// This description is used by the documentation generator and the language server.
				Description: "Url of your Splunk app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"splunk_api_key": {
				// This description is used by the documentation generator and the language server.
				Description: "API Key of Splunk.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceIntegrationLogsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	typeApp := d.Get("type").(string)
	ddSite := d.Get("dd_site").(string)
	ddApiKey := d.Get("dd_api_key").(string)
	ddAccountId := d.Get("dd_account_id").(string)
	splunkUrl := d.Get("splunk_url").(string)
	splunkApiKey := d.Get("splunk_api_key").(string)

	res, err := c.Grpc.Sdk.LogsServiceClient.CreateIntegrationLogs(ctx, connect.NewRequest(&adminv1.CreateIntegrationLogsRequest{
		Name:         name,
		Type:         typeApp,
		DdSite:       ddSite,
		DdApiKey:     ddApiKey,
		DdAccountId:  ddAccountId,
		SplunkUrl:    splunkUrl,
		SplunkApiKey: splunkApiKey,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Integration.Id)

	resourceIntegrationLogsRead(ctx, d, m)
	return diags
}

func resourceIntegrationLogsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	res, err := c.Grpc.Sdk.LogsServiceClient.GetIntegrationLogById(ctx, connect.NewRequest(&adminv1.GetIntegrationLogByIdRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", res.Msg.Integration.Name)
	d.Set("type", res.Msg.Integration.Type)
	d.Set("dd_site", res.Msg.Integration.DdSite)
	d.Set("dd_account_id", res.Msg.Integration.DdAccountId)
	d.Set("splunk_url", res.Msg.Integration.SplunkUrl)

	d.SetId(res.Msg.Integration.Id)

	return diags
}

func resourceIntegrationLogsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	_, err := c.Grpc.Sdk.LogsServiceClient.DeleteIntegrationLogs(ctx, connect.NewRequest(&adminv1.DeleteIntegrationLogsRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
