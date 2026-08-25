package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/formalco/go-sdk/v3/oidc"
)

type oidcTokenSourceConfig interface {
	tokenSource(context.Context) (oidc.TokenSource, error)
}

type oidcTokenSourceParser func(map[string]any) mo.Option[oidcTokenSourceConfig]

var oidcTokenSourceParsers = []oidcTokenSourceParser{
	parseAWSOIDCTokenSource,
	parseEnvOIDCTokenSource,
}

func newOIDCTokenSource(ctx context.Context, raw []any) (oidc.TokenSource, error) {
	config, err := parseOIDCTokenSourceConfig(raw)
	if err != nil {
		return nil, err
	}
	return config.tokenSource(ctx)
}

func parseOIDCTokenSourceConfig(raw []any) (oidcTokenSourceConfig, error) {
	if len(raw) != 1 {
		return nil, errors.New("oidc must contain exactly one configuration block")
	}

	settings, ok := raw[0].(map[string]any)
	if !ok {
		return nil, errors.New("oidc must configure exactly one token source")
	}

	configured := lo.FilterMap(
		oidcTokenSourceParsers,
		func(parse oidcTokenSourceParser, _ int) (oidcTokenSourceConfig, bool) {
			return parse(settings).Get()
		},
	)

	if len(configured) != 1 {
		return nil, errors.New("oidc must configure exactly one token source")
	}
	return configured[0], nil
}

func validateOIDCAudience(value any, key string) ([]string, []error) {
	if err := oidc.ValidateAudience(value.(string)); err != nil {
		return nil, []error{fmt.Errorf("%s: %w", key, err)}
	}
	return nil, nil
}
