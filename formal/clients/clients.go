package clients

import (
	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/formalco/terraform-provider-formal/formal/apiv2"
)

type Clients struct {
	Http *api.Client
	Grpc *apiv2.GrpcClient
}
