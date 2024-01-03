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

	// Customize the content ofz descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }

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
				"formal_policy":                   resource.ResourcePolicy(),
				"formal_group":                    resource.ResourceGroup(),
				"formal_group_link_role":          resource.ResourceGroupLinkRole(),
				"formal_datastore":                resource.ResourceDatastore(),
				"formal_sidecar_datastore_link":   resource.ResourceSidecarDatastoreLink(),
				"formal_sidecar":                  resource.ResourceSidecar(),
				"formal_key":                      resource.ResourceKey(),
				"formal_field_encryption":         resource.ResourceFieldEncryption(),
				"formal_default_field_encryption": resource.ResourceDefaultFieldEncryption(),
				"formal_native_role":              resource.ResourceNativeRole(),
				"formal_native_role_link":         resource.ResourceNativeRoleLink(),
				"formal_user":                     resource.ResourceUser(),
				"formal_integration_log":          resource.ResourceIntegrationLogs(),
				"formal_integration_app":          resource.ResourceIntegrationApp(),
				"formal_integration_external_api": resource.ResourceIntegrationExternalApi(),
				"formal_integration_datahub":      resource.ResourceIntegrationDatahub(),
				"formal_satellite":                resource.ResourceSatellite(),
				"formal_encryption_key":           resource.ResourceEncryptionKey(),
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
