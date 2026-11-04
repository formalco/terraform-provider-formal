package api

import (
	"errors"
	"net/http"
	"os"

	formal "github.com/formalco/go-sdk/v3"
)

type GrpcClient struct {
	ReturnSensitiveValue bool
	Sdk                  *formal.Client
}

func NewClient(opts []formal.Option, userAgent string, returnSensitiveValue bool) (*GrpcClient, error) {
	if os.Getenv("FORMAL_ENV") == "dev" {
		url := os.Getenv("FORMAL_DEV_URL")
		if url == "" {
			return nil, errors.New("FORMAL_ENV=dev requires FORMAL_DEV_URL")
		}
		opts = append(opts, formal.WithBaseURL(url))
	}
	if userAgent != "" {
		opts = append(opts, formal.WithHTTPClient(&http.Client{
			Transport: userAgentTransport{userAgent: userAgent, base: http.DefaultTransport},
		}))
	}
	client, err := formal.New(opts...)
	if err != nil {
		return nil, err
	}
	return &GrpcClient{Sdk: client, ReturnSensitiveValue: returnSensitiveValue}, nil
}

// userAgentTransport overrides the connect-go default User-Agent, which is
// identical for every Go SDK caller.
type userAgentTransport struct {
	userAgent string
	base      http.RoundTripper
}

func (t userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", t.userAgent)
	return t.base.RoundTrip(req)
}
