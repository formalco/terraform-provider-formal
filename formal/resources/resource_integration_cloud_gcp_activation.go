package resource

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceIntegrationCloudGCPActivation() *schema.Resource {
	return &schema.Resource{
		Description:   "Reports the GCP service account and workload identity pool provider back to Formal to activate a GCP Cloud Integration.",
		CreateContext: resourceIntegrationCloudGCPActivationUpsert,
		ReadContext:   resourceIntegrationCloudGCPActivationRead,
		UpdateContext: resourceIntegrationCloudGCPActivationUpsert,
		DeleteContext: resourceIntegrationCloudGCPActivationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"integration_id": {
				Description: "The ID of the GCP Cloud Integration to activate.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"service_account_email": {
				Description: "The GCP service account email created for this integration.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"workload_identity_pool_provider": {
				Description: "The GCP workload identity pool provider created for this integration.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceIntegrationCloudGCPActivationUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	integrationId := d.Get("integration_id").(string)
	serviceAccountEmail := d.Get("service_account_email").(string)
	workloadIdentityPoolProvider := d.Get("workload_identity_pool_provider").(string)

	_, err := c.Grpc.Sdk.IntegrationCloudServiceClient.UpdateCloudIntegration(ctx, connect.NewRequest(&corev1.UpdateCloudIntegrationRequest{
		Id: integrationId,
		Cloud: &corev1.UpdateCloudIntegrationRequest_Gcp{
			Gcp: &corev1.UpdateCloudIntegrationRequest_GCP{
				ServiceAccountEmail:          &serviceAccountEmail,
				WorkloadIdentityPoolProvider: &workloadIdentityPoolProvider,
			},
		},
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(integrationId)
	return resourceIntegrationCloudGCPActivationRead(ctx, d, meta)
}

func resourceIntegrationCloudGCPActivationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	integrationId := d.Id()

	res, err := c.Grpc.Sdk.IntegrationCloudServiceClient.GetIntegrationCloud(ctx, connect.NewRequest(&corev1.GetIntegrationCloudRequest{
		Id: integrationId,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Integration was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("integration_id", res.Msg.Cloud.Id)

	if gcp, ok := res.Msg.Cloud.Cloud.(*corev1.CloudIntegration_Gcp); ok {
		d.Set("service_account_email", gcp.Gcp.GcpServiceAccountEmail)
		d.Set("workload_identity_pool_provider", gcp.Gcp.GcpWorkloadIdentityPoolProvider)
	}

	return diags
}

func resourceIntegrationCloudGCPActivationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	d.SetId("")
	return diags
}
