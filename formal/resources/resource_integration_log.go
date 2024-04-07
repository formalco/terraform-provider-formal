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
				Description: "Type of the Integration app: `datadog`, `splunk` or `s3`.",
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
			"splunk_host": {
				// This description is used by the documentation generator and the language server.
				Description: "Url of your Splunk app.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"splunk_port": {
				// This description is used by the documentation generator and the language server.
				Description: "Port of your Splunk app.",
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
			},
			"splunk_access_token": {
				// This description is used by the documentation generator and the language server.
				Description: "Access Token of Splunk.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"aws_access_key_id": {
				Description: "AWS Access Key ID. Required if type is s3.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"aws_access_key_secret": {
				Description: "AWS Access Key Secret. Required if type is s3.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"aws_region": {
				Description: "AWS Region. Required if type is s3.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"aws_s3_bucket_name": {
				Description: "AWS S3 Bucket Name. Required if type is s3.",
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

	integrationType := d.Get("type").(string)
	var res *connect.Response[corev1.CreateIntegrationLogResponse]
	var err error

	switch integrationType {
	case "splunk":
		integration := &corev1.CreateIntegrationLogRequest_Splunk_{
			Splunk: &corev1.CreateIntegrationLogRequest_Splunk{
				Host:        d.Get("splunk_host").(string),
				Port:        int32(d.Get("splunk_port").(int)),
				AccessToken: d.Get("splunk_access_token").(string),
			},
		}
		res, err = c.Grpc.Sdk.IntegrationsLogServiceClient.CreateIntegrationLog(ctx, connect.NewRequest(&corev1.CreateIntegrationLogRequest{
			Name:        name,
			Integration: integration,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	case "aws_s3":
		integration := &corev1.CreateIntegrationLogRequest_AwsS3_{
			AwsS3: &corev1.CreateIntegrationLogRequest_AwsS3{
				AccessKeyId:     d.Get("aws_access_key_id").(string),
				SecretAccessKey: d.Get("aws_access_key_secret").(string),
				Region:          d.Get("aws_region").(string),
				BucketName:      d.Get("aws_s3_bucket_name").(string),
			},
		}
		res, err = c.Grpc.Sdk.IntegrationsLogServiceClient.CreateIntegrationLog(ctx, connect.NewRequest(&corev1.CreateIntegrationLogRequest{
			Name:        name,
			Integration: integration,
		}))
		if err != nil {
			return diag.FromErr(err)
		}
	case "datadog":
		integration := &corev1.CreateIntegrationLogRequest_Datadog_{
			Datadog: &corev1.CreateIntegrationLogRequest_Datadog{
				Site:      d.Get("dd_site").(string),
				ApiKey:    d.Get("dd_api_key").(string),
				AccountId: d.Get("dd_account_id").(string),
			},
		}
		res, err = c.Grpc.Sdk.IntegrationsLogServiceClient.CreateIntegrationLog(ctx, connect.NewRequest(&corev1.CreateIntegrationLogRequest{
			Name:        name,
			Integration: integration,
		}))
		if err != nil {
			return diag.FromErr(err)
		}

	default:
		return diag.Errorf("Unsupported integration type: %s", integrationType)
	}

	d.SetId(res.Msg.Integration.Id)

	resourceIntegrationLogsRead(ctx, d, m)
	return diags
}

func resourceIntegrationLogsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	res, err := c.Grpc.Sdk.IntegrationsLogServiceClient.GetIntegrationLog(ctx, connect.NewRequest(&corev1.GetIntegrationLogRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", res.Msg.Integration.Name)
	d.Set("integration", res.Msg.Integration.Integration)
	d.Set("termination_protection", res.Msg.Integration.TerminationProtection)
	d.Set("created_at", res.Msg.Integration.CreatedAt.AsTime().Unix())

	d.SetId(res.Msg.Integration.Id)

	return diags
}

func resourceIntegrationLogsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	_, err := c.Grpc.Sdk.IntegrationsLogServiceClient.DeleteIntegrationLog(ctx, connect.NewRequest(&corev1.DeleteIntegrationLogRequest{
		Id: id,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
