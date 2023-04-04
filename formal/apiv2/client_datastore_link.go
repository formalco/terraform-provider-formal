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

func (c GrpcClient) CreateLink(ctx context.Context, link SidecarDatastoreLink) (string, error) {

	req := connect.NewRequest(&adminv1.CreateSidecarDatastoreLinkRequest{
		DatastoreId: link.DatastoreId,
		SidecarId:   link.SidecarId,
		Port:        int32(link.Port),
	})
	datastoreLink, err := c.SidecarServiceClient.CreateSidecarDatastoreLink(ctx, req)
	if err != nil {
		return "", err
	}
	return datastoreLink.Msg.LinkId, nil
}

func (c GrpcClient) GetLink(ctx context.Context, linkId string) (*SidecarDatastoreLink, error) {

	req := connect.NewRequest(&adminv1.GetLinkByIdRequest{
		Id: linkId,
	})
	datastoreLink, err := c.SidecarServiceClient.GetLinkById(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &SidecarDatastoreLink{
		Id:          datastoreLink.Msg.Link.Id,
		SidecarId:   datastoreLink.Msg.Link.DatastoreId,
		DatastoreId: datastoreLink.Msg.Link.SidecarId,
		Port:        int(datastoreLink.Msg.Link.Port),
	}
	return res, nil
}

func (c GrpcClient) DeleteLink(ctx context.Context, linkId string) error {

	req := connect.NewRequest(&adminv1.RemoveSidecarDatastoreLinkRequest{
		LinkId: linkId,
	})
	_, err := c.SidecarServiceClient.RemoveSidecarDatastoreLink(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
