package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
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
						"enable_eks_autodiscovery": {
							Description: "Enables resource autodiscovery for EKS clusters.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"enable_rds_autodiscovery": {
							Description: "Enables resource autodiscovery for RDS instances (PostgreSQL, MySQL, MongoDB).",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"enable_redshift_autodiscovery": {
							Description: "Enables resource autodiscovery for Redshift clusters.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"enable_ecs_autodiscovery": {
							Description: "Enables resource autodiscovery for ECS clusters.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"enable_ec2_autodiscovery": {
							Description: "Enables resource autodiscovery for EC2 instances.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"enable_s3_autodiscovery": {
							Description: "Enables resource autodiscovery for S3 buckets.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"allow_s3_access": {
							Description: "Allows the Cloud Integration to access S3 buckets for Log Integrations.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"s3_bucket_arn": {
							Description: "The S3 bucket ARN this Cloud Integration is allowed to use for Log Integrations.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "*",
						},
						"aws_customer_role_arn": {
							Description: "The ARN of the IAM role that Formal assumes in your AWS account to access your resources.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
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
			"aws_enable_eks_autodiscovery": {
				Description: "Whether AWS EKS autodiscovery is enabled or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"aws_enable_rds_autodiscovery": {
				Description: "Whether AWS RDS autodiscovery is enabled or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"aws_enable_redshift_autodiscovery": {
				Description: "Whether AWS Redshift autodiscovery is enabled or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"aws_enable_ecs_autodiscovery": {
				Description: "Whether AWS ECS autodiscovery is enabled or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"aws_enable_ec2_autodiscovery": {
				Description: "Whether AWS EC2 autodiscovery is enabled or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"aws_enable_s3_autodiscovery": {
				Description: "Whether AWS S3 autodiscovery is enabled or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"aws_allow_s3_access": {
				Description: "Whether AWS S3 access is allowed or not.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"aws_s3_bucket_arn": {
				Description: "The AWS S3 bucket ARN this Cloud Integration is allowed to use for Log Integrations, if it is allowed to access S3.",
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

					for _, key := range []string{"enable_eks_autodiscovery", "enable_rds_autodiscovery", "enable_redshift_autodiscovery", "enable_ecs_autodiscovery", "enable_ec2_autodiscovery", "enable_s3_autodiscovery", "allow_s3_access", "s3_bucket_arn"} {
						oldVal, newVal := d.GetChange(fmt.Sprintf("aws.0.%s", key))
						if oldVal != newVal {
							d.SetNew(fmt.Sprintf("aws_%s", key), newVal)
						}
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
			enableEksAutodiscovery := awsConfig["enable_eks_autodiscovery"].(bool)
			enableRdsAutodiscovery := awsConfig["enable_rds_autodiscovery"].(bool)
			enableRedshiftAutodiscovery := awsConfig["enable_redshift_autodiscovery"].(bool)
			enableEcsAutodiscovery := awsConfig["enable_ecs_autodiscovery"].(bool)
			enableEc2Autodiscovery := awsConfig["enable_ec2_autodiscovery"].(bool)
			enableS3Autodiscovery := awsConfig["enable_s3_autodiscovery"].(bool)
			allowS3Access := awsConfig["allow_s3_access"].(bool)

			var customerRoleArn *string
			awsCustomerRoleArn := awsConfig["aws_customer_role_arn"].(string)
			if awsCustomerRoleArn != "" {
				customerRoleArn = &awsCustomerRoleArn
			}

			res, err := c.Grpc.Sdk.IntegrationCloudServiceClient.CreateCloudIntegration(ctx, connect.NewRequest(&corev1.CreateCloudIntegrationRequest{
				Name:        name,
				CloudRegion: cloudRegion,

				Cloud: &corev1.CreateCloudIntegrationRequest_Aws{
					Aws: &corev1.CreateCloudIntegrationRequest_AWS{
						TemplateVersion:             awsConfig["template_version"].(string),
						EnableEksAutodiscovery:      &enableEksAutodiscovery,
						EnableRdsAutodiscovery:      &enableRdsAutodiscovery,
						EnableRedshiftAutodiscovery: &enableRedshiftAutodiscovery,
						EnableEcsAutodiscovery:      &enableEcsAutodiscovery,
						EnableEc2Autodiscovery:      &enableEc2Autodiscovery,
						EnableS3Autodiscovery:       &enableS3Autodiscovery,
						AllowS3Access:               &allowS3Access,
						S3BucketArn:                 awsConfig["s3_bucket_arn"].(string),
						CustomerRoleArn:             customerRoleArn,
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

	existingAwsConfig := d.Get("aws").([]interface{})
	var existingAwsCustomerRoleArn string

	if len(existingAwsConfig) > 0 {
		existingAwsConfig := existingAwsConfig[0].(map[string]interface{})
		existingAwsCustomerRoleArn = existingAwsConfig["aws_customer_role_arn"].(string)
	}

	switch data := res.Msg.Cloud.Cloud.(type) {
	case *corev1.CloudIntegration_Aws:
		d.Set("type", "aws")
		d.Set("cloud_region", data.Aws.AwsCloudRegion)

		awsConfig := map[string]interface{}{
			"template_version":              data.Aws.AwsTemplateVersion,
			"enable_eks_autodiscovery":      data.Aws.AwsEnableEksAutodiscovery,
			"enable_rds_autodiscovery":      data.Aws.AwsEnableRdsAutodiscovery,
			"enable_redshift_autodiscovery": data.Aws.AwsEnableRedshiftAutodiscovery,
			"enable_ecs_autodiscovery":      data.Aws.AwsEnableEcsAutodiscovery,
			"enable_ec2_autodiscovery":      data.Aws.AwsEnableEc2Autodiscovery,
			"enable_s3_autodiscovery":       data.Aws.AwsEnableS3Autodiscovery,
			"allow_s3_access":               data.Aws.AwsAllowS3Access,
			"s3_bucket_arn":                 data.Aws.AwsS3BucketArn,
		}

		// Only set the customer role ARN if it was set in the existing config
		if existingAwsCustomerRoleArn != "" {
			awsConfig["aws_customer_role_arn"] = data.Aws.AwsCustomerRoleArn
		}

		if err := d.Set("aws", []interface{}{awsConfig}); err != nil {
			return diag.FromErr(err)
		}

		d.Set("aws_template_body", data.Aws.TemplateBody)
		d.Set("aws_formal_stack_name", data.Aws.AwsFormalStackName)
		d.Set("aws_formal_iam_role", data.Aws.AwsFormalIamRole)
		d.Set("aws_formal_pingback_arn", data.Aws.AwsFormalPingbackArn)

		d.Set("aws_enable_eks_autodiscovery", data.Aws.AwsEnableEksAutodiscovery)
		d.Set("aws_enable_rds_autodiscovery", data.Aws.AwsEnableRdsAutodiscovery)
		d.Set("aws_enable_redshift_autodiscovery", data.Aws.AwsEnableRedshiftAutodiscovery)
		d.Set("aws_enable_ecs_autodiscovery", data.Aws.AwsEnableEcsAutodiscovery)
		d.Set("aws_enable_ec2_autodiscovery", data.Aws.AwsEnableEc2Autodiscovery)
		d.Set("aws_enable_s3_autodiscovery", data.Aws.AwsEnableS3Autodiscovery)
		d.Set("aws_allow_s3_access", data.Aws.AwsAllowS3Access)
		d.Set("aws_s3_bucket_arn", data.Aws.AwsS3BucketArn)
	}

	return diags
}

func resourceIntegrationCloudUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	integrationId := d.Id()

	fieldsThatCanBeUpdated := []string{"aws"}

	// These fields can't be updated, but they can still be changed by
	// CustomizeDiff when their 'aws.0.' counterpart has changes
	fieldsThatCanChange := append(fieldsThatCanBeUpdated, []string{"aws_enable_eks_autodiscovery", "aws_enable_rds_autodiscovery", "aws_enable_redshift_autodiscovery", "aws_enable_ecs_autodiscovery", "aws_enable_ec2_autodiscovery", "aws_enable_s3_autodiscovery", "aws_allow_s3_access", "aws_s3_bucket_arn"}...)

	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanBeUpdated, ", "))
	}

	var err error

	if v, ok := d.GetOk("aws"); ok {
		awsConfigs := v.([]interface{})
		if len(awsConfigs) > 0 {
			awsConfig := awsConfigs[0].(map[string]interface{})
			enableEksAutodiscovery := awsConfig["enable_eks_autodiscovery"].(bool)
			enableRdsAutodiscovery := awsConfig["enable_rds_autodiscovery"].(bool)
			enableRedshiftAutodiscovery := awsConfig["enable_redshift_autodiscovery"].(bool)
			enableEcsAutodiscovery := awsConfig["enable_ecs_autodiscovery"].(bool)
			enableEc2Autodiscovery := awsConfig["enable_ec2_autodiscovery"].(bool)
			enableS3Autodiscovery := awsConfig["enable_s3_autodiscovery"].(bool)
			allowS3Access := awsConfig["allow_s3_access"].(bool)

			// Don't attempt to change a customer role ARN if it was computed from CloudFormation
			var customerRoleArn *string
			awsCustomerRoleArn := awsConfig["aws_customer_role_arn"].(string)
			if awsCustomerRoleArn != "" {
				customerRoleArn = &awsCustomerRoleArn
			}

			_, err = c.Grpc.Sdk.IntegrationCloudServiceClient.UpdateCloudIntegration(ctx, connect.NewRequest(&corev1.UpdateCloudIntegrationRequest{
				Id: integrationId,
				Cloud: &corev1.UpdateCloudIntegrationRequest_Aws{
					Aws: &corev1.UpdateCloudIntegrationRequest_AWS{
						TemplateVersion:             awsConfig["template_version"].(string),
						EnableEksAutodiscovery:      &enableEksAutodiscovery,
						EnableRdsAutodiscovery:      &enableRdsAutodiscovery,
						EnableRedshiftAutodiscovery: &enableRedshiftAutodiscovery,
						EnableEcsAutodiscovery:      &enableEcsAutodiscovery,
						EnableEc2Autodiscovery:      &enableEc2Autodiscovery,
						EnableS3Autodiscovery:       &enableS3Autodiscovery,
						AllowS3Access:               &allowS3Access,
						S3BucketArn:                 awsConfig["s3_bucket_arn"].(string),
						CustomerRoleArn:             customerRoleArn,
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
