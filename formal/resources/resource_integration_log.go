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
			"datadog": {
				Description:   "Configuration block for Datadog integration.",
				Type:          schema.TypeSet,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"splunk", "aws_s3"},
				ForceNew:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"site": {
							Description: "URL of your Datadog app.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"api_key": {
							Description: "API Key of Datadog.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"account_id": {
							Description: "Account ID of Datadog.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"splunk": {
				Description:   "Configuration block for Splunk integration.",
				Type:          schema.TypeSet,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"datadog", "aws_s3"},
				ForceNew:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Description: "URL of your Splunk app.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"port": {
							Description: "Port of your Splunk app.",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"access_token": {
							Description: "Access Token of Splunk.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"aws_s3": {
				Description:   "Configuration block for AWS S3 integration.",
				Type:          schema.TypeSet,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"splunk", "datadog"},
				ForceNew:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key_id": {
							Description: "AWS Access Key ID.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"access_key_secret": {
							Description: "AWS Access Key Secret.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"region": {
							Description: "AWS Region.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"s3_bucket_name": {
							Description: "AWS S3 Bucket Name.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceIntegrationLogsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*clients.Clients)

	var diags diag.Diagnostics

	name := d.Get("name").(string)

	var res *connect.Response[corev1.CreateIntegrationLogResponse]
	var err error

	// Check if Datadog is configured
	if v, ok := d.GetOk("datadog"); ok {
		ddConfigs := v.(*schema.Set).List()
		if len(ddConfigs) > 0 {
			ddConfig := ddConfigs[0].(map[string]interface{})

			integration := &corev1.CreateIntegrationLogRequest_Datadog_{
				Datadog: &corev1.CreateIntegrationLogRequest_Datadog{
					Site:      ddConfig["site"].(string),
					ApiKey:    ddConfig["api_key"].(string),
					AccountId: ddConfig["account_id"].(string),
				},
			}
			res, err = c.Grpc.Sdk.IntegrationsLogServiceClient.CreateIntegrationLog(ctx, connect.NewRequest(&corev1.CreateIntegrationLogRequest{
				Name:        name,
				Integration: integration,
			}))
		}
	}

	// Check if Splunk is configured
	if v, ok := d.GetOk("splunk"); ok {
		splunkConfigs := v.(*schema.Set).List()
		if len(splunkConfigs) > 0 {
			splunkConfig := splunkConfigs[0].(map[string]interface{})

			integration := &corev1.CreateIntegrationLogRequest_Splunk_{
				Splunk: &corev1.CreateIntegrationLogRequest_Splunk{
					Host:        splunkConfig["host"].(string),
					Port:        int32(splunkConfig["port"].(int)),
					AccessToken: splunkConfig["access_token"].(string),
				},
			}
			res, err = c.Grpc.Sdk.IntegrationsLogServiceClient.CreateIntegrationLog(ctx, connect.NewRequest(&corev1.CreateIntegrationLogRequest{
				Name:        name,
				Integration: integration,
			}))
		}
	}

	// Check if AWS S3 is configured
	if v, ok := d.GetOk("aws_s3"); ok {
		awsConfigs := v.(*schema.Set).List()
		if len(awsConfigs) > 0 {
			awsConfig := awsConfigs[0].(map[string]interface{})

			integration := &corev1.CreateIntegrationLogRequest_AwsS3_{
				AwsS3: &corev1.CreateIntegrationLogRequest_AwsS3{
					AccessKeyId:     awsConfig["access_key_id"].(string),
					SecretAccessKey: awsConfig["access_key_secret"].(string),
					Region:          awsConfig["region"].(string),
					BucketName:      awsConfig["s3_bucket_name"].(string),
				},
			}
			res, err = c.Grpc.Sdk.IntegrationsLogServiceClient.CreateIntegrationLog(ctx, connect.NewRequest(&corev1.CreateIntegrationLogRequest{
				Name:        name,
				Integration: integration,
			}))
		}
	}

	// Handle error if any
	if err != nil {
		return diag.FromErr(err)
	}

	// Assuming you need to handle a situation where none are configured
	if res == nil {
		return diag.Errorf("No integration configuration found")
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
