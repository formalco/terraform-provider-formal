package api

import (
	"os"

	"github.com/formalco/go-sdk/sdk"
)

type GrpcClient struct {
	ReturnSensitiveValue bool
	Sdk                  *sdk.FormalSDK
}

func NewClient(apiKey string, returnSensitiveValue bool) *GrpcClient {
	if os.Getenv("FORMAL_ENV") == "dev" {
		url := os.Getenv("FORMAL_DEV_URL")
		return &GrpcClient{Sdk: sdk.NewWithUrl(apiKey, url), ReturnSensitiveValue: returnSensitiveValue}
	}
	return &GrpcClient{Sdk: sdk.New(apiKey), ReturnSensitiveValue: returnSensitiveValue}
}
