package resource

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"errors"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceCloudAccount() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Integrate a Cloud Account with Formal to deploy Managed Cloud resources. Note that the paired CloudFormation Stack must be deployed in eu-west-1 or us-east-1.",

		CreateContext: resourceCloudAccountCreate,
		ReadContext:   resourceCloudAccountRead,
		UpdateContext: resourceCloudAccountUpdate,
		DeleteContext: resourceCloudAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "Use this value as the cloud_account_id for formal managed resources.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cloud_account_name": {
				// This description is used by the documentation generator and the language server.
				Description: "A friendly name to refer to this Cloud Account when using Formal.",
				Type:        schema.TypeString,
				Optional:    true,
				// ForceNew:    true,
			},
			"aws_cloud_region": {
				// This description is used by the documentation generator and the language server.
				Description: "The AWS Region you would like to deploy the CloudFormation stack in. Supported values are us-east-1, us-east-2, and eu-west-1.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"cloud_provider": {
				// This description is used by the documentation generator and the language server.
				Description: "The Cloud Provider you are connecting the cloud account from. The only currently supported value is `aws`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"aws_formal_stack_name": {
				// This description is used by the documentation generator and the language server.
				Description: "Use this value for the name field for your aws_cloudformation_stack resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Use this value for the parameters.FormalID field for your aws_cloudformation_stack resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_iam_role": {
				// This description is used by the documentation generator and the language server.
				Description: "Use this value for the parameters.FormalIamRole field for your aws_cloudformation_stack resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_handshake_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Use this value for the parameters.FormalHandshakeID field for your aws_cloudformation_stack resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_pingback_arn": {
				// This description is used by the documentation generator and the language server.
				Description: "Use this value for the parameters.FormalPingbackArn field for your aws_cloudformation_stack resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_template_body": {
				// This description is used by the documentation generator and the language server.
				Description: "Use this value for the template_body field for your aws_cloudformation_stack resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_r53_private_hosted_zone_id": {
				// This description is used by the documentation generator and the language server.
				Description: "This is the id of the AWS Route 53 Private Zone Formal creates in your account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceCloudAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	cloudAccountName := d.Get("cloud_account_name").(string)
	cloudProvider := d.Get("cloud_provider").(string)
	awsCloudRegion := d.Get("aws_cloud_region").(string)

	if cloudProvider != "aws" {
		return diag.FromErr(errors.New("cloud_provider must be 'aws'"))
	}

	awsConnectionSession, err := c.Grpc.Sdk.CloudServiceClient.CreateAwsConnectionSession(ctx, connect.NewRequest(&adminv1.CreateAwsConnectionSessionRequest{
		CloudAccountName:   cloudAccountName,
		CloudAccountRegion: awsCloudRegion,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	// Delay before creating CloudFormation Stack
	time.Sleep(10 * time.Second)

	d.SetId(awsConnectionSession.Msg.CloudIntegration.Id)

	resourceCloudAccountRead(ctx, d, meta)
	return diags
}

func resourceCloudAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.CloudServiceClient.GetIntegrationsCloudAccountById(ctx, connect.NewRequest(&adminv1.GetIntegrationsCloudAccountByIdRequest{Id: d.Id()}))
	if err != nil {
		if status.Code(err) == codes.NotFound {
			tflog.Warn(ctx, "The Cloud Account was not found, which means the stack was deleted or the integration was deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Read AWS fields
	d.Set("cloud_account_name", res.Msg.Integration.CloudAccountName)
	d.Set("cloud_provider", res.Msg.Integration.CloudProvider)
	d.Set("aws_cloud_region", res.Msg.Integration.AwsCloudRegion)
	d.Set("aws_formal_id", res.Msg.Integration.AwsFormalId)
	d.Set("aws_formal_iam_role", res.Msg.Integration.AwsFormalIamRole)
	d.Set("aws_formal_handshake_id", res.Msg.Integration.AwsFormalHandshakeId)
	d.Set("aws_formal_pingback_arn", res.Msg.Integration.AwsFormalPingbackArn)
	d.Set("aws_formal_stack_name", res.Msg.Integration.AwsFormalStackName)
	d.Set("aws_formal_template_body", res.Msg.Integration.TemplateBody)
	d.Set("aws_formal_r53_private_hosted_zone_id", res.Msg.Integration.AwsFormalR53PrivateHostedZoneId)
	d.Set("id", res.Msg.Integration.Id)

	d.SetId(res.Msg.Integration.Id)

	return diags
}

func resourceCloudAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Cloud Account links are not updateable at the moment. Please create a new one.")
}

func resourceCloudAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	accountId := d.Id()

	_, err := c.Grpc.Sdk.CloudServiceClient.DeleteAwsConnectionSession(ctx, connect.NewRequest(&adminv1.DeleteAwsConnectionSessionRequest{Id: accountId}))
	if err != nil {
		return diags
	}
	if err != nil {
		if status.Code(err) == codes.NotFound {
			tflog.Warn(ctx, "The Cloud Account was not found, which means the stack was deleted, likely by CloudFormation.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
