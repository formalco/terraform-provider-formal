package apiv2

import (
	"buf.build/gen/go/formal/admin/bufbuild/connect-go/admin/v1/adminv1connect"
	"net/http"
	"time"
)

const FORMAL_HOST_URL string = "https://adminv2.api.formalcloud.net"

type GrpcClient struct {
	SidecarServiceClient          adminv1connect.SidecarServiceClient
	IntegrationExternalAPIService adminv1connect.IntegrationExternalAPIServiceClient
}

func NewClient(apiKey string) *GrpcClient {
	sidecarServiceClient := adminv1connect.NewSidecarServiceClient(
		&http.Client{Timeout: 100 * time.Second, Transport: &transport{underlyingTransport: http.DefaultTransport, apiKey: apiKey}},
		FORMAL_HOST_URL)
	integrationExternalApiClient := adminv1connect.NewIntegrationExternalAPIServiceClient(
		&http.Client{Timeout: 100 * time.Second, Transport: &transport{underlyingTransport: http.DefaultTransport, apiKey: apiKey}},
		FORMAL_HOST_URL)
	return &GrpcClient{
		SidecarServiceClient:          sidecarServiceClient,
		IntegrationExternalAPIService: integrationExternalApiClient,
	}
}

type transport struct {
	underlyingTransport http.RoundTripper
	apiKey              string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Api-Key", t.apiKey)
	return t.underlyingTransport.RoundTrip(req)
}
