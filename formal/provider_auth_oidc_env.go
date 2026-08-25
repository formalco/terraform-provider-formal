package provider

import (
	"context"
	"os"

	"github.com/samber/mo"

	"github.com/formalco/go-sdk/v3/oidc"
)

type envOIDCTokenSourceConfig struct {
	name string
}

func parseEnvOIDCTokenSource(settings map[string]any) mo.Option[oidcTokenSourceConfig] {
	name, _ := settings["env"].(string)
	if name == "" {
		return mo.None[oidcTokenSourceConfig]()
	}
	return mo.Some[oidcTokenSourceConfig](envOIDCTokenSourceConfig{name: name})
}

func (c envOIDCTokenSourceConfig) tokenSource(context.Context) (oidc.TokenSource, error) {
	return oidc.Static(os.Getenv(c.name)), nil
}
