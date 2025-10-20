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

func ResourceIntegrationDataCatalog() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:        "(Deprecated) Registering a Data Catalog integration.",
		DeprecationMessage: "This resource is deprecated and will be removed in a future version. DataHub integration support is being phased out.",
		CreateContext:      resourceIntegrationDataCatalogCreate,
		ReadContext:        resourceIntegrationDataCatalogRead,
		UpdateContext:      resourceIntegrationDataCatalogUpdate,
		DeleteContext:      resourceIntegrationDataCatalogDelete,

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
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Name of the Integration",
				Type:        schema.TypeString,
				Required:    true,
			},
			"sync_direction": {
				// This description is used by the documentation generator and the language server.
				Description: "Sync direction of the Integration: supported values are 'bidirectional', 'formal_to_datahub', 'datahub_to_formal'.",
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
			"datahub": {
				Description: "Configuration block for Datahub integration. This block is optional and may be omitted if not configuring a Datahub integration.",
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"webhook_secret": {
							Description: "Webhook secret of the Datahub instance.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"generalized_metadata_service_url": {
							Description: "Generalized metadata service url for the Datahub instance.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"api_key": {
							Description: "Api Key for the Datahub instance.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
					},
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

	var res *connect.Response[corev1.CreateDataCatalogIntegrationResponse]
	var err error

	if v, ok := d.GetOk("datahub"); ok {
		// Since 'datahub' is a TypeSet, get the first element as it's expected to have only one item.
		datahubConfigs := v.(*schema.Set).List()
		if len(datahubConfigs) > 0 {
			datahubConfig := datahubConfigs[0].(map[string]interface{})

			datahub := &corev1.CreateDataCatalogIntegrationRequest_Datahub{
				Datahub: &corev1.Datahub{
					ApiKey:                        datahubConfig["api_key"].(string),
					GeneralizedMetadataServiceUrl: datahubConfig["generalized_metadata_service_url"].(string),
					// Assuming 'syncedEntities' is defined earlier in your code and relevant here.
					SyncedEntities: syncedEntities,
				},
			}

			res, err = c.Grpc.Sdk.IntegrationDataCatalogServiceClient.CreateDataCatalogIntegration(ctx, connect.NewRequest(&corev1.CreateDataCatalogIntegrationRequest{
				Name:          Name,          // Assuming 'Name' is defined and contains the name of the integration.
				SyncDirection: SyncDirection, // Assuming 'SyncDirection' is defined and specifies the synchronization direction.
				Catalog:       datahub,
			}))
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.Errorf("Datahub configuration is required.")
		}
	} else {
		return diag.Errorf("Unsupported data catalog type or missing configuration")
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
		d.Set("synced_entities", data.Datahub.SyncedEntities)
		datahubConfig := map[string]interface{}{
			"api_key":                          data.Datahub.ApiKey,
			"generalized_metadata_service_url": data.Datahub.GeneralizedMetadataServiceUrl,
			"webhook_secret":                   data.Datahub.WebhookSecret,
		}

		// Create a new set with the 'datahub' configuration map
		datahubSet := schema.NewSet(schema.HashResource(ResourceIntegrationDataCatalog()), []interface{}{datahubConfig})
		if err := d.Set("datahub", datahubSet); err != nil {
			return diag.FromErr(err)
		}
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

	integration := &corev1.IntegrationDataCatalog{
		Id:   d.Id(),
		Name: d.Get("name").(string),
	}
	if v, ok := d.GetOk("datahub"); ok {
		// Since 'datahub' is a TypeSet, get the first element as it's expected to have only one item.
		datahubConfigs := v.(*schema.Set).List()
		if len(datahubConfigs) > 0 {
			datahubConfig := datahubConfigs[0].(map[string]interface{})

			datahub := &corev1.IntegrationDataCatalog_Datahub{
				Datahub: &corev1.Datahub{
					WebhookSecret:                 datahubConfig["webhook_secret"].(string),
					ApiKey:                        datahubConfig["api_key"].(string),
					GeneralizedMetadataServiceUrl: datahubConfig["generalized_metadata_service_url"].(string),
					SyncedEntities:                syncedEntities,
				},
			}
			integration.Catalog = datahub
		} else {
			return diag.Errorf("Datahub configuration is required for 'datahub' type catalog.")
		}
	} else {
		return diag.Errorf("Unsupported data catalog type")
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
