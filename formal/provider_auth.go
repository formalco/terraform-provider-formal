package provider

import (
	"context"
	"errors"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	formal "github.com/formalco/go-sdk/v3"
)

func providerAuthOption(ctx context.Context, d *schema.ResourceData) (formal.Option, error) {
	apiKey := d.Get("api_key").(string)
	oidcConfig, hasOIDC := d.GetOk("oidc")

	if hasOIDC {
		if apiKey != "" {
			return nil, errors.New("api_key and oidc are mutually exclusive")
		}
		source, err := newOIDCTokenSource(ctx, oidcConfig.([]any))
		if err != nil {
			return nil, err
		}
		return formal.WithOIDCTokenSource(source), nil
	}

	if apiKey == "" {
		apiKey = os.Getenv("FORMAL_API_KEY")
	}
	if apiKey == "" {
		return nil, errors.New("authentication required: configure api_key, FORMAL_API_KEY, or oidc")
	}
	return formal.WithAPIKey(apiKey), nil
}
