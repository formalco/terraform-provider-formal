package api

import (
	"os"

	"github.com/formalco/go-sdk/sdk"
	v2 "github.com/formalco/go-sdk/sdk/v2"
)

type GrpcClient struct {
	ReturnSensitiveValue bool
	Sdk                  *sdk.FormalSDK
	SdkV2                *v2.FormalSDK
}

func NewClient(apiKey string, returnSensitiveValue bool) *GrpcClient {
	if os.Getenv("FORMAL_ENV") == "dev" {
		url := os.Getenv("FORMAL_DEV_URL")
		return &GrpcClient{Sdk: sdk.NewWithUrl(apiKey, url), SdkV2: v2.NewWithUrl(apiKey, url), ReturnSensitiveValue: returnSensitiveValue}
	}
	return &GrpcClient{Sdk: sdk.New(apiKey), ReturnSensitiveValue: returnSensitiveValue}
}
