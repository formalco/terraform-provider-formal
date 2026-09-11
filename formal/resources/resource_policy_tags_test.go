package resource

import (
	"context"
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	formal "github.com/formalco/go-sdk/v3"
	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/go-sdk/v3/core/v1/corev1connect"
	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

type policyTagsService struct {
	corev1connect.UnimplementedPoliciesServiceHandler
	mu      sync.Mutex
	policy  *corev1.Policy
	created chan *corev1.CreatePolicyRequest
	updated chan *corev1.UpdatePolicyRequest
}

func (s *policyTagsService) CreatePolicy(_ context.Context, req *connect.Request[corev1.CreatePolicyRequest]) (*connect.Response[corev1.CreatePolicyResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.created <- req.Msg
	s.policy = &corev1.Policy{
		Id: "policy_test", Name: req.Msg.Name, Description: req.Msg.Description, Code: req.Msg.Code,
		Status: req.Msg.Status, TerminationProtection: req.Msg.TerminationProtection, Tags: req.Msg.Tags,
	}
	return connect.NewResponse(&corev1.CreatePolicyResponse{Policy: proto.Clone(s.policy).(*corev1.Policy)}), nil
}

func (s *policyTagsService) GetPolicy(_ context.Context, _ *connect.Request[corev1.GetPolicyRequest]) (*connect.Response[corev1.GetPolicyResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return connect.NewResponse(&corev1.GetPolicyResponse{Policy: proto.Clone(s.policy).(*corev1.Policy)}), nil
}

func (s *policyTagsService) UpdatePolicy(_ context.Context, req *connect.Request[corev1.UpdatePolicyRequest]) (*connect.Response[corev1.UpdatePolicyResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updated <- req.Msg
	if req.Msg.Tags != nil {
		s.policy.Tags = req.Msg.Tags.Values
	}
	return connect.NewResponse(&corev1.UpdatePolicyResponse{Policy: proto.Clone(s.policy).(*corev1.Policy)}), nil
}

func TestResourcePolicyTagLimit(t *testing.T) {
	tagsSchema := ResourcePolicy().Schema["tags"]
	for _, size := range []int{0, 500, 501} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			tags := make(map[string]any, size)
			for i := range size {
				tags[fmt.Sprint(i)] = "value"
			}
			_, errors := tagsSchema.ValidateFunc(tags, "tags")
			if size > 500 {
				require.NotEmpty(t, errors)
			} else {
				require.Empty(t, errors)
			}
		})
	}
}

func TestResourcePolicyTagsRoundTrip(t *testing.T) {
	service := &policyTagsService{
		created: make(chan *corev1.CreatePolicyRequest, 1),
		updated: make(chan *corev1.UpdatePolicyRequest, 2),
	}
	_, handler := corev1connect.NewPoliciesServiceHandler(service)
	server := httptest.NewServer(handler)
	defer server.Close()
	sdk, err := formal.New(formal.WithBaseURL(server.URL), formal.WithAPIKey("test"))
	require.NoError(t, err)
	meta := &clients.Clients{Grpc: &api.GrpcClient{Sdk: sdk}}
	resource := ResourcePolicy()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"name": "test", "description": "description", "module": "code", "status": "active",
		"tags": map[string]any{"env": "prod"},
	})
	require.Empty(t, resourcePolicyCreate(t.Context(), data, meta))
	require.Equal(t, map[string]string{"env": "prod"}, (<-service.created).Tags)
	require.Equal(t, map[string]any{"env": "prod"}, data.Get("tags"))

	changed, err := schema.InternalMap(resource.Schema).Data(data.State(), &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"tags.env": {Old: "prod", New: "dev"},
		},
	})
	require.NoError(t, err)
	require.True(t, changed.HasChange("tags"))
	require.False(t, changed.HasChange("module"))
	require.False(t, changed.HasChange("name"))
	require.Empty(t, resourcePolicyUpdate(t.Context(), changed, meta))
	update := <-service.updated
	require.NotNil(t, update.Tags)
	require.Equal(t, map[string]string{"env": "dev"}, update.Tags.Values)
	require.Equal(t, map[string]any{"env": "dev"}, changed.Get("tags"))

	service.mu.Lock()
	service.policy.Tags = map[string]string{"env": "drifted"}
	service.mu.Unlock()
	require.Empty(t, resourcePolicyRead(t.Context(), changed, meta))
	require.Equal(t, map[string]any{"env": "drifted"}, changed.Get("tags"))

	cleared, err := schema.InternalMap(resource.Schema).Data(changed.State(), &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"tags.%":   {Old: "1", New: "0"},
			"tags.env": {Old: "drifted", NewRemoved: true},
		},
	})
	require.NoError(t, err)
	require.True(t, cleared.HasChange("tags"))
	require.Empty(t, resourcePolicyUpdate(t.Context(), cleared, meta))
	update = <-service.updated
	require.NotNil(t, update.Tags)
	require.Empty(t, update.Tags.Values)
	require.Empty(t, cleared.Get("tags"))
}
