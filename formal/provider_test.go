package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	require.NoError(t, New("dev")().InternalValidate())
}

func TestProviderOIDCSchema(t *testing.T) {
	const integrationID = "integrationoidc_01h45ytscbebyvny4gc8cr8ma2"
	p := New("dev")()

	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
	}{
		{
			name: "environment source with integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"integration_id": integrationID,
					"env":            "SPACELIFT_OIDC_TOKEN",
				}},
			},
		},
		{
			name: "AWS source with integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"integration_id": integrationID,
					"aws":            []any{map[string]any{}},
				}},
			},
		},
		{
			name: "Azure source with integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"integration_id": integrationID,
					"azure":          []any{map[string]any{}},
				}},
			},
		},
		{
			name: "AWS source without integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"aws": []any{map[string]any{}},
				}},
			},
			wantErr: true,
		},
		{
			name: "Azure source without integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"azure": []any{map[string]any{}},
				}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := p.Validate(terraform.NewResourceConfigRaw(tt.config))
			require.Equal(t, tt.wantErr, diags.HasError())
		})
	}
}

func TestProviderAuthOption(t *testing.T) {
	t.Setenv("FORMAL_API_KEY", "")
	p := New("dev")()
	const integrationID = "integrationoidc_01h45ytscbebyvny4gc8cr8ma2"

	tests := []struct {
		name    string
		config  map[string]any
		wantErr string
	}{
		{
			name:   "explicit API key",
			config: map[string]any{"api_key": "formal_api_key"},
		},
		{
			name: "environment token source",
			config: map[string]any{
				"oidc": []any{map[string]any{"env": "GITLAB_OIDC_TOKEN"}},
			},
		},
		{
			name: "environment token source with integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"integration_id": integrationID,
					"env":            "SPACELIFT_OIDC_TOKEN",
				}},
			},
		},
		{
			name: "Azure token source with integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"integration_id": integrationID,
					"azure":          []any{map[string]any{}},
				}},
			},
		},
		{
			name: "AWS token source without integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"aws": []any{map[string]any{}},
				}},
			},
			wantErr: "oidc.integration_id is required with aws",
		},
		{
			name: "Azure token source without integration ID",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"azure": []any{map[string]any{}},
				}},
			},
			wantErr: "oidc.integration_id is required with azure",
		},
		{
			name:    "authentication missing",
			config:  map[string]any{},
			wantErr: "authentication required",
		},
		{
			name: "API key and OIDC",
			config: map[string]any{
				"api_key": "formal_api_key",
				"oidc":    []any{map[string]any{"env": "GITLAB_OIDC_TOKEN"}},
			},
			wantErr: "api_key and oidc are mutually exclusive",
		},
		{
			name: "OIDC source missing",
			config: map[string]any{
				"oidc": []any{map[string]any{}},
			},
			wantErr: "oidc must configure exactly one token source",
		},
		{
			name: "multiple OIDC sources",
			config: map[string]any{
				"oidc": []any{map[string]any{
					"integration_id": integrationID,
					"aws":            []any{map[string]any{}},
					"azure":          []any{map[string]any{}},
					"env":            "GITLAB_OIDC_TOKEN",
				}},
			},
			wantErr: "oidc must configure exactly one token source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, p.Schema, tt.config)
			options, err := providerAuthOptions(t.Context(), d)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, options)
		})
	}
}

func TestProviderAuthOptionUsesAPIKeyEnvironmentFallback(t *testing.T) {
	t.Setenv("FORMAL_API_KEY", "formal_api_key")
	p := New("dev")()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{})

	options, err := providerAuthOptions(t.Context(), d)

	require.NoError(t, err)
	require.Len(t, options, 1)
}

func TestProviderAuthOptionOIDCIgnoresAPIKeyEnvironmentFallback(t *testing.T) {
	t.Setenv("FORMAL_API_KEY", "formal_api_key")
	p := New("dev")()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"oidc": []any{map[string]any{"env": "GITLAB_OIDC_TOKEN"}},
	})

	options, err := providerAuthOptions(t.Context(), d)

	require.NoError(t, err)
	require.Len(t, options, 1)
}

func TestProviderAuthOptionsEnvIntegrationIDAddsIntegrationHeader(t *testing.T) {
	t.Setenv("FORMAL_API_KEY", "")
	p := New("dev")()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"oidc": []any{map[string]any{
			"integration_id": "integrationoidc_01h45ytscbebyvny4gc8cr8ma2",
			"env":            "SPACELIFT_OIDC_TOKEN",
		}},
	})

	options, err := providerAuthOptions(t.Context(), d)

	require.NoError(t, err)
	require.Len(t, options, 1)
}

func TestEnvTokenSourceUsesEnvironmentToken(t *testing.T) {
	const name = "FORMAL_TEST_OIDC_TOKEN"
	const integrationID = "integrationoidc_01h45ytscbebyvny4gc8cr8ma2"
	t.Setenv(name, "token.jwt")

	source, err := (envOIDCTokenSourceConfig{
		name:          name,
		integrationID: integrationID,
	}).tokenSource(t.Context())
	require.NoError(t, err)
	token, err := source.Token(t.Context())
	require.NoError(t, err)
	require.Equal(t, "token.jwt", token.JWT)
	headerIntegrationID, ok := token.HeaderIntegrationID.Get()
	require.True(t, ok)
	require.Equal(t, integrationID, headerIntegrationID)
}

func TestEnvTokenSourceWithoutIntegrationIDOmitsHeader(t *testing.T) {
	const name = "FORMAL_TEST_OIDC_TOKEN"
	t.Setenv(name, "token.jwt")

	source, err := (envOIDCTokenSourceConfig{name: name}).tokenSource(t.Context())
	require.NoError(t, err)
	token, err := source.Token(t.Context())
	require.NoError(t, err)
	_, ok := token.HeaderIntegrationID.Get()
	require.False(t, ok)
}

func TestEnvTokenSourceRejectsEmptyToken(t *testing.T) {
	const name = "FORMAL_TEST_OIDC_TOKEN"
	t.Setenv(name, "")

	source, err := (envOIDCTokenSourceConfig{name: name}).tokenSource(t.Context())
	require.NoError(t, err)
	_, err = source.Token(t.Context())

	require.ErrorContains(t, err, "token must not be empty")
}

func TestValidateOIDCIntegrationID(t *testing.T) {
	warnings, errors := validateOIDCIntegrationID(
		"integrationoidc_01h45ytscbebyvny4gc8cr8ma2",
		"integration_id",
	)
	require.Empty(t, warnings)
	require.Empty(t, errors)

	warnings, errors = validateOIDCIntegrationID("https://example.com", "integration_id")
	require.Empty(t, warnings)
	require.Len(t, errors, 1)
}

func TestParseOIDCAuthConfig(t *testing.T) {
	const integrationID = "integrationoidc_01h45ytscbebyvny4gc8cr8ma2"
	const audience = "oidc.formal.ai/integrationoidc_01h45ytscbebyvny4gc8cr8ma2"

	tests := []struct {
		name string
		raw  []any
		want oidcAuthConfig
	}{
		{
			name: "AWS",
			raw: []any{map[string]any{
				"integration_id": integrationID,
				"aws":            []any{map[string]any{}},
			}},
			want: oidcAuthConfig{
				tokenSourceConfig: awsOIDCTokenSourceConfig{audience: audience},
			},
		},
		{
			name: "environment",
			raw: []any{map[string]any{
				"env": "GITLAB_OIDC_TOKEN",
			}},
			want: oidcAuthConfig{
				tokenSourceConfig: envOIDCTokenSourceConfig{name: "GITLAB_OIDC_TOKEN"},
			},
		},
		{
			name: "Azure",
			raw: []any{map[string]any{
				"integration_id": integrationID,
				"azure":          []any{map[string]any{}},
			}},
			want: oidcAuthConfig{
				tokenSourceConfig: azureOIDCTokenSourceConfig{integrationID: integrationID},
			},
		},
		{
			name: "environment with integration ID",
			raw: []any{map[string]any{
				"integration_id": integrationID,
				"env":            "SPACELIFT_OIDC_TOKEN",
			}},
			want: oidcAuthConfig{
				tokenSourceConfig: envOIDCTokenSourceConfig{
					name:          "SPACELIFT_OIDC_TOKEN",
					integrationID: integrationID,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOIDCAuthConfig(tt.raw)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
