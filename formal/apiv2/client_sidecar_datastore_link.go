package apiv2

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"context"
	"github.com/bufbuild/connect-go"
)

type SidecarDatastoreLink struct {
	Id          string `json:"id"`
	SidecarId   string `json:"sidecar_id"`
	DatastoreId string `json:"datastore_id"`
	Port        int    `json:"port"`
}

func (c GrpcClient) CreateSidecarDatastoreLink(ctx context.Context, link SidecarDatastoreLink) (string, error) {

	req := connect.NewRequest(&adminv1.CreateSidecarDatastoreLinkRequest{
		DatastoreId: link.DatastoreId,
		SidecarId:   link.SidecarId,
		Port:        int32(link.Port),
	})
	sidecarDatastoreLink, err := c.SidecarServiceClient.CreateSidecarDatastoreLink(ctx, req)
	if err != nil {
		return "", err
	}
	return sidecarDatastoreLink.Msg.LinkId, nil
}

func (c GrpcClient) GetSidecarDatastoreLink(ctx context.Context, linkId string) (*SidecarDatastoreLink, error) {

	req := connect.NewRequest(&adminv1.GetLinkByIdRequest{
		Id: linkId,
	})
	sidecarDatastoreLink, err := c.SidecarServiceClient.GetLinkById(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &SidecarDatastoreLink{
		Id:          sidecarDatastoreLink.Msg.Link.Id,
		SidecarId:   sidecarDatastoreLink.Msg.Link.SidecarId,
		DatastoreId: sidecarDatastoreLink.Msg.Link.DatastoreId,
		Port:        int(sidecarDatastoreLink.Msg.Link.Port),
	}
	return res, nil
}

func (c GrpcClient) DeleteSidecarDatastoreLink(ctx context.Context, linkId string) error {

	req := connect.NewRequest(&adminv1.RemoveSidecarDatastoreLinkRequest{
		LinkId: linkId,
	})
	_, err := c.SidecarServiceClient.RemoveSidecarDatastoreLink(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
