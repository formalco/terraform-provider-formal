package apiv2

import (
	"net/http"

	core_connect "buf.build/gen/go/formal/core/connectrpc/go/core/v1/corev1connect"
)

type transport struct {
	underlyingTransport http.RoundTripper
	apiKey              string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Api-Key", t.apiKey)
	return t.underlyingTransport.RoundTrip(req)
}

func NewApiClientV2(apiKey string) core_connect.ResourceServiceClient {
	httpClient := &http.Client{Transport: &transport{
		underlyingTransport: http.DefaultTransport,
		apiKey:              apiKey,
	}}

	return core_connect.NewResourceServiceClient(httpClient, "https://v2api.formalcloud.net")
}
