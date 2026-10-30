package resource

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	formal "github.com/formalco/go-sdk/v3"
	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/go-sdk/v3/core/v1/corev1connect"
	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

type workflowSecretTestServer struct {
	corev1connect.UnimplementedWorkflowServiceHandler

	mu       sync.Mutex
	secrets  []*corev1.WorkflowSecret
	values   map[string]string
	pageSize int
	updates  []*corev1.UpdateWorkflowSecretRequest
	deleted  []string
}

func (s *workflowSecretTestServer) CreateWorkflowSecret(_ context.Context, req *connect.Request[corev1.CreateWorkflowSecretRequest]) (*connect.Response[corev1.CreateWorkflowSecretResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.secrets {
		if existing.WorkflowId == req.Msg.WorkflowId && existing.Name == req.Msg.Name {
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("workflow secret already exists"))
		}
	}
	secret := &corev1.WorkflowSecret{
		Id:         "workflowsecret_" + req.Msg.Name,
		Name:       req.Msg.Name,
		WorkflowId: req.Msg.WorkflowId,
		CreatedAt:  timestamppb.Now(),
		UpdatedAt:  timestamppb.Now(),
	}
	s.secrets = append(s.secrets, secret)
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[secret.Id] = req.Msg.Value
	return connect.NewResponse(&corev1.CreateWorkflowSecretResponse{WorkflowSecret: secret}), nil
}

func (s *workflowSecretTestServer) ListWorkflowSecrets(_ context.Context, req *connect.Request[corev1.ListWorkflowSecretsRequest]) (*connect.Response[corev1.ListWorkflowSecretsResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var matches []*corev1.WorkflowSecret
	for _, secret := range s.secrets {
		if secret.WorkflowId == req.Msg.WorkflowId {
			matches = append(matches, secret)
		}
	}
	start := 0
	for i, secret := range matches {
		if secret.Id == req.Msg.Cursor {
			start = i + 1
		}
	}
	pageSize := s.pageSize
	if pageSize == 0 {
		pageSize = int(req.Msg.Limit)
	}
	end := min(start+pageSize, len(matches))
	res := &corev1.ListWorkflowSecretsResponse{WorkflowSecrets: matches[start:end], ListMetadata: &corev1.ListMetadata{}}
	if end < len(matches) {
		res.ListMetadata.NextCursor = matches[end-1].Id
	}
	return connect.NewResponse(res), nil
}

func (s *workflowSecretTestServer) UpdateWorkflowSecret(_ context.Context, req *connect.Request[corev1.UpdateWorkflowSecretRequest]) (*connect.Response[corev1.UpdateWorkflowSecretResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, secret := range s.secrets {
		if secret.Id == req.Msg.Id {
			s.updates = append(s.updates, req.Msg)
			s.values[secret.Id] = req.Msg.Value
			secret.UpdatedAt = timestamppb.Now()
			return connect.NewResponse(&corev1.UpdateWorkflowSecretResponse{WorkflowSecret: secret}), nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("workflow secret not found"))
}

func (s *workflowSecretTestServer) DeleteWorkflowSecret(_ context.Context, req *connect.Request[corev1.DeleteWorkflowSecretRequest]) (*connect.Response[corev1.DeleteWorkflowSecretResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deleted = append(s.deleted, req.Msg.Id)
	for i, secret := range s.secrets {
		if secret.Id == req.Msg.Id {
			s.secrets = append(s.secrets[:i], s.secrets[i+1:]...)
			return connect.NewResponse(&corev1.DeleteWorkflowSecretResponse{Id: req.Msg.Id}), nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("workflow secret not found"))
}

func newWorkflowSecretTestClients(t *testing.T, server *workflowSecretTestServer) *clients.Clients {
	t.Helper()

	mux := http.NewServeMux()
	path, handler := corev1connect.NewWorkflowServiceHandler(server)
	mux.Handle(path, handler)
	httpServer := httptest.NewServer(mux)
	t.Cleanup(httpServer.Close)

	sdkClient, err := formal.New(formal.WithAPIKey("test-api-key"), formal.WithBaseURL(httpServer.URL))
	require.NoError(t, err)
	return &clients.Clients{Grpc: &api.GrpcClient{Sdk: sdkClient}}
}

// applyWorkflowSecret plans next against prior (nil to create), with raw as the
// configuration Terraform keeps for write-only attributes. An empty plan is a no-op.
func applyWorkflowSecret(t *testing.T, meta *clients.Clients, prior *terraform.InstanceState, next map[string]any, raw cty.Value) (*terraform.InstanceState, diag.Diagnostics) {
	t.Helper()
	r := ResourceWorkflowSecret()
	if prior != nil && !raw.IsNull() {
		prior.RawConfig = raw
	}
	diff, err := r.Diff(t.Context(), prior, terraform.NewResourceConfigRaw(next), meta)
	require.NoError(t, err)
	if diff == nil {
		return prior, nil
	}
	if !raw.IsNull() {
		diff.RawConfig = raw
	}
	return r.Apply(t.Context(), prior, diff, meta)
}

func TestWorkflowSecretSchema(t *testing.T) {
	t.Parallel()
	r := ResourceWorkflowSecret()
	require.NoError(t, r.InternalValidate(r.Schema, true))
	require.True(t, r.Schema["value"].Sensitive)
	require.True(t, r.Schema["value_wo"].WriteOnly)
	for _, key := range []string{"workflow_id", "name"} {
		require.True(t, r.Schema[key].ForceNew, key)
	}
	for _, key := range []string{"value", "value_wo", "value_wo_version"} {
		require.False(t, r.Schema[key].ForceNew, key)
	}
}

func TestWorkflowSecretNameValidation(t *testing.T) {
	t.Parallel()
	validate := ResourceWorkflowSecret().Schema["name"].ValidateFunc
	for _, name := range []string{"slack_webhook_url", "_token", "API_KEY2"} {
		_, errs := validate(name, "name")
		require.Empty(t, errs, name)
	}
	for _, name := range []string{"", "1token", "slack-webhook", "secrets.token"} {
		_, errs := validate(name, "name")
		require.NotEmpty(t, errs, name)
	}
}

func TestWorkflowSecretCreateAndUpdateValue(t *testing.T) {
	t.Parallel()

	server := &workflowSecretTestServer{}
	meta := newWorkflowSecretTestClients(t, server)
	config := map[string]any{"workflow_id": "workflow_1", "name": "token", "value": "v1"}

	state, diags := applyWorkflowSecret(t, meta, nil, config, cty.NullVal(cty.EmptyObject))
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "workflowsecret_token", state.ID)
	require.Equal(t, "workflow_1", state.Attributes["workflow_id"])
	require.Equal(t, "v1", state.Attributes["value"])
	require.NotEmpty(t, state.Attributes["created_at"])
	require.Equal(t, "v1", server.values["workflowsecret_token"])

	state, diags = applyWorkflowSecret(t, meta, state, config, cty.NullVal(cty.EmptyObject))
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Empty(t, server.updates, "an unchanged value is not resent")

	rotated := maps.Clone(config)
	rotated["value"] = "v2"
	state, diags = applyWorkflowSecret(t, meta, state, rotated, cty.NullVal(cty.EmptyObject))
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, server.updates, 1)
	require.Equal(t, "workflowsecret_token", server.updates[0].Id)
	require.Equal(t, "v2", server.values["workflowsecret_token"])
	require.Equal(t, "workflowsecret_token", state.ID, "changing the value updates in place")
}

func TestWorkflowSecretWriteOnlyValue(t *testing.T) {
	t.Parallel()

	server := &workflowSecretTestServer{}
	meta := newWorkflowSecretTestClients(t, server)
	config := map[string]any{"workflow_id": "workflow_1", "name": "token", "value_wo_version": 1}

	state, diags := applyWorkflowSecret(t, meta, nil, config, cty.ObjectVal(map[string]cty.Value{"value_wo": cty.StringVal("s3cret")}))
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "s3cret", server.values["workflowsecret_token"])
	require.NotContains(t, state.Attributes, "value_wo")
	require.Empty(t, state.Attributes["value"])

	rotatedRaw := cty.ObjectVal(map[string]cty.Value{"value_wo": cty.StringVal("rotated")})
	state, diags = applyWorkflowSecret(t, meta, state, config, rotatedRaw)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Empty(t, server.updates, "a new write-only value is only sent when value_wo_version changes")

	bumped := maps.Clone(config)
	bumped["value_wo_version"] = 2
	_, diags = applyWorkflowSecret(t, meta, state, bumped, rotatedRaw)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, server.updates, 1)
	require.Equal(t, "rotated", server.values["workflowsecret_token"])
}

func TestWorkflowSecretReadFindsSecretAcrossPages(t *testing.T) {
	t.Parallel()

	server := &workflowSecretTestServer{
		pageSize: 1,
		secrets: []*corev1.WorkflowSecret{
			{Id: "workflowsecret_1", WorkflowId: "workflow_1", Name: "other", CreatedAt: timestamppb.Now(), UpdatedAt: timestamppb.Now()},
			{Id: "workflowsecret_2", WorkflowId: "workflow_2", Name: "token", CreatedAt: timestamppb.Now(), UpdatedAt: timestamppb.Now()},
			{Id: "workflowsecret_3", WorkflowId: "workflow_1", Name: "token", CreatedAt: timestamppb.Now(), UpdatedAt: timestamppb.Now()},
		},
	}
	meta := newWorkflowSecretTestClients(t, server)
	r := ResourceWorkflowSecret()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"workflow_id": "workflow_1", "value": "kept"})
	d.SetId("workflowsecret_3")

	diags := r.ReadContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "workflowsecret_3", d.Id())
	require.Equal(t, "token", d.Get("name"))
	require.Equal(t, "kept", d.Get("value"), "the API never returns the value, so state keeps it")
}

func TestWorkflowSecretReadClearsMissingSecret(t *testing.T) {
	t.Parallel()

	meta := newWorkflowSecretTestClients(t, &workflowSecretTestServer{})
	r := ResourceWorkflowSecret()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"workflow_id": "workflow_1"})
	d.SetId("workflowsecret_gone")

	diags := r.ReadContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Empty(t, d.Id())
}

func TestWorkflowSecretImport(t *testing.T) {
	t.Parallel()

	server := &workflowSecretTestServer{
		secrets: []*corev1.WorkflowSecret{
			{Id: "workflowsecret_1", WorkflowId: "workflow_1", Name: "token", CreatedAt: timestamppb.Now(), UpdatedAt: timestamppb.Now()},
		},
	}
	meta := newWorkflowSecretTestClients(t, server)
	r := ResourceWorkflowSecret()

	d := schema.TestResourceDataRaw(t, r.Schema, nil)
	d.SetId("workflow_1/workflowsecret_1")
	imported, err := r.Importer.StateContext(t.Context(), d, meta)
	require.NoError(t, err)
	require.Len(t, imported, 1)
	require.Equal(t, "workflowsecret_1", imported[0].Id())
	require.Equal(t, "workflow_1", imported[0].Get("workflow_id"))
	require.Equal(t, "token", imported[0].Get("name"))

	for _, id := range []string{"workflowsecret_1", "workflow_1/", "/workflowsecret_1", "workflow_2/workflowsecret_1"} {
		d := schema.TestResourceDataRaw(t, r.Schema, nil)
		d.SetId(id)
		_, err := r.Importer.StateContext(t.Context(), d, meta)
		require.Error(t, err, id)
	}
}

func TestWorkflowSecretDeleteIgnoresMissingSecret(t *testing.T) {
	t.Parallel()

	server := &workflowSecretTestServer{}
	meta := newWorkflowSecretTestClients(t, server)
	r := ResourceWorkflowSecret()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"workflow_id": "workflow_1"})
	d.SetId("workflowsecret_gone")

	diags := r.DeleteContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, []string{"workflowsecret_gone"}, server.deleted)
	require.Empty(t, d.Id())
}
