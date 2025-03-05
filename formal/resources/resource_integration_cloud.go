package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceIntegrationCloud() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Cloud integration.",
		CreateContext: resourceIntegrationCloudCreate,
		ReadContext:   resourceIntegrationCloudRead,
		UpdateContext: resourceIntegrationCloudUpdate,
		DeleteContext: resourceIntegrationCloudDelete,

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
				Description: "Type of the Integration. (Supported: aws)",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "aws",
				Deprecated:  "This field is deprecated and will be removed in a future version.",
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Name of the Integration.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"cloud_region": {
				// This description is used by the documentation generator and the language server.
				Description: "Region of the cloud provider.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"aws": {
				Description: "Configuration block for AWS integration.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"template_version": {
							Description: "The template version of the CloudFormation stack. Use `latest` to stay in sync.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"s3_bucket_arn": {
							Description: "The S3 bucket ARN this Cloud Integration is allowed to use for Log Integrations.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "*",
						},
					},
				},
			},
			"aws_template_body": {
				Description: "The template body of the CloudFormation stack.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_stack_name": {
				Description: "A generated name for your CloudFormation stack.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_iam_role": {
				Description: "The IAM role ID Formal will use to access your resources.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_pingback_arn": {
				Description: "The SNS topic ARN CloudFormation can use to send events to Formal.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
			if v, ok := d.GetOk("aws"); ok {
				awsConfigs := v.([]interface{})
				if len(awsConfigs) > 0 {
					oldVersion, newVersion := d.GetChange("aws.0.template_version")
					if oldVersion != newVersion {
						d.SetNewComputed("aws_template_body")
					}
				}
			}
			return nil
		},
	}
}

func resourceIntegrationCloudCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	name := d.Get("name").(string)
	cloudRegion := d.Get("cloud_region").(string)

	if v, ok := d.GetOk("aws"); ok {
		awsConfigs := v.([]interface{})
		if len(awsConfigs) > 0 {
			awsConfig := awsConfigs[0].(map[string]interface{})

			res, err := c.Grpc.Sdk.IntegrationCloudServiceClient.CreateCloudIntegration(ctx, connect.NewRequest(&corev1.CreateCloudIntegrationRequest{
				Name:        name,
				CloudRegion: cloudRegion,

				Cloud: &corev1.CreateCloudIntegrationRequest_Aws{
					Aws: &corev1.CreateCloudIntegrationRequest_AWS{
						S3BucketArn:     awsConfig["s3_bucket_arn"].(string),
						TemplateVersion: awsConfig["template_version"].(string),
					},
				},
			}))
			if err != nil {
				return diag.FromErr(err)
			}
			d.SetId(res.Msg.Id)
		}
	}
	return resourceIntegrationCloudRead(ctx, d, meta)
}

func resourceIntegrationCloudRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	integrationId := d.Id()

	res, err := c.Grpc.Sdk.IntegrationCloudServiceClient.GetIntegrationCloud(ctx, connect.NewRequest(&corev1.GetIntegrationCloudRequest{
		Id: integrationId,
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Integration was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Cloud.Id)
	d.Set("name", res.Msg.Cloud.Name)

	switch data := res.Msg.Cloud.Cloud.(type) {
	case *corev1.CloudIntegration_Aws:
		d.Set("type", "aws")
		d.Set("cloud_region", data.Aws.AwsCloudRegion)

		awsConfig := map[string]interface{}{
			"template_version": data.Aws.AwsTemplateVersion,
			"s3_bucket_arn":    data.Aws.AwsS3BucketArn,
		}
		if err := d.Set("aws", []interface{}{awsConfig}); err != nil {
			return diag.FromErr(err)
		}

		d.Set("aws_template_body", data.Aws.TemplateBody)
		d.Set("aws_formal_stack_name", data.Aws.AwsFormalStackName)
		d.Set("aws_formal_iam_role", data.Aws.AwsFormalIamRole)
		d.Set("aws_formal_pingback_arn", data.Aws.AwsFormalPingbackArn)
	}

	return diags
}

func resourceIntegrationCloudUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	integrationId := d.Id()

	fieldsThatCanChange := []string{"aws"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	var err error

	if v, ok := d.GetOk("aws"); ok {
		awsConfigs := v.([]interface{})
		if len(awsConfigs) > 0 {
			awsConfig := awsConfigs[0].(map[string]interface{})

			awsTemplateVersion := awsConfig["template_version"].(string)
			awsS3BucketArn := awsConfig["s3_bucket_arn"].(string)

			_, err = c.Grpc.Sdk.IntegrationCloudServiceClient.UpdateCloudIntegration(ctx, connect.NewRequest(&corev1.UpdateCloudIntegrationRequest{
				Id: integrationId,
				Cloud: &corev1.UpdateCloudIntegrationRequest_Aws{
					Aws: &corev1.UpdateCloudIntegrationRequest_AWS{
						TemplateVersion: awsTemplateVersion,
						S3BucketArn:     awsS3BucketArn,
					},
				},
			}))
		}
	}

	if err != nil {
		return diag.FromErr(err)
	}
	return resourceIntegrationCloudRead(ctx, d, meta)
}

func resourceIntegrationCloudDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	integrationId := d.Id()

	_, err := c.Grpc.Sdk.IntegrationCloudServiceClient.DeleteCloudIntegration(ctx, connect.NewRequest(&corev1.DeleteCloudIntegrationRequest{Id: integrationId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
