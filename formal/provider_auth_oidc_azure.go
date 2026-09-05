package provider

import (
	"context"

	"github.com/samber/mo"

	"github.com/formalco/go-sdk/v3/oidc"
	oidcazure "github.com/formalco/go-sdk/v3/oidc/azure"
)

type azureOIDCTokenSourceConfig struct {
	integrationID string
}

func parseAzureOIDCTokenSource(settings map[string]any) mo.Option[oidcTokenSourceConfig] {
	azureSettings, _ := settings["azure"].([]any)
	if len(azureSettings) != 1 {
		return mo.None[oidcTokenSourceConfig]()
	}
	integrationID, _ := settings["integration_id"].(string)
	return mo.Some[oidcTokenSourceConfig](azureOIDCTokenSourceConfig{
		integrationID: integrationID,
	})
}

func (c azureOIDCTokenSourceConfig) tokenSource(context.Context) (oidc.TokenSource, error) {
	return oidcazure.NewDefaultTokenSource(c.integrationID)
}
