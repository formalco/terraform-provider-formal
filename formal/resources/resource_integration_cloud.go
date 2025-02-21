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
				Required:    true,
				ForceNew:    true,
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
	}
}

func resourceIntegrationCloudCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	res, err := c.Grpc.Sdk.IntegrationCloudServiceClient.CreateCloudIntegration(ctx, connect.NewRequest(&corev1.CreateCloudIntegrationRequest{
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		CloudRegion: d.Get("cloud_region").(string),
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.Msg.Id)

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

	fieldsThatCanChange := []string{"cloud_region"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	cloudRegion := d.Get("cloud_region").(string)

	_, err := c.Grpc.Sdk.IntegrationCloudServiceClient.UpdateCloudIntegration(ctx, connect.NewRequest(&corev1.UpdateCloudIntegrationRequest{
		Id:          integrationId,
		CloudRegion: &cloudRegion,
	}))
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
