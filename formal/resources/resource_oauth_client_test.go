package resource

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/hashicorp/go-cty/cty"
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

type oauthClientTestServer struct {
	corev1connect.UnimplementedOAuthServiceHandler
	create       func(*corev1.CreateOAuthClientRequest) (*corev1.CreateOAuthClientResponse, error)
	get          func(*corev1.GetOAuthClientRequest) (*corev1.GetOAuthClientResponse, error)
	update       func(*corev1.UpdateOAuthClientRequest) (*corev1.UpdateOAuthClientResponse, error)
	createSecret func(*corev1.CreateOAuthClientSecretRequest) (*corev1.CreateOAuthClientSecretResponse, error)
	delete       func(*corev1.DeleteOAuthClientRequest) (*corev1.DeleteOAuthClientResponse, error)
}

func (s oauthClientTestServer) CreateOAuthClient(_ context.Context, req *connect.Request[corev1.CreateOAuthClientRequest]) (*connect.Response[corev1.CreateOAuthClientResponse], error) {
	res, err := s.create(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (s oauthClientTestServer) GetOAuthClient(_ context.Context, req *connect.Request[corev1.GetOAuthClientRequest]) (*connect.Response[corev1.GetOAuthClientResponse], error) {
	res, err := s.get(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (s oauthClientTestServer) UpdateOAuthClient(_ context.Context, req *connect.Request[corev1.UpdateOAuthClientRequest]) (*connect.Response[corev1.UpdateOAuthClientResponse], error) {
	res, err := s.update(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (s oauthClientTestServer) CreateOAuthClientSecret(_ context.Context, req *connect.Request[corev1.CreateOAuthClientSecretRequest]) (*connect.Response[corev1.CreateOAuthClientSecretResponse], error) {
	res, err := s.createSecret(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (s oauthClientTestServer) DeleteOAuthClient(_ context.Context, req *connect.Request[corev1.DeleteOAuthClientRequest]) (*connect.Response[corev1.DeleteOAuthClientResponse], error) {
	res, err := s.delete(req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func newOAuthClientTestClients(t *testing.T, server corev1connect.OAuthServiceHandler, returnSensitiveValue bool) *clients.Clients {
	t.Helper()

	mux := http.NewServeMux()
	path, handler := corev1connect.NewOAuthServiceHandler(server)
	mux.Handle(path, handler)
	httpServer := httptest.NewServer(mux)
	t.Cleanup(httpServer.Close)

	sdkClient, err := formal.New(formal.WithAPIKey("test-api-key"), formal.WithBaseURL(httpServer.URL))
	require.NoError(t, err)
	return &clients.Clients{Grpc: &api.GrpcClient{Sdk: sdkClient, ReturnSensitiveValue: returnSensitiveValue}}
}

// planOAuthClient returns the ResourceData Terraform hands to Update when
// config replaces state.
func planOAuthClient(t *testing.T, state map[string]string, config map[string]any) (*schema.ResourceData, error) {
	t.Helper()
	r := ResourceOAuthClient()
	instanceState := &terraform.InstanceState{ID: "client_1", Attributes: state}
	diff, err := r.Diff(t.Context(), instanceState, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		return nil, err
	}
	return schema.InternalMap(r.Schema).Data(instanceState, diff)
}

func confidentialOAuthClientState() map[string]string {
	return map[string]string{
		"id":                                    "client_1",
		"name":                                  "Desktop agent",
		"redirect_uris.#":                       "1",
		"redirect_uris.0":                       "com.example.agent://oauth/callback",
		"confidential":                          "true",
		"client_secret":                         "old-secret",
		"client_secret_rotation_triggers.%":     "1",
		"client_secret_rotation_triggers.epoch": "1",
	}
}

func TestOAuthClientSchema(t *testing.T) {
	t.Parallel()
	r := ResourceOAuthClient()
	require.NoError(t, r.InternalValidate(r.Schema, true))
	require.True(t, r.Schema["client_secret"].Sensitive)
	require.True(t, r.Schema["client_secret"].Computed)
	require.True(t, r.Schema["confidential"].ForceNew)
	require.False(t, r.Schema["name"].ForceNew)
	require.False(t, r.Schema["redirect_uris"].ForceNew)
}

func TestOAuthClientCreateStoresSecret(t *testing.T) {
	t.Parallel()

	var created *corev1.CreateOAuthClientRequest
	meta := newOAuthClientTestClients(t, oauthClientTestServer{
		create: func(req *corev1.CreateOAuthClientRequest) (*corev1.CreateOAuthClientResponse, error) {
			created = req
			return &corev1.CreateOAuthClientResponse{
				OauthClient: &corev1.OAuthClient{
					Id:              "client_1",
					Name:            req.Name,
					RedirectUris:    req.RedirectUris,
					ClientSecretSet: true,
					CreatedAt:       timestamppb.Now(),
					UpdatedAt:       timestamppb.Now(),
				},
				ClientSecret: "generated-secret",
			}, nil
		},
	}, true)

	r := ResourceOAuthClient()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"name":          "Cloud agent",
		"redirect_uris": []any{"https://agent.example.com/oauth/callback", "http://127.0.0.1/callback"},
		"confidential":  true,
	})

	diags := r.CreateContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "Cloud agent", created.Name)
	require.Equal(t, []string{"https://agent.example.com/oauth/callback", "http://127.0.0.1/callback"}, created.RedirectUris)
	require.True(t, created.Confidential)
	require.Equal(t, "client_1", d.Id())
	require.Equal(t, "generated-secret", d.Get("client_secret"))
	require.True(t, d.Get("confidential").(bool))
	require.NotEmpty(t, d.Get("created_at"))
}

func TestOAuthClientCreateHonorsRetrieveSensitiveValues(t *testing.T) {
	t.Parallel()

	meta := newOAuthClientTestClients(t, oauthClientTestServer{
		create: func(req *corev1.CreateOAuthClientRequest) (*corev1.CreateOAuthClientResponse, error) {
			return &corev1.CreateOAuthClientResponse{
				OauthClient:  &corev1.OAuthClient{Id: "client_1", Name: req.Name, RedirectUris: req.RedirectUris, ClientSecretSet: true},
				ClientSecret: "generated-secret",
			}, nil
		},
	}, false)

	r := ResourceOAuthClient()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{
		"name":          "Cloud agent",
		"redirect_uris": []any{"https://agent.example.com/oauth/callback"},
		"confidential":  true,
	})

	diags := r.CreateContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Empty(t, d.Get("client_secret"))
}

func TestOAuthClientUpdateRenamesAndReplacesRedirectURIs(t *testing.T) {
	t.Parallel()

	var updated *corev1.UpdateOAuthClientRequest
	meta := newOAuthClientTestClients(t, oauthClientTestServer{
		update: func(req *corev1.UpdateOAuthClientRequest) (*corev1.UpdateOAuthClientResponse, error) {
			updated = req
			return &corev1.UpdateOAuthClientResponse{OauthClient: &corev1.OAuthClient{
				Id:              req.Id,
				Name:            req.GetName(),
				RedirectUris:    req.RedirectUris,
				ClientSecretSet: true,
			}}, nil
		},
		createSecret: func(*corev1.CreateOAuthClientSecretRequest) (*corev1.CreateOAuthClientSecretResponse, error) {
			t.Error("secret rotated without a trigger change")
			return nil, connect.NewError(connect.CodeInternal, errors.New("unexpected rotation"))
		},
	}, true)

	d, err := planOAuthClient(t, confidentialOAuthClientState(), map[string]any{
		"name":                            "Desktop agent (prod)",
		"redirect_uris":                   []any{"com.example.agent://oauth/callback", "https://agent.example.com/oauth/callback"},
		"confidential":                    true,
		"client_secret_rotation_triggers": map[string]any{"epoch": "1"},
	})
	require.NoError(t, err)

	diags := ResourceOAuthClient().UpdateContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "client_1", updated.Id)
	require.Equal(t, "Desktop agent (prod)", updated.GetName())
	require.Equal(t, []string{"com.example.agent://oauth/callback", "https://agent.example.com/oauth/callback"}, updated.RedirectUris)
	require.Equal(t, "old-secret", d.Get("client_secret"))
}

func TestOAuthClientUpdateRotatesSecretWhenTriggersChange(t *testing.T) {
	t.Parallel()

	var rotatedID string
	meta := newOAuthClientTestClients(t, oauthClientTestServer{
		createSecret: func(req *corev1.CreateOAuthClientSecretRequest) (*corev1.CreateOAuthClientSecretResponse, error) {
			rotatedID = req.Id
			return &corev1.CreateOAuthClientSecretResponse{
				OauthClient: &corev1.OAuthClient{
					Id:              req.Id,
					Name:            "Desktop agent",
					RedirectUris:    []string{"com.example.agent://oauth/callback"},
					ClientSecretSet: true,
				},
				ClientSecret: "new-secret",
			}, nil
		},
		update: func(*corev1.UpdateOAuthClientRequest) (*corev1.UpdateOAuthClientResponse, error) {
			t.Error("client updated without a name or redirect URI change")
			return nil, connect.NewError(connect.CodeInternal, errors.New("unexpected update"))
		},
	}, true)

	d, err := planOAuthClient(t, confidentialOAuthClientState(), map[string]any{
		"name":                            "Desktop agent",
		"redirect_uris":                   []any{"com.example.agent://oauth/callback"},
		"confidential":                    true,
		"client_secret_rotation_triggers": map[string]any{"epoch": "2"},
	})
	require.NoError(t, err)

	diags := ResourceOAuthClient().UpdateContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "client_1", rotatedID)
	require.Equal(t, "new-secret", d.Get("client_secret"))
}

func TestOAuthClientAddingOrRemovingTriggersKeepsSecret(t *testing.T) {
	t.Parallel()

	meta := newOAuthClientTestClients(t, oauthClientTestServer{
		createSecret: func(*corev1.CreateOAuthClientSecretRequest) (*corev1.CreateOAuthClientSecretResponse, error) {
			t.Error("secret rotated when triggers were added or removed")
			return nil, connect.NewError(connect.CodeInternal, errors.New("unexpected rotation"))
		},
	}, true)

	imported := confidentialOAuthClientState()
	delete(imported, "client_secret_rotation_triggers.%")
	delete(imported, "client_secret_rotation_triggers.epoch")
	imported["client_secret"] = ""

	tests := map[string]struct {
		state  map[string]string
		config map[string]any
	}{
		"added after import": {
			state: imported,
			config: map[string]any{
				"name":                            "Desktop agent",
				"redirect_uris":                   []any{"com.example.agent://oauth/callback"},
				"confidential":                    true,
				"client_secret_rotation_triggers": map[string]any{"epoch": "1"},
			},
		},
		"removed": {
			state: confidentialOAuthClientState(),
			config: map[string]any{
				"name":          "Desktop agent",
				"redirect_uris": []any{"com.example.agent://oauth/callback"},
				"confidential":  true,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := ResourceOAuthClient()
			instanceState := &terraform.InstanceState{ID: "client_1", Attributes: tt.state}
			diff, err := r.Diff(t.Context(), instanceState, terraform.NewResourceConfigRaw(tt.config), nil)
			require.NoError(t, err)
			require.NotContains(t, diff.Attributes, "client_secret")

			d, err := schema.InternalMap(r.Schema).Data(instanceState, diff)
			require.NoError(t, err)
			diags := r.UpdateContext(t.Context(), d, meta)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			require.Equal(t, tt.state["client_secret"], d.Get("client_secret"))
		})
	}
}

func TestOAuthClientPlanAllowsNewSecretForUnknownTriggers(t *testing.T) {
	t.Parallel()
	r := ResourceOAuthClient()

	block := r.CoreConfigSchema()
	attrs := map[string]cty.Value{}
	for name, attrType := range block.ImpliedType().AttributeTypes() {
		attrs[name] = cty.NullVal(attrType)
	}
	attrs["name"] = cty.StringVal("Desktop agent")
	attrs["redirect_uris"] = cty.ListVal([]cty.Value{cty.StringVal("com.example.agent://oauth/callback")})
	attrs["confidential"] = cty.True
	attrs["client_secret_rotation_triggers"] = cty.MapVal(map[string]cty.Value{"epoch": cty.UnknownVal(cty.String)})

	instanceState := &terraform.InstanceState{ID: "client_1", Attributes: confidentialOAuthClientState()}
	diff, err := r.Diff(t.Context(), instanceState, terraform.NewResourceConfigShimmed(cty.ObjectVal(attrs), block), nil)
	require.NoError(t, err)
	require.Contains(t, diff.Attributes, "client_secret")
	require.True(t, diff.Attributes["client_secret"].NewComputed)
}

func TestOAuthClientPlanRejectsRotatingPublicClient(t *testing.T) {
	t.Parallel()

	state := confidentialOAuthClientState()
	state["confidential"] = "false"
	state["client_secret"] = ""

	_, err := planOAuthClient(t, state, map[string]any{
		"name":                            "Desktop agent",
		"redirect_uris":                   []any{"com.example.agent://oauth/callback"},
		"client_secret_rotation_triggers": map[string]any{"epoch": "2"},
	})
	require.ErrorContains(t, err, "only rotate the secret of a confidential client")
}

func TestOAuthClientPlanAllowsTriggerChangeWhenReplacing(t *testing.T) {
	t.Parallel()

	_, err := planOAuthClient(t, confidentialOAuthClientState(), map[string]any{
		"name":                            "Desktop agent",
		"redirect_uris":                   []any{"com.example.agent://oauth/callback"},
		"confidential":                    false,
		"client_secret_rotation_triggers": map[string]any{"epoch": "2"},
	})
	require.NoError(t, err)
}

func TestOAuthClientReadClearsMissingClient(t *testing.T) {
	t.Parallel()

	meta := newOAuthClientTestClients(t, oauthClientTestServer{
		get: func(*corev1.GetOAuthClientRequest) (*corev1.GetOAuthClientResponse, error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("OAuth client not found"))
		},
	}, true)
	r := ResourceOAuthClient()
	d := schema.TestResourceDataRaw(t, r.Schema, nil)
	d.SetId("client_1")

	diags := r.ReadContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Empty(t, d.Id())
}

func TestOAuthClientReadKeepsSecretFromState(t *testing.T) {
	t.Parallel()

	meta := newOAuthClientTestClients(t, oauthClientTestServer{
		get: func(req *corev1.GetOAuthClientRequest) (*corev1.GetOAuthClientResponse, error) {
			return &corev1.GetOAuthClientResponse{OauthClient: &corev1.OAuthClient{
				Id:              req.Id,
				Name:            "Desktop agent",
				RedirectUris:    []string{"com.example.agent://oauth/callback"},
				ClientSecretSet: true,
			}}, nil
		},
	}, true)

	d, err := planOAuthClient(t, confidentialOAuthClientState(), map[string]any{
		"name":                            "Desktop agent",
		"redirect_uris":                   []any{"com.example.agent://oauth/callback"},
		"confidential":                    true,
		"client_secret_rotation_triggers": map[string]any{"epoch": "1"},
	})
	require.NoError(t, err)

	diags := ResourceOAuthClient().ReadContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "old-secret", d.Get("client_secret"))
	require.True(t, d.Get("confidential").(bool))
}

func TestOAuthClientDeleteIgnoresMissingClient(t *testing.T) {
	t.Parallel()

	var deletedID string
	meta := newOAuthClientTestClients(t, oauthClientTestServer{
		delete: func(req *corev1.DeleteOAuthClientRequest) (*corev1.DeleteOAuthClientResponse, error) {
			deletedID = req.Id
			return nil, connect.NewError(connect.CodeNotFound, errors.New("OAuth client not found"))
		},
	}, true)
	r := ResourceOAuthClient()
	d := schema.TestResourceDataRaw(t, r.Schema, nil)
	d.SetId("client_1")

	diags := r.DeleteContext(t.Context(), d, meta)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "client_1", deletedID)
	require.Empty(t, d.Id())
}
