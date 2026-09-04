package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/samber/mo"

	formal "github.com/formalco/go-sdk/v3"
	"github.com/formalco/go-sdk/v3/oidc"
)

type oidcTokenSourceConfig interface {
	tokenSource(context.Context) (oidc.TokenSource, error)
}

type oidcTokenSourceParser func(map[string]any) mo.Option[oidcTokenSourceConfig]

type oidcAuthConfig struct {
	tokenSourceConfig oidcTokenSourceConfig
}

var oidcTokenSourceParsers = []oidcTokenSourceParser{
	parseAWSOIDCTokenSource,
	parseAzureOIDCTokenSource,
	parseEnvOIDCTokenSource,
}

func newOIDCAuthOptions(ctx context.Context, raw []any) ([]formal.Option, error) {
	config, err := parseOIDCAuthConfig(raw)
	if err != nil {
		return nil, err
	}
	source, err := config.tokenSourceConfig.tokenSource(ctx)
	if err != nil {
		return nil, err
	}

	return []formal.Option{formal.WithOIDCTokenSource(source)}, nil
}

func parseOIDCAuthConfig(raw []any) (oidcAuthConfig, error) {
	if len(raw) != 1 {
		return oidcAuthConfig{}, errors.New("oidc must contain exactly one configuration block")
	}

	settings, ok := raw[0].(map[string]any)
	if !ok {
		return oidcAuthConfig{}, errors.New("oidc must configure exactly one token source")
	}

	configured := lo.FilterMap(
		oidcTokenSourceParsers,
		func(parse oidcTokenSourceParser, _ int) (oidcTokenSourceConfig, bool) {
			return parse(settings).Get()
		},
	)

	if len(configured) != 1 {
		return oidcAuthConfig{}, errors.New("oidc must configure exactly one token source")
	}

	integrationID, _ := settings["integration_id"].(string)
	if integrationID != "" {
		if err := oidc.ValidateAudience(oidc.AudiencePrefix + integrationID); err != nil {
			return oidcAuthConfig{}, err
		}
	}

	if integrationID == "" {
		switch configured[0].(type) {
		case awsOIDCTokenSourceConfig:
			return oidcAuthConfig{}, errors.New("oidc.integration_id is required with aws")
		case azureOIDCTokenSourceConfig:
			return oidcAuthConfig{}, errors.New("oidc.integration_id is required with azure")
		}
	}
	return oidcAuthConfig{
		tokenSourceConfig: configured[0],
	}, nil
}

func validateOIDCIntegrationID(value any, key string) ([]string, []error) {
	if err := oidc.ValidateAudience(oidc.AudiencePrefix + value.(string)); err != nil {
		return nil, []error{fmt.Errorf("%s: %w", key, err)}
	}
	return nil, nil
}
