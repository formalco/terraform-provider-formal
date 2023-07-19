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

func ResourceIntegrationDatahub() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Datahub integration.",
		CreateContext: resourceIntegrationDatahubCreate,
		ReadContext:   resourceIntegrationDatahubRead,
		UpdateContext: resourceIntegrationDatahubUpdate,
		DeleteContext: resourceIntegrationDatahubDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Integration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"webhook_secret": {
				// This description is used by the documentation generator and the language server.
				Description: "Webhook secret of the Integration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"communication_type": {
				// This description is used by the documentation generator and the language server.
				Description: "Communication type of the Integration: supported values are 'bidirectional', 'formal_to_datahub', 'datahub_to_formal'.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"gms_url": {
				// This description is used by the documentation generator and the language server.
				Description: "GMS URL of the Integration.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"active": {
				// This description is used by the documentation generator and the language server.
				Description: "Active status of the Integration.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"api_key": {
				// This description is used by the documentation generator and the language server.
				Description: "Api Key for the GMS server.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"subscribed_entities": {
				// This description is used by the documentation generator and the language server.
				Description: "Subscribed entities of the Integration: currently supported values are 'tags', 'data_labels'.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"organization_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Organization ID of the Integration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceIntegrationDatahubCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	var subscribedEntities []string
	for _, subscribedEntity := range d.Get("subscribed_entities").([]interface{}) {
		subscribedEntities = append(subscribedEntities, subscribedEntity.(string))
	}

	CommunicationType := d.Get("communication_type").(string)
	GMSUrl := d.Get("gms_url").(string)
	Active := d.Get("active").(bool)
	ApiKey := d.Get("api_key").(string)

	res, err := c.Grpc.Sdk.DatahubServiceClient.CreateDatahubIntegration(ctx, connect.NewRequest(&adminv1.CreateDatahubIntegrationRequest{
		CommunicationType:  CommunicationType,
		GmsUrl:             GMSUrl,
		SubscribedEntities: subscribedEntities,
		Active:             Active,
		ApiKey:             ApiKey,
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.Msg.Id)

	resourceIntegrationDatahubRead(ctx, d, meta)
	return diags
}

func resourceIntegrationDatahubRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	res, err := c.Grpc.Sdk.DatahubServiceClient.GetDatahubIntegration(ctx, connect.NewRequest(&adminv1.GetDatahubIntegrationRequest{
		Id: appId,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	if res.Msg.Integration == nil {
		d.SetId("")
		return diags
	}

	d.Set("communication_type", res.Msg.Integration.CommunicationType)
	d.Set("gms_url", res.Msg.Integration.GmsUrl)
	d.Set("webhook_secret", res.Msg.Integration.WebhookSecret)
	d.Set("organization_id", res.Msg.Integration.OrganizationId)
	d.Set("active", res.Msg.Integration.Active)
	d.Set("subscribed_entities", res.Msg.Integration.SubscribedEntities)

	d.SetId(res.Msg.Integration.Id)
	return diags
}

func resourceIntegrationDatahubUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	var subscribedEntities []string
	for _, subscribedEntity := range d.Get("subscribed_entities").([]interface{}) {
		subscribedEntities = append(subscribedEntities, subscribedEntity.(string))
	}

	CommunicationType := d.Get("communication_type").(string)
	GMSUrl := d.Get("gms_url").(string)
	Active := d.Get("active").(bool)
	ApiKey := d.Get("api_key").(string)

	_, err := c.Grpc.Sdk.DatahubServiceClient.UpdateDatahubIntegration(ctx, connect.NewRequest(&adminv1.UpdateDatahubIntegrationRequest{
		Id:                 d.Id(),
		CommunicationType:  CommunicationType,
		GmsUrl:             GMSUrl,
		SubscribedEntities: subscribedEntities,
		Active:             Active,
		ApiKey:             ApiKey,
	}))

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceIntegrationDatahubDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	_, err := c.Grpc.Sdk.DatahubServiceClient.DeleteDatahubIntegration(ctx, connect.NewRequest(&adminv1.DeleteDatahubIntegrationRequest{Id: appId}))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
