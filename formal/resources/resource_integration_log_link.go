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

func ResourceIntegrationLogLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Integration Logs Link.",
		CreateContext: resourceIntegrationLogLinkCreate,
		ReadContext:   resourceIntegrationLogLinkRead,
		DeleteContext: resourceIntegrationLogLinkDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the LogLink.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"integration_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Integration.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Datastore ID of the Integration.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceIntegrationLogLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	integrationId := d.Get("integration_id").(string)
	dataStoreId := d.Get("datastore_id").(string)

	res, err := c.Grpc.Sdk.LogsServiceClient.CreateLogsLinkItem(ctx, connect.NewRequest(&adminv1.CreateLogsLinkItemRequest{
		IntegrationId: integrationId,
		DatastoreId:   dataStoreId,
	}))

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.Msg.LogItemLink.Id)

	resourceIntegrationLogLinkRead(ctx, d, meta)
	return diags
}

func resourceIntegrationLogLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.LogsServiceClient.GetLogsLinkItemById(ctx, connect.NewRequest(&adminv1.GetLogsLinkItemByIdRequest{
		Id: d.Id(),
	}))

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("integration_id", res.Msg.LogItemLinks.IntegrationLogId)
	d.Set("datastore_id", res.Msg.LogItemLinks.DatastoreId)

	d.SetId(res.Msg.LogItemLinks.Id)

	return diags
}

func resourceIntegrationLogLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.LogsServiceClient.DeleteLogsLinkItem(ctx, connect.NewRequest(&adminv1.DeleteLogsLinkItemRequest{Id: d.Id()}))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
