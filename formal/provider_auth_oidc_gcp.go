package provider

import (
	"context"

	"github.com/samber/mo"

	"github.com/formalco/go-sdk/v3/oidc"
	oidcgcp "github.com/formalco/go-sdk/v3/oidc/gcp"
)

type gcpOIDCTokenSourceConfig struct {
	audience string
}

func parseGCPOIDCTokenSource(settings map[string]any) mo.Option[oidcTokenSourceConfig] {
	gcpSettings, _ := settings["gcp"].([]any)
	if len(gcpSettings) != 1 {
		return mo.None[oidcTokenSourceConfig]()
	}
	integrationID, _ := settings["integration_id"].(string)
	return mo.Some[oidcTokenSourceConfig](gcpOIDCTokenSourceConfig{
		audience: oidc.AudiencePrefix + integrationID,
	})
}

func (c gcpOIDCTokenSourceConfig) tokenSource(context.Context) (oidc.TokenSource, error) {
	return oidcgcp.NewDefaultTokenSource(c.audience)
}
