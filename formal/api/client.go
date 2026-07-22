package api

import (
	"errors"
	"os"

	formal "github.com/formalco/go-sdk/v3"
)

type GrpcClient struct {
	ReturnSensitiveValue bool
	Sdk                  *formal.Client
}

func NewClient(apiKey string, returnSensitiveValue bool) (*GrpcClient, error) {
	opts := []formal.Option{formal.WithAPIKey(apiKey)}
	if os.Getenv("FORMAL_ENV") == "dev" {
		url := os.Getenv("FORMAL_DEV_URL")
		if url == "" {
			return nil, errors.New("FORMAL_ENV=dev requires FORMAL_DEV_URL")
		}
		opts = append(opts, formal.WithBaseURL(url))
	}
	client, err := formal.New(opts...)
	if err != nil {
		return nil, err
	}
	return &GrpcClient{Sdk: client, ReturnSensitiveValue: returnSensitiveValue}, nil
}
