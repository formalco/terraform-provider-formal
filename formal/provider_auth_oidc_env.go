package provider

import (
	"context"
	"os"

	"github.com/samber/mo"

	"github.com/formalco/go-sdk/v3/oidc"
)

type envOIDCTokenSourceConfig struct {
	name          string
	integrationID string
}

func parseEnvOIDCTokenSource(settings map[string]any) mo.Option[oidcTokenSourceConfig] {
	name, _ := settings["env"].(string)
	if name == "" {
		return mo.None[oidcTokenSourceConfig]()
	}
	integrationID, _ := settings["integration_id"].(string)
	return mo.Some[oidcTokenSourceConfig](envOIDCTokenSourceConfig{
		name:          name,
		integrationID: integrationID,
	})
}

func (c envOIDCTokenSourceConfig) tokenSource(context.Context) (oidc.TokenSource, error) {
	headerIntegrationID := mo.None[string]()
	if c.integrationID != "" {
		headerIntegrationID = mo.Some(c.integrationID)
	}
	return headerIntegrationIDTokenSource{
		TokenSource:         oidc.Static(os.Getenv(c.name)),
		headerIntegrationID: headerIntegrationID,
	}, nil
}

type headerIntegrationIDTokenSource struct {
	oidc.TokenSource
	headerIntegrationID mo.Option[string]
}

func (s headerIntegrationIDTokenSource) Token(ctx context.Context) (oidc.Token, error) {
	token, err := s.TokenSource.Token(ctx)
	if err != nil {
		return oidc.Token{}, err
	}
	token.HeaderIntegrationID = s.headerIntegrationID
	return token, nil
}
