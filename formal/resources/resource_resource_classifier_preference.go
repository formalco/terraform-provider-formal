package resource

import (
	"context"

	"buf.build/go/protovalidate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceResourceClassifierConfiguration() *schema.Resource {
	return &schema.Resource{
		Description:   "A Resource Classifier Configuration is a configuration for a resource classifier.",
		CreateContext: resourceResourceClassifierConfigurationCreate,
		ReadContext:   resourceResourceClassifierConfigurationRead,
		UpdateContext: resourceResourceClassifierConfigurationUpdate,
		DeleteContext: resourceResourceClassifierConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the Resource Classifier Configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				Description: "The ID of the Resource.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"preference": {
				Description: "The preference. Supported values are `nlp`, `llm`, `both`, and `none`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ai_analysis_timeout_seconds": {
				Description: "The timeout for the AI analysis in seconds.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"enforce_strict_classifier_result_count": {
				Description: "Whether to fail requests if the number of results from the classifier is not equal to the number of key-value pairs sent to it.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"ai_analysis_scope": {
				Description: "Which direction to apply AI analysis. Required. Supported values are `request` or `response`. Use preference=none to disable AI analysis entirely.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"created_at": {
				Description: "The timestamp of the Resource Classifier Preference creation.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"updated_at": {
				Description: "The timestamp of the Resource Classifier Preference update.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func resourceResourceClassifierConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceId := d.Get("resource_id").(string)
	preference := d.Get("preference").(string)
	aiAnalysisTimeout := int32(d.Get("ai_analysis_timeout_seconds").(int))
	enforceStrictClassifierResultCount := d.Get("enforce_strict_classifier_result_count").(bool)
	aiAnalysisScope := d.Get("ai_analysis_scope").(string)

	msg := &corev1.CreateResourceClassifierConfigurationRequest{
		ResourceId:                  resourceId,
		Preference:                  preference,
		AiAnalysisTimeoutSeconds:    &aiAnalysisTimeout,
		StrictClassifierResultCount: &enforceStrictClassifierResultCount,
		AiAnalysisScope:             aiAnalysisScope,
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err = v.Validate(msg); err != nil {
		return diag.FromErr(err)
	}

	response, err := c.Grpc.Sdk.ResourceServiceClient.CreateResourceClassifierConfiguration(ctx, msg)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.ResourceClassifierConfiguration.Id)
	d.Set("resource_id", response.ResourceClassifierConfiguration.ResourceId)
	d.Set("preference", response.ResourceClassifierConfiguration.Preference)
	d.Set("ai_analysis_timeout_seconds", response.ResourceClassifierConfiguration.AiAnalysisTimeoutSeconds)
	d.Set("enforce_strict_classifier_result_count", response.ResourceClassifierConfiguration.StrictClassifierResultCount)
	d.Set("ai_analysis_scope", response.ResourceClassifierConfiguration.AiAnalysisScope)
	d.Set("created_at", response.ResourceClassifierConfiguration.CreatedAt)
	d.Set("updated_at", response.ResourceClassifierConfiguration.UpdatedAt)

	return diags
}

func resourceResourceClassifierConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceId := d.Get("resource_id").(string)

	response, err := c.Grpc.Sdk.ResourceServiceClient.GetResourceClassifierConfiguration(ctx, &corev1.GetResourceClassifierConfigurationRequest{ResourceId: resourceId})
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("id", response.ResourceClassifierConfiguration.Id)
	d.Set("resource_id", response.ResourceClassifierConfiguration.ResourceId)
	d.Set("preference", response.ResourceClassifierConfiguration.Preference)
	d.Set("ai_analysis_timeout_seconds", response.ResourceClassifierConfiguration.AiAnalysisTimeoutSeconds)
	d.Set("enforce_strict_classifier_result_count", response.ResourceClassifierConfiguration.StrictClassifierResultCount)
	d.Set("ai_analysis_scope", response.ResourceClassifierConfiguration.AiAnalysisScope)
	d.Set("created_at", response.ResourceClassifierConfiguration.CreatedAt)
	d.Set("updated_at", response.ResourceClassifierConfiguration.UpdatedAt)

	return diags
}

func resourceResourceClassifierConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceClassifierPreferenceId := d.Id()
	preference := d.Get("preference").(string)
	aiAnalysisTimeout := int32(d.Get("ai_analysis_timeout_seconds").(int))
	enforceStrictClassifierResultCount := d.Get("enforce_strict_classifier_result_count").(bool)
	aiAnalysisScope := d.Get("ai_analysis_scope").(string)

	msg := &corev1.UpdateResourceClassifierConfigurationRequest{
		Id:                          resourceClassifierPreferenceId,
		Preference:                  &preference,
		AiAnalysisTimeoutSeconds:    &aiAnalysisTimeout,
		StrictClassifierResultCount: &enforceStrictClassifierResultCount,
		AiAnalysisScope:             &aiAnalysisScope,
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err = v.Validate(msg); err != nil {
		return diag.FromErr(err)
	}

	response, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResourceClassifierConfiguration(ctx, msg)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("id", response.ResourceClassifierConfiguration.Id)
	d.Set("resource_id", response.ResourceClassifierConfiguration.ResourceId)
	d.Set("preference", response.ResourceClassifierConfiguration.Preference)
	d.Set("ai_analysis_timeout_seconds", response.ResourceClassifierConfiguration.AiAnalysisTimeoutSeconds)
	d.Set("enforce_strict_classifier_result_count", response.ResourceClassifierConfiguration.StrictClassifierResultCount)
	d.Set("ai_analysis_scope", response.ResourceClassifierConfiguration.AiAnalysisScope)
	d.Set("created_at", response.ResourceClassifierConfiguration.CreatedAt)
	d.Set("updated_at", response.ResourceClassifierConfiguration.UpdatedAt)

	return diags
}

func resourceResourceClassifierConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceClassifierConfigurationId := d.Id()

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceClassifierConfiguration(ctx, &corev1.DeleteResourceClassifierConfigurationRequest{Id: resourceClassifierConfigurationId})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
