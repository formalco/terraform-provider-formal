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

func ResourceIntegrationCloud() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Cloud integration.",
		CreateContext: resourceIntegrationCloudCreate,
		ReadContext:   resourceIntegrationCloudRead,
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
				Description: "Type of the Integration mfa app: `datahub`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Webhook secret of the Integration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"region": {
				// This description is used by the documentation generator and the language server.
				Description: "Region of the cloud provider.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"aws_region": {
				// This description is used by the documentation generator and the language server.
				Description: "AWS Region.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_account_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Account ID of AWS account.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_iam_role": {
				// This description is used by the documentation generator and the language server.
				Description: "AWS Iam Role used by Formal.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_id": {
				// This description is used by the documentation generator and the language server.
				Description: "AWS Formal ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aws_formal_stack_name": {
				// This description is used by the documentation generator and the language server.
				Description: "Cloud formation stack name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceIntegrationCloudCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	Name := d.Get("name").(string)
	Type := d.Get("type").(string)
	Region := d.Get("region").(string)

	res, err := c.Grpc.Sdk.IntegrationCloudServiceClient.CreateCloudIntegration(ctx, connect.NewRequest(&corev1.CreateCloudIntegrationRequest{
		Name:        Name,
		Type:        Type,
		CloudRegion: Region,
	}))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.Msg.Cloud.Id)

	resourceIntegrationCloudRead(ctx, d, meta)
	return diags
}

func resourceIntegrationCloudRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	res, err := c.Grpc.Sdk.IntegrationCloudServiceClient.GetIntegrationCloud(ctx, connect.NewRequest(&corev1.GetIntegrationCloudRequest{
		Id: appId,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	if res.Msg.Cloud == nil {
		d.SetId("")
		return diags
	}

	d.Set("name", res.Msg.Cloud.Name)

	switch data := res.Msg.Cloud.Cloud.(type) {
	case *corev1.CloudIntegration_Aws:
		d.Set("aws_region", data.Aws.AwsCloudRegion)
		d.Set("aws_account_id", data.Aws.AwsAccountId)
		d.Set("aws_formal_iam_role", data.Aws.AwsFormalIamRole)
		d.Set("aws_formal_id", data.Aws.AwsFormalId)
		d.Set("aws_formal_stack_name", data.Aws.AwsFormalStackName)
	}

	d.SetId(res.Msg.Cloud.Id)
	return diags
}

func resourceIntegrationCloudDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	appId := d.Id()

	_, err := c.Grpc.Sdk.IntegrationCloudServiceClient.DeleteCloudIntegration(ctx, connect.NewRequest(&corev1.DeleteCloudIntegrationRequest{Id: appId}))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
