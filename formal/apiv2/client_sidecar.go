package apiv2

import (
	"context"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"github.com/bufbuild/connect-go"
)

type Sidecar struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Technology        string `json:"technology"`
	DataplaneId       string `json:"dataplane_id"`
	DsId              string `json:"datastore_id"`
	Deployed          bool   `json:"deployed"`
	DeploymentType    string `json:"deployment_type"`
	FailOpen          bool   `json:"fail_open"`
	FormalHostname    string `json:"formal_hostname"`
	FullKMSDecryption bool   `json:"global_kms_decrypt"`
	NetworkType       string `json:"network_type"`
	Version           string `json:"version"`
	CreatedAt         int64  `json:"created_at"`
}

func (c GrpcClient) CreateSidecar(ctx context.Context, sidecar Sidecar) (string, error) {
	r := &adminv1.CreateSidecarRequest{
		Name:             sidecar.Name,
		Technology:       sidecar.Technology,
		DeploymentType:   sidecar.DeploymentType,
		DatastoreId:      sidecar.DsId,
		DataplaneId:      sidecar.DataplaneId,
		FailOpen:         sidecar.FailOpen,
		GlobalKmsDecrypt: sidecar.FullKMSDecryption,
		FormalHostname:   sidecar.FormalHostname,
		Version:          sidecar.Version,
		NetworkType:      sidecar.NetworkType,
	}

	req := connect.NewRequest(r)

	res, err := c.SidecarServiceClient.CreateSidecar(ctx, req)
	if err != nil {
		return "", err
	}
	return res.Msg.Id, nil
}

func (c GrpcClient) GetSidecar(ctx context.Context, sidecarID string) (*Sidecar, error) {
	req := connect.NewRequest(&adminv1.GetSidecarByIdRequest{Id: sidecarID})
	res, err := c.SidecarServiceClient.GetSidecarById(ctx, req)
	if err != nil {
		return nil, err
	}
	return &Sidecar{
		Id:                res.Msg.Sidecar.Id,
		Name:              res.Msg.Sidecar.Name,
		Technology:        res.Msg.Sidecar.Technology,
		DeploymentType:    res.Msg.Sidecar.DeploymentType,
		DsId:              res.Msg.Sidecar.DatastoreId,
		DataplaneId:       res.Msg.Sidecar.DataplaneId,
		FailOpen:          res.Msg.Sidecar.FailOpen,
		FullKMSDecryption: res.Msg.Sidecar.GlobalKmsDecrypt,
		FormalHostname:    res.Msg.Sidecar.FormalHostname,
		Version:           res.Msg.Sidecar.Version,
		NetworkType:       res.Msg.Sidecar.NetworkType,
		Deployed:          res.Msg.Sidecar.Deployed,
	}, nil
}

func (c GrpcClient) DeleteSidecar(ctx context.Context, sidecarID string) error {
	req := connect.NewRequest(&adminv1.DeleteSidecarRequest{Id: sidecarID})
	_, err := c.SidecarServiceClient.DeleteSidecar(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
