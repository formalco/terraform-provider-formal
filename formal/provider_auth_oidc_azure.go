package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
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
	credential, err := newAzureOIDCCredential()
	if err != nil {
		return nil, fmt.Errorf("load Azure credential for OIDC: %w", err)
	}
	return oidcazure.NewTokenSource(credential, c.integrationID)
}

func newAzureOIDCCredential() (azcore.TokenCredential, error) {
	if os.Getenv("AZURE_FEDERATED_TOKEN_FILE") != "" {
		return azidentity.NewWorkloadIdentityCredential(nil)
	}

	options := &azidentity.ManagedIdentityCredentialOptions{}
	if clientID := os.Getenv("AZURE_CLIENT_ID"); clientID != "" {
		options.ID = azidentity.ClientID(clientID)
	}
	return azidentity.NewManagedIdentityCredential(options)
}
