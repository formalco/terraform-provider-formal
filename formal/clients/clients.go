package clients

import (
	"github.com/formalco/terraform-provider-formal/formal/api"
)

type Clients struct {
	Grpc   *api.GrpcClient
	ApiKey string
}
