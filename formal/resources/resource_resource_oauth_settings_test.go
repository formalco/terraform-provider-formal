package resource

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	formal "github.com/formalco/go-sdk/v3"
	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/go-sdk/v3/core/v1/corev1connect"
	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

type resourceOAuthSettingsTestServer struct {
	corev1connect.UnimplementedResourceServiceHandler
	create func(*corev1.CreateResourceOAuthSettingsRequest) (*corev1.CreateResourceOAuthSettingsResponse, error)
	get    func(*corev1.GetResourceOAuthSettingsRequest) (*corev1.GetResourceOAuthSettingsResponse, error)
	update func(*corev1.UpdateResourceOAuthSettingsRequest) (*corev1.UpdateResourceOAuthSettingsResponse, error)
	delete func(*corev1.DeleteResourceOAuthSettingsRequest) (*corev1.DeleteResourceOAuthSettingsResponse, error)
}

func (s resourceOAuthSettingsTestServer) UpdateResourceOAuthSettings(_ context.Context, req *connect.Request[corev1.UpdateResourceOAuthSettingsRequest]) (*connect.Response[corev1.UpdateResourceOAuthSettingsResponse], error) {
	res, err := s.update(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (s resourceOAuthSettingsTestServer) CreateResourceOAuthSettings(_ context.Context, req *connect.Request[corev1.CreateResourceOAuthSettingsRequest]) (*connect.Response[corev1.CreateResourceOAuthSettingsResponse], error) {
	res, err := s.create(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (s resourceOAuthSettingsTestServer) GetResourceOAuthSettings(_ context.Context, req *connect.Request[corev1.GetResourceOAuthSettingsRequest]) (*connect.Response[corev1.GetResourceOAuthSettingsResponse], error) {
	res, err := s.get(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (s resourceOAuthSettingsTestServer) DeleteResourceOAuthSettings(_ context.Context, req *connect.Request[corev1.DeleteResourceOAuthSettingsRequest]) (*connect.Response[corev1.DeleteResourceOAuthSettingsResponse], error) {
	res, err := s.delete(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func newResourceOAuthSettingsTestClients(t *testing.T, server corev1connect.ResourceServiceHandler) *clients.Clients {
	t.Helper()

	mux := http.NewServeMux()
	path, handler := corev1connect.NewResourceServiceHandler(server)
	mux.Handle(path, handler)
	httpServer := httptest.NewServer(mux)
	t.Cleanup(httpServer.Close)

	sdkClient, err := formal.New(formal.WithAPIKey("test-api-key"), formal.WithBaseURL(httpServer.URL))
	require.NoError(t, err)
	return &clients.Clients{Grpc: &api.GrpcClient{Sdk: sdkClient}}
}

func TestResourceOAuthSettingsSchema(t *testing.T) {
	t.Parallel()
	r := ResourceResourceOAuthSettings()
	require.NoError(t, r.InternalValidate(r.Schema, true))
	require.NotNil(t, r.UpdateContext)
	require.Equal(t, upstreamOAuthModeFormal, r.Schema["upstream_oauth_mode"].Default)
	require.True(t, r.Schema["client_secret"].Sensitive)
	require.True(t, r.Schema["client_secret_wo"].WriteOnly)
	require.Equal(t, []string{"client_secret_wo"}, r.Schema["client_secret"].ConflictsWith)
	require.Equal(t, []string{"client_secret"}, r.Schema["client_secret_wo"].ConflictsWith)
	require.Equal(t, []string{"client_secret_wo_version"}, r.Schema["client_secret_wo"].RequiredWith)
	require.Equal(t, []string{"client_secret_wo"}, r.Schema["client_secret_wo_version"].RequiredWith)
	for _, key := range []string{"resource_id", "upstream_oauth_mode", "client_id", "issuer"} {
		require.True(t, r.Schema[key].ForceNew, key)
	}
	for _, key := range []string{"requested_scopes", "client_secret", "client_secret_wo", "client_secret_wo_version"} {
		require.False(t, r.Schema[key].ForceNew, key)
	}
}

func TestValidateResourceOAuthSettingsArgs(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		args    resourceOAuthSettingsArgs
		wantErr string
	}{
		"formal with scopes and issuer": {
			args: resourceOAuthSettingsArgs{mode: upstreamOAuthModeFormal, hasScopes: true, hasIssuer: true},
		},
		"formal with client_id": {
			args:    resourceOAuthSettingsArgs{mode: upstreamOAuthModeFormal, hasClientID: true},
			wantErr: "only allowed when upstream_oauth_mode is pre_registered",
		},
		"pre_registered public client": {
			args: resourceOAuthSettingsArgs{mode: upstreamOAuthModePreRegistered, hasClientID: true},
		},
		"pre_registered confidential client": {
			args: resourceOAuthSettingsArgs{mode: upstreamOAuthModePreRegistered, hasClientID: true, hasClientSecret: true, hasScopes: true, hasIssuer: true},
		},
		"pre_registered without client_id": {
			args:    resourceOAuthSettingsArgs{mode: upstreamOAuthModePreRegistered, hasClientSecret: true},
			wantErr: "client_id is required",
		},
		"none": {
			args: resourceOAuthSettingsArgs{mode: upstreamOAuthModeNone},
		},
		"none with scopes": {
			args:    resourceOAuthSettingsArgs{mode: upstreamOAuthModeNone, hasScopes: true},
			wantErr: "requested_scopes must be empty",
		},
		"none with issuer": {
			args:    resourceOAuthSettingsArgs{mode: upstreamOAuthModeNone, hasIssuer: true},
			wantErr: "issuer must be unset",
		},
		"none with client_secret": {
			args:    resourceOAuthSettingsArgs{mode: upstreamOAuthModeNone, hasClientSecret: true},
			wantErr: "only allowed when upstream_oauth_mode is pre_registered",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := validateResourceOAuthSettingsArgs(tt.args)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestResourceOAuthSettingsPlanRejectsInvalidCombination(t *testing.T) {
	t.Parallel()
	r := ResourceResourceOAuthSettings()

	_, err := r.Diff(t.Context(), nil, terraform.NewResourceConfigRaw(map[string]any{
		"resource_id":         "resource_mcp",
		"upstream_oauth_mode": "none",
		"requested_scopes":    []any{"repo"},
	}), nil)
	require.ErrorContains(t, err, "requested_scopes must be empty when upstream_oauth_mode is none")

	_, err = r.Diff(t.Context(), nil, terraform.NewResourceConfigRaw(map[string]any{
		"resource_id":         "resource_mcp",
		"upstream_oauth_mode": "pre_registered",
	}), nil)
	require.ErrorContains(t, err, "client_id is required")

	_, err = r.Diff(t.Context(), &terraform.InstanceState{
		RawConfig: cty.ObjectVal(map[string]cty.Value{"client_secret_wo": cty.StringVal("s3cret")}),
	}, terraform.NewResourceConfigRaw(map[string]any{
		"resource_id":              "resource_mcp",
		"upstream_oauth_mode":      upstreamOAuthModeFormal,
		"client_secret_wo_version": 1,
	}), nil)
	require.ErrorContains(t, err, "only allowed when upstream_oauth_mode is pre_registered")
}

func TestResourceOAuthSettingsPlanSkipsUnknownValues(t *testing.T) {
	t.Parallel()
	r := ResourceResourceOAuthSettings()

	block := r.CoreConfigSchema()
	attrs := map[string]cty.Value{}
	for name, attrType := range block.ImpliedType().AttributeTypes() {
		attrs[name] = cty.NullVal(attrType)
	}
	attrs["resource_id"] = cty.StringVal("resource_mcp")
	attrs["upstream_oauth_mode"] = cty.UnknownVal(cty.String)
	attrs["client_id"] = cty.StringVal("Iv1.abc")

	_, err := r.Diff(t.Context(), nil, terraform.NewResourceConfigShimmed(cty.ObjectVal(attrs), block), nil)
	require.NoError(t, err)
}

// applyResourceOAuthSettings plans next against the state prior would have
// been created with, then applies the plan.
func applyResourceOAuthSettings(t *testing.T, meta *clients.Clients, prior, next map[string]any) diag.Diagnostics {
	t.Helper()
	return applyResourceOAuthSettingsRaw(t, meta, prior, next, cty.NullVal(cty.EmptyObject))
}

// applyResourceOAuthSettingsRaw plans next against prior, with raw as the
// configuration Terraform keeps for write-only attributes.
func applyResourceOAuthSettingsRaw(t *testing.T, meta *clients.Clients, prior, next map[string]any, raw cty.Value) diag.Diagnostics {
	t.Helper()
	r := ResourceResourceOAuthSettings()
	d := schema.TestResourceDataRaw(t, r.Schema, prior)
	d.SetId(prior["resource_id"].(string))
	state := d.State()
	if !raw.IsNull() {
		state.RawConfig = raw
	}
	diff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(next), meta)
	require.NoError(t, err)
	if !raw.IsNull() {
		diff.RawConfig = raw
	}
	_, diags := r.Apply(t.Context(), state, diff, meta)
	return diags
}

func TestResourceOAuthSettingsUpdate(t *testing.T) {
	t.Parallel()

	var updates []*corev1.UpdateResourceOAuthSettingsRequest
	meta := newResourceOAuthSettingsTestClients(t, resourceOAuthSettingsTestServer{
		update: func(req *corev1.UpdateResourceOAuthSettingsRequest) (*corev1.UpdateResourceOAuthSettingsResponse, error) {
			updates = append(updates, req)
			return &corev1.UpdateResourceOAuthSettingsResponse{ResourceOauthSettings: &corev1.ResourceOAuthSettings{
				Id:                "resourceoauthsetting_1",
				ResourceId:        req.ResourceId,
				UpstreamOauthMode: upstreamOAuthModePreRegistered,
				RequestedScopes:   req.RequestedScopes,
				ClientId:          "Iv1.abc",
				ClientSecretSet:   req.ClientSecret == nil || req.GetClientSecret() != "",
				Issuer:            new("https://github.com/login/oauth"),
			}}, nil
		},
	})
	prior := map[string]any{
		"resource_id":         "resource_github",
		"upstream_oauth_mode": upstreamOAuthModePreRegistered,
		"client_id":           "Iv1.abc",
		"client_secret":       "shhh",
		"issuer":              "https://github.com/login/oauth",
		"requested_scopes":    []any{"repo", "read:org"},
	}
	with := func(changes map[string]any) map[string]any {
		next := maps.Clone(prior)
		maps.Copy(next, changes)
		return next
	}

	diags := applyResourceOAuthSettings(t, meta, prior, with(map[string]any{"requested_scopes": []any{"repo"}}))
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, updates, 1)
	require.Equal(t, []string{"repo"}, updates[0].RequestedScopes)
	require.Nil(t, updates[0].ClientSecret, "an unchanged secret is kept rather than resent")

	diags = applyResourceOAuthSettings(t, meta, prior, with(map[string]any{"client_secret": "rotated"}))
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, updates, 2)
	require.Equal(t, "rotated", updates[1].GetClientSecret())

	public := maps.Clone(prior)
	delete(public, "client_secret")
	diags = applyResourceOAuthSettings(t, meta, prior, public)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, updates, 3)
	require.NotNil(t, updates[2].ClientSecret, "removing the secret makes the client public")
	require.Empty(t, updates[2].GetClientSecret())
}

func TestResourceOAuthSettingsWriteOnlySecret(t *testing.T) {
	t.Parallel()

	var updates []*corev1.UpdateResourceOAuthSettingsRequest
	meta := newResourceOAuthSettingsTestClients(t, resourceOAuthSettingsTestServer{
		update: func(req *corev1.UpdateResourceOAuthSettingsRequest) (*corev1.UpdateResourceOAuthSettingsResponse, error) {
			updates = append(updates, req)
			return &corev1.UpdateResourceOAuthSettingsResponse{ResourceOauthSettings: &corev1.ResourceOAuthSettings{
				Id:                "resourceoauthsetting_1",
				ResourceId:        req.ResourceId,
				UpstreamOauthMode: upstreamOAuthModePreRegistered,
				RequestedScopes:   req.RequestedScopes,
				ClientId:          "Iv1.abc",
				ClientSecretSet:   req.ClientSecret == nil || req.GetClientSecret() != "",
				Issuer:            new("https://github.com/login/oauth"),
			}}, nil
		},
	})
	prior := map[string]any{
		"resource_id":              "resource_github",
		"upstream_oauth_mode":      upstreamOAuthModePreRegistered,
		"client_id":                "Iv1.abc",
		"issuer":                   "https://github.com/login/oauth",
		"client_secret_wo_version": 1,
		"requested_scopes":         []any{"repo", "read:org"},
	}
	raw := cty.ObjectVal(map[string]cty.Value{"client_secret_wo": cty.StringVal("s3cret")})
	with := func(changes map[string]any) map[string]any {
		next := maps.Clone(prior)
		maps.Copy(next, changes)
		return next
	}

	diags := applyResourceOAuthSettingsRaw(t, meta, prior, with(map[string]any{"requested_scopes": []any{"repo"}}), raw)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, updates, 1)
	require.Nil(t, updates[0].ClientSecret, "an unchanged write-only secret is kept rather than resent")

	diags = applyResourceOAuthSettingsRaw(t, meta, prior, with(map[string]any{"client_secret_wo_version": 2}), raw)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, updates, 2)
	require.Equal(t, "s3cret", updates[1].GetClientSecret())

	diags = applyResourceOAuthSettingsRaw(t, meta, prior, with(map[string]any{"client_secret_wo_version": 3}), cty.ObjectVal(map[string]cty.Value{
		"client_secret_wo": cty.StringVal(""),
	}))
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, updates, 3)
	require.NotNil(t, updates[2].ClientSecret, "an empty write-only secret makes the client public")
	require.Empty(t, updates[2].GetClientSecret())
}

func TestResourceOAuthSettingsCreateAndRead(t *testing.T) {
	t.Parallel()

	settings := &corev1.ResourceOAuthSettings{
		Id:                "resourceoauthsetting_1",
		ResourceId:        "resource_github",
		UpstreamOauthMode: upstreamOAuthModePreRegistered,
		RequestedScopes:   []string{"repo"},
		ClientId:          "Iv1.abc",
		ClientSecretSet:   true,
		Issuer:            proto.String("https://github.com/login/oauth"),
		CreatedAt:         timestamppb.Now(),
		UpdatedAt:         timestamppb.Now(),
	}
	var created *corev1.CreateResourceOAuthSettingsRequest
	var readResourceID string
	meta := newResourceOAuthSettingsTestClients(t, resourceOAuthSettingsTestServer{
		create: func(req *corev1.CreateResourceOAuthSettingsRequest) (*corev1.CreateResourceOAuthSettingsResponse, error) {
			created = req
			return &corev1.CreateResourceOAuthSettingsResponse{ResourceOauthSettings: settings}, nil
		},
		get: func(req *corev1.GetResourceOAuthSettingsRequest) (*corev1.GetResourceOAuthSettingsResponse, error) {
			readResourceID = req.ResourceId
			return &corev1.GetResourceOAuthSettingsResponse{ResourceOauthSettings: settings}, nil
		},
	})

	r := ResourceResourceOAuthSettings()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"resource_id":         "resource_github",
		"upstream_oauth_mode": upstreamOAuthModePreRegistered,
		"requested_scopes":    []any{"repo"},
		"client_id":           "Iv1.abc",
		"client_secret":       "shhh",
		"issuer":              "https://github.com/login/oauth",
	})

	diags := r.CreateContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "resource_github", created.ResourceId)
	require.Equal(t, upstreamOAuthModePreRegistered, created.UpstreamOauthMode)
	require.Equal(t, []string{"repo"}, created.RequestedScopes)
	require.Equal(t, "Iv1.abc", created.GetClientId())
	require.Equal(t, "shhh", created.ClientSecret)
	require.Equal(t, "https://github.com/login/oauth", created.GetIssuer())

	require.Equal(t, "resource_github", d.Id())
	require.Equal(t, "resourceoauthsetting_1", d.Get("oauth_settings_id"))
	require.True(t, d.Get("client_secret_set").(bool))
	require.Equal(t, "shhh", d.Get("client_secret"))

	diags = r.ReadContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "resource_github", readResourceID)
	require.Equal(t, "shhh", d.Get("client_secret"))
}

func TestResourceOAuthSettingsCreateWriteOnlySecret(t *testing.T) {
	t.Parallel()

	var created *corev1.CreateResourceOAuthSettingsRequest
	meta := newResourceOAuthSettingsTestClients(t, resourceOAuthSettingsTestServer{
		create: func(req *corev1.CreateResourceOAuthSettingsRequest) (*corev1.CreateResourceOAuthSettingsResponse, error) {
			created = req
			return &corev1.CreateResourceOAuthSettingsResponse{ResourceOauthSettings: &corev1.ResourceOAuthSettings{
				Id:                "resourceoauthsetting_1",
				ResourceId:        req.ResourceId,
				UpstreamOauthMode: upstreamOAuthModePreRegistered,
				ClientId:          req.GetClientId(),
				ClientSecretSet:   req.ClientSecret != "",
			}}, nil
		},
	})
	r := ResourceResourceOAuthSettings()
	diff, err := r.Diff(t.Context(), nil, terraform.NewResourceConfigRaw(map[string]any{
		"resource_id":              "resource_github",
		"upstream_oauth_mode":      upstreamOAuthModePreRegistered,
		"client_id":                "Iv1.abc",
		"client_secret_wo_version": 1,
	}), meta)
	require.NoError(t, err)
	diff.RawConfig = cty.ObjectVal(map[string]cty.Value{"client_secret_wo": cty.StringVal("s3cret")})

	state, diags := r.Apply(t.Context(), nil, diff, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "s3cret", created.ClientSecret)
	require.NotContains(t, state.Attributes, "client_secret_wo")
	require.Empty(t, state.Attributes["client_secret"])
	require.Equal(t, "1", state.Attributes["client_secret_wo_version"])
}

func TestResourceOAuthSettingsCreateOmitsUnsetOptionalFields(t *testing.T) {
	t.Parallel()

	var created *corev1.CreateResourceOAuthSettingsRequest
	meta := newResourceOAuthSettingsTestClients(t, resourceOAuthSettingsTestServer{
		create: func(req *corev1.CreateResourceOAuthSettingsRequest) (*corev1.CreateResourceOAuthSettingsResponse, error) {
			created = req
			return &corev1.CreateResourceOAuthSettingsResponse{ResourceOauthSettings: &corev1.ResourceOAuthSettings{
				Id:                "resourceoauthsetting_1",
				ResourceId:        req.ResourceId,
				UpstreamOauthMode: upstreamOAuthModeFormal,
			}}, nil
		},
	})

	r := ResourceResourceOAuthSettings()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"resource_id": "resource_mcp"})

	diags := r.CreateContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, upstreamOAuthModeFormal, created.UpstreamOauthMode)
	require.Empty(t, created.RequestedScopes)
	require.Nil(t, created.ClientId)
	require.Nil(t, created.Issuer)
	require.Empty(t, created.ClientSecret)
	require.Empty(t, d.Get("issuer"))
}

func TestResourceOAuthSettingsReadClearsMissingSettings(t *testing.T) {
	t.Parallel()

	meta := newResourceOAuthSettingsTestClients(t, resourceOAuthSettingsTestServer{
		get: func(*corev1.GetResourceOAuthSettingsRequest) (*corev1.GetResourceOAuthSettingsResponse, error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("resource OAuth settings not found"))
		},
	})
	r := ResourceResourceOAuthSettings()
	d := schema.TestResourceDataRaw(t, r.Schema, nil)
	d.SetId("resource_mcp")

	diags := r.ReadContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Empty(t, d.Id())
}

func TestResourceOAuthSettingsDeleteIgnoresMissingSettings(t *testing.T) {
	t.Parallel()

	var deletedResourceID string
	meta := newResourceOAuthSettingsTestClients(t, resourceOAuthSettingsTestServer{
		delete: func(req *corev1.DeleteResourceOAuthSettingsRequest) (*corev1.DeleteResourceOAuthSettingsResponse, error) {
			deletedResourceID = req.ResourceId
			return nil, connect.NewError(connect.CodeNotFound, errors.New("resource OAuth settings not found"))
		},
	})
	r := ResourceResourceOAuthSettings()
	d := schema.TestResourceDataRaw(t, r.Schema, nil)
	d.SetId("resource_mcp")

	diags := r.DeleteContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "resource_mcp", deletedResourceID)
	require.Empty(t, d.Id())
}
