package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIntegrationDataCatalog() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Data Catalog integration.",
		CreateContext: resourceIntegrationDataCatalogCreate,
		ReadContext:   resourceIntegrationDataCatalogRead,
		UpdateContext: resourceIntegrationDataCatalogUpdate,
		DeleteContext: resourceIntegrationDataCatalogDelete,

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
			"type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of the Integration mfa app: `datahub`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"webhook_secret": {
				// This description is used by the documentation generator and the language server.
				Description: "Webhook secret of the Integration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"sync_direction": {
				// This description is used by the documentation generator and the language server.
				Description: "Sync direction of the Integration: supported values are 'bidirectional', 'formal_to_DataCatalog', 'DataCatalog_to_formal'.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"generalized_metadata_service_url": {
				// This description is used by the documentation generator and the language server.
				Description: "Generalized Metadata Service URL of the Integration.",
				Type:        schema.TypeString,
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
		},
	}
}

func resourceIntegrationDataCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	var syncedEntities []string
	for _, syncedEntity := range d.Get("synced_entities").([]interface{}) {
		syncedEntities = append(syncedEntities, syncedEntity.(string))
	}

	Name := d.Get("name").(string)
	SyncDirection := d.Get("sync_direction").(string)
	typeDataCatalog := d.Get("type").(string)

	var res *connect.Response[corev1.CreateDataCatalogIntegrationResponse]
	var err error

	switch typeDataCatalog {
	case "datahub":
		datahub := &corev1.CreateDataCatalogIntegrationRequest_Datahub{
			Datahub: &corev1.Datahub{
				ApiKey:                        d.Get("api_key").(string),
				GeneralizedMetadataServiceUrl: d.Get("generalized_metadata_service_url").(string),
				SyncedEntities:                syncedEntities,
			},
		}
		res, err = c.Grpc.Sdk.IntegrationDataCatalogServiceClient.CreateDataCatalogIntegration(ctx, connect.NewRequest(&corev1.CreateDataCatalogIntegrationRequest{
			Name:          Name,
			SyncDirection: SyncDirection,
			Catalog:       datahub,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("Unsupported data catalog type: %s", typeDataCatalog)
	}

	d.SetId(res.Msg.Integration.Id)

	resourceIntegrationDataCatalogRead(ctx, d, meta)
	return diags
}

func resourceIntegrationDataCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	res, err := c.Grpc.Sdk.IntegrationDataCatalogServiceClient.GetDataCatalogIntegration(ctx, connect.NewRequest(&corev1.GetDataCatalogIntegrationRequest{
		Id: appId,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	if res.Msg.Integration == nil {
		d.SetId("")
		return diags
	}

	switch data := res.Msg.Integration.Catalog.(type) {
	case *corev1.IntegrationDataCatalog_Datahub:
		d.Set("api_key", data.Datahub.ApiKey)
		d.Set("synced_entities", data.Datahub.SyncedEntities)
		d.Set("generalized_metadata_service_url", data.Datahub.GeneralizedMetadataServiceUrl)
		d.Set("webhook_secret", data.Datahub.WebhookSecret)
	}

	d.Set("sync_direction", res.Msg.Integration.SyncDirection)

	d.SetId(res.Msg.Integration.Id)
	return diags
}

func resourceIntegrationDataCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	var syncedEntities []string
	for _, syncedEntity := range d.Get("synced_entities").([]interface{}) {
		syncedEntities = append(syncedEntities, syncedEntity.(string))
	}

	typeDataCatalog := d.Get("type").(string)
	integration := &corev1.IntegrationDataCatalog{
		Id:   d.Id(),
		Name: d.Get("name").(string),
	}
	switch typeDataCatalog {
	case "datahub":
		datahub := &corev1.IntegrationDataCatalog_Datahub{
			Datahub: &corev1.Datahub{
				WebhookSecret:                 d.Get("webhook_secret").(string),
				ApiKey:                        d.Get("api_key").(string),
				GeneralizedMetadataServiceUrl: d.Get("generalized_metadata_service_url").(string),
				SyncedEntities:                syncedEntities,
			},
		}
		integration.Catalog = datahub
	default:
		return diag.Errorf("Unsupported data catalog type: %s", typeDataCatalog)
	}
	_, err := c.Grpc.Sdk.IntegrationDataCatalogServiceClient.UpdateDataCatalogIntegration(ctx, connect.NewRequest(&corev1.UpdateDataCatalogIntegrationRequest{
		Integration: integration,
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceIntegrationDataCatalogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	_, err := c.Grpc.Sdk.IntegrationDataCatalogServiceClient.DeleteDataCatalogIntegration(ctx, connect.NewRequest(&corev1.DeleteDataCatalogIntegrationRequest{Id: appId}))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
