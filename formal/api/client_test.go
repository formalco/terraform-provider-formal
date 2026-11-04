package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	formal "github.com/formalco/go-sdk/v3"
	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func TestNewClientSendsUserAgent(t *testing.T) {
	userAgents := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgents <- r.UserAgent()
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(
		[]formal.Option{formal.WithAPIKey("test"), formal.WithBaseURL(server.URL)},
		"Terraform/1.9.0 (+https://www.terraform.io) terraform-provider-formal/4.0.0",
		false,
	)
	require.NoError(t, err)

	_, err = client.Sdk.UserServiceClient.GetUser(t.Context(), &corev1.GetUserRequest{Id: "user_123"})
	require.Error(t, err)
	require.Equal(t, "Terraform/1.9.0 (+https://www.terraform.io) terraform-provider-formal/4.0.0", <-userAgents)
}
