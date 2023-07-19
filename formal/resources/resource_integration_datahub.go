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
			"sync_direction": {
				// This description is used by the documentation generator and the language server.
				Description: "Sync direction of the Integration: supported values are 'bidirectional', 'formal_to_datahub', 'datahub_to_formal'.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"generalized_metadata_service_url": {
				// This description is used by the documentation generator and the language server.
				Description: "Generalized Metadata Service URL of the Integration.",
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
			"synced_entities": {
				// This description is used by the documentation generator and the language server.
				Description: "Synced entities of the Integration: currently supported values are 'tags', 'data_labels'.",
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

	var syncedEntities []string
	for _, syncedEntity := range d.Get("synced_entities").([]interface{}) {
		syncedEntities = append(syncedEntities, syncedEntity.(string))
	}

	SyncDirection := d.Get("sync_direction").(string)
	GeneralizedMetadataServiceUrl := d.Get("generalized_metadata_service_url").(string)
	Active := d.Get("active").(bool)
	ApiKey := d.Get("api_key").(string)

	res, err := c.Grpc.Sdk.DatahubServiceClient.CreateDatahubIntegration(ctx, connect.NewRequest(&adminv1.CreateDatahubIntegrationRequest{
		SyncDirection:                 SyncDirection,
		GeneralizedMetadataServiceUrl: GeneralizedMetadataServiceUrl,
		SyncedEntities:                syncedEntities,
		Active:                        Active,
		ApiKey:                        ApiKey,
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

	d.Set("sync_direction", res.Msg.Integration.SyncDirection)
	d.Set("generalized_metadata_service_url", res.Msg.Integration.GeneralizedMetadataServiceUrl)
	d.Set("webhook_secret", res.Msg.Integration.WebhookSecret)
	d.Set("organization_id", res.Msg.Integration.OrganizationId)
	d.Set("active", res.Msg.Integration.Active)
	d.Set("synced_entities", res.Msg.Integration.SyncedEntities)

	d.SetId(res.Msg.Integration.Id)
	return diags
}

func resourceIntegrationDatahubUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	var syncedEntities []string
	for _, syncedEntity := range d.Get("synced_entities").([]interface{}) {
		syncedEntities = append(syncedEntities, syncedEntity.(string))
	}

	SyncDirection := d.Get("sync_direction").(string)
	GeneralizedMetadataServiceUrl := d.Get("generalized_metadata_service_url").(string)
	Active := d.Get("active").(bool)
	ApiKey := d.Get("api_key").(string)

	_, err := c.Grpc.Sdk.DatahubServiceClient.UpdateDatahubIntegration(ctx, connect.NewRequest(&adminv1.UpdateDatahubIntegrationRequest{
		Id:                            d.Id(),
		SyncDirection:                 SyncDirection,
		GeneralizedMetadataServiceUrl: GeneralizedMetadataServiceUrl,
		SyncedEntities:                syncedEntities,
		Active:                        Active,
		ApiKey:                        ApiKey,
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
