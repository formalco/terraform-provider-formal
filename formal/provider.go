package provider

import (
	"context"
	"os"

	"github.com/formalco/terraform-provider-formal/formal/clients"

	"github.com/formalco/terraform-provider-formal/formal/api"
	resource "github.com/formalco/terraform-provider-formal/formal/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_key": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"retrieve_sensitive_values": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
			},
			DataSourcesMap: map[string]*schema.Resource{},
			ResourcesMap: map[string]*schema.Resource{
				"formal_policy":                        resource.ResourcePolicy(),
				"formal_group":                         resource.ResourceGroup(),
				"formal_group_link_user":               resource.ResourceGroupLinkUser(),
				"formal_resource":                      resource.ResourceResource(),
				"formal_sidecar_resource_link":         resource.ResourceSidecarResourceLink(),
				"formal_sidecar":                       resource.ResourceSidecar(),
				"formal_native_user":                   resource.ResourceNativeUser(),
				"formal_native_user_link":              resource.ResourceNativeUserLink(),
				"formal_user":                          resource.ResourceUser(),
				"formal_integration_log":               resource.ResourceIntegrationLogs(),
				"formal_integration_mfa":               resource.ResourceIntegrationMfa(),
				"formal_integration_bi":                resource.ResourceIntegrationBI(),
				"formal_integration_data_catalog":      resource.ResourceIntegrationDataCatalog(),
				"formal_integration_cloud":             resource.ResourceIntegrationCloud(),
				"formal_satellite":                     resource.ResourceSatellite(),
				"formal_data_domain":                   resource.ResourceDataDomain(),
				"formal_tracker":                       resource.ResourceTracker(),
				"formal_data_discovery":                resource.ResourceDataDiscovery(),
				"formal_resource_health_check":         resource.ResourceHealthCheck(),
				"formal_policies_external_data_loader": resource.ResourcePoliciesExternalDataLoader(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		apiKey := d.Get("api_key").(string)
		if apiKey == "" {
			apiKey = os.Getenv("FORMAL_API_KEY")
			if apiKey == "" {
				return nil, diag.Errorf("api_key must be set in the provider or as an environment variable")
			}
		}
		returnSensitiveValue := d.Get("retrieve_sensitive_values").(bool)

		grpc := api.NewClient(apiKey, returnSensitiveValue)

		return &clients.Clients{Grpc: grpc}, nil
	}
}
