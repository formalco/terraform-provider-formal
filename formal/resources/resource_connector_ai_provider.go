package resource

import (
	"context"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

var providerBlocks = []string{"formal_ai_satellite", "gemini", "google_vertex_ai", "anthropic", "aws_bedrock", "openai", "azure_ai"}

func writeOnlyApiKeySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"api_key": {
			Description: "The API key. This value is not stored in Terraform state.",
			Type:        schema.TypeString,
			Required:    true,
			WriteOnly:   true,
		},
		"api_key_version": {
			Description: "Version trigger for `api_key`. Increment this value to update the key.",
			Type:        schema.TypeInt,
			Required:    true,
		},
	}
}

func ResourceConnectorAiProvider() *schema.Resource {
	geminiSchema := writeOnlyApiKeySchema()
	anthropicSchema := writeOnlyApiKeySchema()
	openaiSchema := writeOnlyApiKeySchema()

	azureSchema := writeOnlyApiKeySchema()
	azureSchema["endpoint"] = &schema.Schema{
		Description: "The Azure AI Foundry endpoint URL.",
		Type:        schema.TypeString,
		Required:    true,
	}

	return &schema.Resource{
		Description:   "Configures the AI provider for a connector's session analyzer.",
		CreateContext: resourceConnectorAiProviderCreate,
		ReadContext:   resourceConnectorAiProviderRead,
		UpdateContext: resourceConnectorAiProviderUpdate,
		DeleteContext: resourceConnectorAiProviderDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this connector AI provider.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector_id": {
				Description: "The ID of the connector this AI provider is linked to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"formal_ai_satellite": {
				Description:  "Use the Formal AI satellite as the provider.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: providerBlocks,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"gemini": {
				Description:  "Use Google Gemini as the provider.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: providerBlocks,
				Elem:         &schema.Resource{Schema: geminiSchema},
			},
			"google_vertex_ai": {
				Description:  "Use Google Vertex AI as the provider.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: providerBlocks,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gcp_project_id": {
							Description: "The GCP project ID.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"region": {
							Description: "The GCP region.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"anthropic": {
				Description:  "Use Anthropic as the provider.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: providerBlocks,
				Elem:         &schema.Resource{Schema: anthropicSchema},
			},
			"aws_bedrock": {
				Description:  "Use AWS Bedrock as the provider.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: providerBlocks,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Description: "The AWS region.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"openai": {
				Description:  "Use OpenAI as the provider.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: providerBlocks,
				Elem:         &schema.Resource{Schema: openaiSchema},
			},
			"azure_ai": {
				Description:  "Use Azure AI Foundry as the provider.",
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: providerBlocks,
				Elem:         &schema.Resource{Schema: azureSchema},
			},
		},
	}
}

// getWriteOnlyApiKey reads the write-only api_key from raw config for a provider block.
func getWriteOnlyApiKey(d *schema.ResourceData, providerBlock string) (string, diag.Diagnostics) {
	val, rawDiags := d.GetRawConfigAt(cty.GetAttrPath(providerBlock).IndexInt(0).GetAttr("api_key"))
	if rawDiags.HasError() {
		return "", diag.Errorf("failed to get %s.0.api_key: %v", providerBlock, rawDiags)
	}
	if !val.IsNull() && val.Type() == cty.String {
		return val.AsString(), nil
	}
	return "", diag.Errorf("%s.0.api_key must be specified", providerBlock)
}

func buildAiProviderConfig(d *schema.ResourceData) (*corev1.ConnectorAiProviderConfig, diag.Diagnostics) {
	if v, ok := d.GetOk("formal_ai_satellite"); ok {
		configs := v.([]interface{})
		if len(configs) > 0 {
			return &corev1.ConnectorAiProviderConfig{
				Provider: &corev1.ConnectorAiProviderConfig_FormalAiSatellite{
					FormalAiSatellite: &corev1.FormalAiSatelliteConfig{},
				},
			}, nil
		}
	}
	if v, ok := d.GetOk("gemini"); ok {
		configs := v.([]interface{})
		if len(configs) > 0 {
			apiKey, diags := getWriteOnlyApiKey(d, "gemini")
			if diags.HasError() {
				return nil, diags
			}
			return &corev1.ConnectorAiProviderConfig{
				Provider: &corev1.ConnectorAiProviderConfig_Gemini{
					Gemini: &corev1.GeminiConfig{ApiKey: apiKey},
				},
			}, nil
		}
	}
	if v, ok := d.GetOk("google_vertex_ai"); ok {
		configs := v.([]interface{})
		if len(configs) > 0 {
			cfg := configs[0].(map[string]interface{})
			return &corev1.ConnectorAiProviderConfig{
				Provider: &corev1.ConnectorAiProviderConfig_GoogleVertexAi{
					GoogleVertexAi: &corev1.GoogleVertexAiConfig{
						GcpProjectId: cfg["gcp_project_id"].(string),
						Region:       cfg["region"].(string),
					},
				},
			}, nil
		}
	}
	if v, ok := d.GetOk("anthropic"); ok {
		configs := v.([]interface{})
		if len(configs) > 0 {
			apiKey, diags := getWriteOnlyApiKey(d, "anthropic")
			if diags.HasError() {
				return nil, diags
			}
			return &corev1.ConnectorAiProviderConfig{
				Provider: &corev1.ConnectorAiProviderConfig_Anthropic{
					Anthropic: &corev1.AnthropicConfig{ApiKey: apiKey},
				},
			}, nil
		}
	}
	if v, ok := d.GetOk("aws_bedrock"); ok {
		configs := v.([]interface{})
		if len(configs) > 0 {
			cfg := configs[0].(map[string]interface{})
			return &corev1.ConnectorAiProviderConfig{
				Provider: &corev1.ConnectorAiProviderConfig_AwsBedrock{
					AwsBedrock: &corev1.AwsBedrockConfig{Region: cfg["region"].(string)},
				},
			}, nil
		}
	}
	if v, ok := d.GetOk("openai"); ok {
		configs := v.([]interface{})
		if len(configs) > 0 {
			apiKey, diags := getWriteOnlyApiKey(d, "openai")
			if diags.HasError() {
				return nil, diags
			}
			return &corev1.ConnectorAiProviderConfig{
				Provider: &corev1.ConnectorAiProviderConfig_Openai{
					Openai: &corev1.OpenAiConfig{ApiKey: apiKey},
				},
			}, nil
		}
	}
	if v, ok := d.GetOk("azure_ai"); ok {
		configs := v.([]interface{})
		if len(configs) > 0 {
			cfg := configs[0].(map[string]interface{})
			apiKey, diags := getWriteOnlyApiKey(d, "azure_ai")
			if diags.HasError() {
				return nil, diags
			}
			return &corev1.ConnectorAiProviderConfig{
				Provider: &corev1.ConnectorAiProviderConfig_AzureAi{
					AzureAi: &corev1.AzureAiConfig{
						ApiKey:   apiKey,
						Endpoint: cfg["endpoint"].(string),
					},
				},
			}, nil
		}
	}
	return nil, diag.Errorf("exactly one provider configuration block must be set")
}

func setAiProviderState(d *schema.ResourceData, provider *corev1.ConnectorAiProvider) {
	d.Set("id", provider.Id)
	d.Set("connector_id", provider.ConnectorId)

	switch p := provider.Config.Provider.(type) {
	case *corev1.ConnectorAiProviderConfig_FormalAiSatellite:
		d.Set("formal_ai_satellite", []interface{}{map[string]interface{}{}})
	case *corev1.ConnectorAiProviderConfig_Gemini:
		// api_key is write-only, nothing to read back
	case *corev1.ConnectorAiProviderConfig_GoogleVertexAi:
		d.Set("google_vertex_ai", []interface{}{map[string]interface{}{
			"gcp_project_id": p.GoogleVertexAi.GcpProjectId,
			"region":         p.GoogleVertexAi.Region,
		}})
	case *corev1.ConnectorAiProviderConfig_Anthropic:
		// api_key is write-only, nothing to read back
	case *corev1.ConnectorAiProviderConfig_AwsBedrock:
		d.Set("aws_bedrock", []interface{}{map[string]interface{}{
			"region": p.AwsBedrock.Region,
		}})
	case *corev1.ConnectorAiProviderConfig_Openai:
		// api_key is write-only, nothing to read back
	case *corev1.ConnectorAiProviderConfig_AzureAi:
		d.Set("azure_ai", []interface{}{map[string]interface{}{
			"endpoint":        p.AzureAi.Endpoint,
			"api_key_version": d.Get("azure_ai.0.api_key_version"),
		}})
	}
}

func resourceConnectorAiProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	config, diags := buildAiProviderConfig(d)
	if diags.HasError() {
		return diags
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorAiProvider(ctx, connect.NewRequest(&corev1.CreateConnectorAiProviderRequest{
		ConnectorId: d.Get("connector_id").(string),
		Config:      config,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ConnectorAiProvider.Id)
	setAiProviderState(d, res.Msg.ConnectorAiProvider)
	return nil
}

func resourceConnectorAiProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorAiProvider(ctx, connect.NewRequest(&corev1.GetConnectorAiProviderRequest{
		Id: &corev1.GetConnectorAiProviderRequest_ProviderId{
			ProviderId: d.Id(),
		},
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The connector AI provider was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	setAiProviderState(d, res.Msg.ConnectorAiProvider)
	d.SetId(res.Msg.ConnectorAiProvider.Id)
	return diags
}

func resourceConnectorAiProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	config, diags := buildAiProviderConfig(d)
	if diags.HasError() {
		return diags
	}

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnectorAiProvider(ctx, connect.NewRequest(&corev1.UpdateConnectorAiProviderRequest{
		Id:     d.Id(),
		Config: config,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceConnectorAiProviderRead(ctx, d, meta)
}

func resourceConnectorAiProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorAiProvider(ctx, connect.NewRequest(&corev1.DeleteConnectorAiProviderRequest{
		Id: d.Id(),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
