package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func TestResourceIntegrationOIDCSchema(t *testing.T) {
	t.Parallel()
	r := ResourceIntegrationOIDC()
	require.NoError(t, r.InternalValidate(r.Schema, true))

	require.True(t, r.Schema["name"].Required)
	require.True(t, r.Schema["issuer"].Required)
	require.True(t, r.Schema["machine_user_id"].Required)
	require.Equal(t, "true", r.Schema["claim_condition"].Default)
	require.Equal(t, "active", r.Schema["status"].Default)
	require.True(t, r.Schema["audience"].Computed)
	require.False(t, r.Schema["issuer"].Sensitive)
	require.False(t, r.Schema["jwks_uri"].Sensitive)
}

func TestValidateOIDCIssuerURL(t *testing.T) {
	t.Parallel()
	_, errs := validateOIDCIssuerURL("https://token.actions.githubusercontent.com", "issuer")
	require.Empty(t, errs)

	_, errs = validateOIDCIssuerURL("https://issuer.example.com/path", "issuer")
	require.Empty(t, errs)

	_, errs = validateOIDCIssuerURL("https://issuer.example.com?x=1", "issuer")
	require.NotEmpty(t, errs)

	_, errs = validateOIDCIssuerURL("ftp://issuer.example.com", "issuer")
	require.NotEmpty(t, errs)
}

func TestOIDCAudience(t *testing.T) {
	t.Parallel()
	require.Equal(t, "oidc.formal.ai/integrationoidc_01abc", oidcAudience("integrationoidc_01abc"))
}

func TestFlattenIntegrationOIDC(t *testing.T) {
	t.Parallel()
	r := ResourceIntegrationOIDC()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	d.SetId("integrationoidc_01h45ytscbebyvny4gc8cr8ma2")

	jwks := "https://issuer.example.com/keys"
	diags := flattenIntegrationOIDC(d, &corev1.IntegrationOIDC{
		Id:             "integrationoidc_01h45ytscbebyvny4gc8cr8ma2",
		Name:           "github-actions",
		Issuer:         "https://token.actions.githubusercontent.com",
		JwksUri:        &jwks,
		MachineUserId:  "user_machine",
		ClaimCondition: `claims.repository == "formalco/monorepo"`,
		Status:         "active",
		CreatedAt:      timestamppb.Now(),
		UpdatedAt:      timestamppb.Now(),
	})
	require.Empty(t, diags)
	require.Equal(t, "oidc.formal.ai/integrationoidc_01h45ytscbebyvny4gc8cr8ma2", d.Get("audience"))
	require.Equal(t, "github-actions", d.Get("name"))
	require.Equal(t, jwks, d.Get("jwks_uri"))
}

func TestFlattenIntegrationOIDCClearsJWKSWhenAbsent(t *testing.T) {
	t.Parallel()
	r := ResourceIntegrationOIDC()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"jwks_uri": "https://old.example.com/keys",
	})
	d.SetId("integrationoidc_01h45ytscbebyvny4gc8cr8ma2")

	diags := flattenIntegrationOIDC(d, &corev1.IntegrationOIDC{
		Id:             "integrationoidc_01h45ytscbebyvny4gc8cr8ma2",
		Name:           "example-oidc",
		Issuer:         "https://issuer.example.com",
		MachineUserId:  "user_machine",
		ClaimCondition: "true",
		Status:         "draft",
	})
	require.Empty(t, diags)
	require.Empty(t, d.Get("jwks_uri"))
}
