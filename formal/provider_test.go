package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	require.NoError(t, New("dev")().InternalValidate())
}

func TestProviderAuthOption(t *testing.T) {
	t.Setenv("FORMAL_API_KEY", "")
	p := New("dev")()

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
					"aws": []any{map[string]any{
						"audience": "oidc.formal.ai/integrationoidc_01h45ytscbebyvny4gc8cr8ma2",
					}},
					"env": "GITLAB_OIDC_TOKEN",
				}},
			},
			wantErr: "oidc must configure exactly one token source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, p.Schema, tt.config)
			option, err := providerAuthOption(t.Context(), d)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, option)
		})
	}
}

func TestProviderAuthOptionUsesAPIKeyEnvironmentFallback(t *testing.T) {
	t.Setenv("FORMAL_API_KEY", "formal_api_key")
	p := New("dev")()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{})

	option, err := providerAuthOption(t.Context(), d)

	require.NoError(t, err)
	require.NotNil(t, option)
}

func TestProviderAuthOptionOIDCIgnoresAPIKeyEnvironmentFallback(t *testing.T) {
	t.Setenv("FORMAL_API_KEY", "formal_api_key")
	p := New("dev")()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"oidc": []any{map[string]any{"env": "GITLAB_OIDC_TOKEN"}},
	})

	option, err := providerAuthOption(t.Context(), d)

	require.NoError(t, err)
	require.NotNil(t, option)
}

func TestEnvTokenSourceUsesEnvironmentToken(t *testing.T) {
	const name = "FORMAL_TEST_OIDC_TOKEN"
	t.Setenv(name, "token.jwt")

	source, err := (envOIDCTokenSourceConfig{name: name}).tokenSource(t.Context())
	require.NoError(t, err)
	token, err := source.Token(t.Context())
	require.NoError(t, err)
	require.Equal(t, "token.jwt", token.JWT)
}

func TestEnvTokenSourceRejectsEmptyToken(t *testing.T) {
	const name = "FORMAL_TEST_OIDC_TOKEN"
	t.Setenv(name, "")

	source, err := (envOIDCTokenSourceConfig{name: name}).tokenSource(t.Context())
	require.NoError(t, err)
	_, err = source.Token(t.Context())

	require.ErrorContains(t, err, "token must not be empty")
}

func TestValidateOIDCAudience(t *testing.T) {
	warnings, errors := validateOIDCAudience(
		"oidc.formal.ai/integrationoidc_01h45ytscbebyvny4gc8cr8ma2",
		"audience",
	)
	require.Empty(t, warnings)
	require.Empty(t, errors)

	warnings, errors = validateOIDCAudience("https://example.com", "audience")
	require.Empty(t, warnings)
	require.Len(t, errors, 1)
}

func TestParseOIDCTokenSourceConfig(t *testing.T) {
	const audience = "oidc.formal.ai/integrationoidc_01h45ytscbebyvny4gc8cr8ma2"

	tests := []struct {
		name string
		raw  []any
		want oidcTokenSourceConfig
	}{
		{
			name: "AWS",
			raw: []any{map[string]any{
				"aws": []any{map[string]any{"audience": audience}},
			}},
			want: awsOIDCTokenSourceConfig{audience: audience},
		},
		{
			name: "environment",
			raw: []any{map[string]any{
				"env": "GITLAB_OIDC_TOKEN",
			}},
			want: envOIDCTokenSourceConfig{name: "GITLAB_OIDC_TOKEN"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOIDCTokenSourceConfig(tt.raw)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
