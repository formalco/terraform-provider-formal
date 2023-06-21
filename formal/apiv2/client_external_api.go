package apiv2

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"github.com/bufbuild/connect-go"
)

type IntegrationExternalAPIAuthBasic struct {
	Username string
	Password string
}

type IntegrationExternalAPIAuth struct {
	Type  string
	Basic IntegrationExternalAPIAuthBasic
}

type IntegrationExternalAPI struct {
	ID      string
	Type    string
	Name    string
	Url     string
	Auth    IntegrationExternalAPIAuth
	Keyword string
}

func (c GrpcClient) CreateExternalAPIIntegration(ctx context.Context, integration IntegrationExternalAPI) (string, error) {

	req := connect.NewRequest(&adminv1.CreateExternalAPIIntegrationRequest{
		Type: integration.Type,
		Name: integration.Name,
		Url:  integration.Url,
		Auth: &adminv1.Auth{
			Type: integration.Auth.Type,
			Basic: &adminv1.Auth_Basic{
				Username: integration.Auth.Basic.Username,
				Password: integration.Auth.Basic.Password,
			},
		},
	})
	externalApi, err := c.IntegrationExternalAPIService.CreateExternalAPIIntegration(ctx, req)
	if err != nil {
		return "", err
	}
	return externalApi.Msg.Id, nil
}

func (c GrpcClient) GetExternalAPIIntegration(ctx context.Context, id string) (*IntegrationExternalAPI, error) {

	req := connect.NewRequest(&adminv1.GetExternalAPIIntegrationRequest{
		Id: id,
	})
	externalApi, err := c.IntegrationExternalAPIService.GetExternalAPIIntegration(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &IntegrationExternalAPI{
		ID:   externalApi.Msg.Integration.Id,
		Type: externalApi.Msg.Integration.Type,
		Name: externalApi.Msg.Integration.Name,
		Url:  externalApi.Msg.Integration.Url,
		Auth: IntegrationExternalAPIAuth{
			Type: externalApi.Msg.Integration.AuthType,
		},
		Keyword: externalApi.Msg.Integration.Keyword,
	}
	return res, nil
}

func (c GrpcClient) DeleteExternalAPIIntegration(ctx context.Context, id string) error {

	req := connect.NewRequest(&adminv1.DeleteExternalAPIIntegrationRequest{
		Id: id,
	})
	_, err := c.IntegrationExternalAPIService.DeleteExternalAPIIntegration(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
