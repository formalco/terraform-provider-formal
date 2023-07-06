package apiv2

import (
	"github.com/formalco/go-sdk/sdk"
)

type GrpcClient struct {
	ReturnSensitiveValue bool
	Sdk                  *sdk.FormalSDK
}

func NewClient(apiKey string, returnSensitiveValue bool) *GrpcClient {
	return &GrpcClient{Sdk: sdk.NewWithUrl(apiKey, "http://localhost:443"), ReturnSensitiveValue: returnSensitiveValue}
}
