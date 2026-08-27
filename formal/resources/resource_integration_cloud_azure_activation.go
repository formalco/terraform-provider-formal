package resource

import (
	"context"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceIntegrationCloudAzureActivation() *schema.Resource {
	return &schema.Resource{
		Description:   "Reports the Azure tenant and managed identity client id back to Formal to activate an Azure Cloud Integration.",
		CreateContext: resourceIntegrationCloudAzureActivationUpsert,
		ReadContext:   resourceIntegrationCloudAzureActivationRead,
		UpdateContext: resourceIntegrationCloudAzureActivationUpsert,
		DeleteContext: resourceIntegrationCloudAzureActivationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Description: "The ID of the Azure Cloud Integration to activate.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"tenant_id": {
				Description: "The Entra tenant the managed identity created for this integration belongs to.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"client_id": {
				Description: "The client id of the managed identity created for this integration.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceIntegrationCloudAzureActivationUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	integrationId := d.Get("integration_id").(string)
	tenantId := d.Get("tenant_id").(string)
	clientId := d.Get("client_id").(string)

	_, err := c.Grpc.Sdk.IntegrationCloudServiceClient.UpdateCloudIntegration(ctx, &corev1.UpdateCloudIntegrationRequest{
		Id: integrationId,
		Cloud: &corev1.UpdateCloudIntegrationRequest_Azure_{
			Azure: &corev1.UpdateCloudIntegrationRequest_Azure{
				TenantId: &tenantId,
				ClientId: &clientId,
			},
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(integrationId)
	return resourceIntegrationCloudAzureActivationRead(ctx, d, meta)
}

func resourceIntegrationCloudAzureActivationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	integrationId := d.Id()

	res, err := c.Grpc.Sdk.IntegrationCloudServiceClient.GetIntegrationCloud(ctx, &corev1.GetIntegrationCloudRequest{
		Id: integrationId,
	})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Integration was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("integration_id", res.Cloud.Id)

	if azure, ok := res.Cloud.Cloud.(*corev1.CloudIntegration_Azure); ok {
		d.Set("tenant_id", azure.Azure.AzureTenantId)
		d.Set("client_id", azure.Azure.AzureClientId)
	}

	return diags
}

func resourceIntegrationCloudAzureActivationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	d.SetId("")
	return diags
}
