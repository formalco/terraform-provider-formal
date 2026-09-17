package resource

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"

	formal "github.com/formalco/go-sdk/v3"
	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/go-sdk/v3/core/v1/corev1connect"
	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

type groupUserLinkTestServer struct {
	corev1connect.UnimplementedGroupServiceHandler
	getUserGroupLink func(*corev1.GetUserGroupLinkRequest) (*corev1.GetUserGroupLinkResponse, error)
}

func (s groupUserLinkTestServer) GetUserGroupLink(_ context.Context, req *connect.Request[corev1.GetUserGroupLinkRequest]) (*connect.Response[corev1.GetUserGroupLinkResponse], error) {
	response, err := s.getUserGroupLink(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(response), nil
}

func TestGroupUserLinkImport(t *testing.T) {
	t.Parallel()

	data := schema.TestResourceDataRaw(t, ResourceGroupLinkUser().Schema, nil)
	data.SetId("link-1")

	imported, err := ResourceGroupLinkUser().Importer.StateContext(t.Context(), data, nil)
	require.NoError(t, err)
	require.Len(t, imported, 1)
	require.Equal(t, "link-1", imported[0].Id())
}

func TestGroupUserLinkReadFindsLinkAndReconstructsState(t *testing.T) {
	t.Parallel()

	var requestedID string
	server := groupUserLinkTestServer{
		getUserGroupLink: func(req *corev1.GetUserGroupLinkRequest) (*corev1.GetUserGroupLinkResponse, error) {
			requestedID = req.Id
			return &corev1.GetUserGroupLinkResponse{
				UserGroupLink: &corev1.UserGroupLink{
					Id:                    "link-1",
					Group:                 &corev1.Group{Id: "group-1"},
					User:                  &corev1.User{Id: "user-1"},
					TerminationProtection: true,
				},
			}, nil
		},
	}
	meta := newGroupUserLinkTestClients(t, server)
	data := schema.TestResourceDataRaw(t, ResourceGroupLinkUser().Schema, nil)
	data.SetId("link-1")

	diags := ResourceGroupLinkUser().ReadContext(t.Context(), data, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "link-1", requestedID)
	require.Equal(t, "link-1", data.Get("id"))
	require.Equal(t, "group-1", data.Get("group_id"))
	require.Equal(t, "user-1", data.Get("user_id"))
	require.Equal(t, true, data.Get("termination_protection"))
	require.Equal(t, "link-1", data.Id())
}

func TestGroupUserLinkReadClearsMissingLink(t *testing.T) {
	t.Parallel()

	server := groupUserLinkTestServer{
		getUserGroupLink: func(*corev1.GetUserGroupLinkRequest) (*corev1.GetUserGroupLinkResponse, error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("not found"))
		},
	}
	meta := newGroupUserLinkTestClients(t, server)
	data := schema.TestResourceDataRaw(t, ResourceGroupLinkUser().Schema, nil)
	data.SetId("missing-link")

	diags := ResourceGroupLinkUser().ReadContext(t.Context(), data, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Empty(t, data.Id())
}

func newGroupUserLinkTestClients(t *testing.T, groupService corev1connect.GroupServiceHandler) *clients.Clients {
	t.Helper()

	mux := http.NewServeMux()
	path, handler := corev1connect.NewGroupServiceHandler(groupService)
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	sdkClient, err := formal.New(
		formal.WithAPIKey("test-api-key"),
		formal.WithBaseURL(server.URL),
	)
	require.NoError(t, err)

	return &clients.Clients{Grpc: &api.GrpcClient{Sdk: sdkClient}}
}
