package provider

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/samber/mo"

	"github.com/formalco/go-sdk/v3/oidc"
	oidcaws "github.com/formalco/go-sdk/v3/oidc/aws"
)

type awsOIDCTokenSourceConfig struct {
	audience string
}

func parseAWSOIDCTokenSource(settings map[string]any) mo.Option[oidcTokenSourceConfig] {
	awsSettings, _ := settings["aws"].([]any)
	if len(awsSettings) != 1 {
		return mo.None[oidcTokenSourceConfig]()
	}
	config, ok := awsSettings[0].(map[string]any)
	if !ok {
		return mo.None[oidcTokenSourceConfig]()
	}
	audience, _ := config["audience"].(string)
	return mo.Some[oidcTokenSourceConfig](awsOIDCTokenSourceConfig{audience: audience})
}

func (c awsOIDCTokenSourceConfig) tokenSource(ctx context.Context) (oidc.TokenSource, error) {
	if err := oidc.ValidateAudience(c.audience); err != nil {
		return nil, err
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration for OIDC: %w", err)
	}
	return oidcaws.NewTokenSource(sts.NewFromConfig(cfg), c.audience)
}
