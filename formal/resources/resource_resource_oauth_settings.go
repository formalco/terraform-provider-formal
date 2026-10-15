package resource

import (
	"context"
	"errors"
	"time"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

const (
	upstreamOAuthModeFormal        = "formal"
	upstreamOAuthModePreRegistered = "pre_registered"
	upstreamOAuthModeNone          = "none"
)

// The API has no update RPC for OAuth settings, so every argument forces
// replacement. Replacing deletes every user's linked account and every OAuth
// client's authorization for the resource.
func ResourceResourceOAuthSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Enables OAuth on an MCP, HTTP, gRPC or LLM Resource. OAuth clients (see `formal_oauth_client`) can then connect to the Resource through the Connector and, unless `upstream_oauth_mode` is `none`, users link their own upstream accounts, whose access tokens the Connector sends upstream in place of native user credentials. Every argument forces replacement, and replacing or destroying this resource deletes every user's linked account and every OAuth client's authorization for the Resource.",

		CreateContext: resourceResourceOAuthSettingsCreate,
		ReadContext:   resourceResourceOAuthSettingsRead,
		DeleteContext: resourceResourceOAuthSettingsDelete,
		CustomizeDiff: resourceResourceOAuthSettingsCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the Resource the OAuth settings belong to. Import with the Resource ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"oauth_settings_id": {
				Description: "The ID of the OAuth settings.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				Description: "The ID of the Resource to enable OAuth on. Its technology must be `mcp`, `http`, `grpc` or `llm`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"upstream_oauth_mode": {
				Description: "How Formal identifies itself to the upstream authorization server when users link their accounts. `formal` (the default) uses Formal's OAuth app: a client Formal registered with the authorization server when there is one, and otherwise its Client ID Metadata Document client. `pre_registered` uses an OAuth app your organization registered with the authorization server, set with `client_id` and `client_secret`. `none` links no upstream account: the Connector authenticates upstream with the user's native user.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     upstreamOAuthModeFormal,
				ValidateFunc: validation.StringInSlice([]string{
					upstreamOAuthModeFormal,
					upstreamOAuthModePreRegistered,
					upstreamOAuthModeNone,
				}, false),
			},
			"requested_scopes": {
				Description: "Scopes to request from the upstream authorization server. When empty, the scope parameter is omitted and the provider's defaults apply. Must be empty when `upstream_oauth_mode` is `none`.",
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    50,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
			},
			"client_id": {
				Description: "The client ID of your OAuth app. Required when `upstream_oauth_mode` is `pre_registered`, and not allowed otherwise. Register the Formal redirect URI shown in the console with the authorization server.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"client_secret": {
				Description: "The client secret of your OAuth app. Leave unset for a public client, which authenticates with PKCE alone. Only allowed when `upstream_oauth_mode` is `pre_registered`. The API never returns it, so it is not read back after creation, and setting it on imported settings that already have a secret does not replace them.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				// Imported state has no secret, and replacing the settings to
				// record it would delete every linked account.
				DiffSuppressFunc: func(_, oldValue, newValue string, d *schema.ResourceData) bool {
					return d.Id() != "" && oldValue == "" && newValue != "" && d.Get("client_secret_set").(bool)
				},
			},
			"client_secret_set": {
				Description: "Whether the OAuth app has a client secret.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"issuer": {
				Description:  "The upstream authorization server's issuer, such as `https://github.com/login/oauth` for the GitHub REST API. Set it when the upstream publishes no protected resource metadata (RFC 9728): Formal then reads the issuer's authorization server metadata directly. It must use https, have no query or fragment, and exactly match the issuer the metadata reports. When unset, the authorization server is discovered from the upstream itself. Not allowed when `upstream_oauth_mode` is `none`.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"created_at": {
				Description: "When the OAuth settings were created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "When the OAuth settings were last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

type resourceOAuthSettingsArgs struct {
	mode            string
	hasScopes       bool
	hasClientID     bool
	hasClientSecret bool
	hasIssuer       bool
}

func validateResourceOAuthSettingsArgs(args resourceOAuthSettingsArgs) error {
	switch {
	case args.mode == upstreamOAuthModeNone && args.hasScopes:
		return errors.New("requested_scopes must be empty when upstream_oauth_mode is none")
	case args.mode == upstreamOAuthModeNone && args.hasIssuer:
		return errors.New("issuer must be unset when upstream_oauth_mode is none")
	case args.mode == upstreamOAuthModePreRegistered && !args.hasClientID:
		return errors.New("client_id is required when upstream_oauth_mode is pre_registered")
	case args.mode != upstreamOAuthModePreRegistered && (args.hasClientID || args.hasClientSecret):
		return errors.New("client_id and client_secret are only allowed when upstream_oauth_mode is pre_registered")
	}
	return nil
}

func resourceResourceOAuthSettingsCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	// Values computed from other resources are unknown at plan time; the API
	// checks them on apply instead.
	for _, key := range []string{"upstream_oauth_mode", "requested_scopes", "client_id", "client_secret", "issuer"} {
		if !d.NewValueKnown(key) {
			return nil
		}
	}
	_, hasClientID := d.GetOk("client_id")
	_, hasClientSecret := d.GetOk("client_secret")
	_, hasIssuer := d.GetOk("issuer")
	return validateResourceOAuthSettingsArgs(resourceOAuthSettingsArgs{
		mode:            d.Get("upstream_oauth_mode").(string),
		hasScopes:       d.Get("requested_scopes").(*schema.Set).Len() > 0,
		hasClientID:     hasClientID,
		hasClientSecret: hasClientSecret,
		hasIssuer:       hasIssuer,
	})
}

func resourceResourceOAuthSettingsCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	req := &corev1.CreateResourceOAuthSettingsRequest{
		ResourceId:        d.Get("resource_id").(string),
		UpstreamOauthMode: d.Get("upstream_oauth_mode").(string),
		RequestedScopes:   lo.Map(d.Get("requested_scopes").(*schema.Set).List(), func(v any, _ int) string { return v.(string) }),
		ClientSecret:      d.Get("client_secret").(string),
	}
	if v, ok := d.GetOk("client_id"); ok {
		clientID := v.(string)
		req.ClientId = &clientID
	}
	if v, ok := d.GetOk("issuer"); ok {
		issuer := v.(string)
		req.Issuer = &issuer
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err := v.Validate(req); err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateResourceOAuthSettings(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.ResourceOauthSettings.ResourceId)
	return flattenResourceOAuthSettings(d, res.ResourceOauthSettings)
}

func resourceResourceOAuthSettingsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	resourceID := d.Id()
	res, err := c.Grpc.Sdk.ResourceServiceClient.GetResourceOAuthSettings(ctx, &corev1.GetResourceOAuthSettingsRequest{ResourceId: resourceID})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The OAuth settings of the resource with ID "+resourceID+" were not found, which means they may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return flattenResourceOAuthSettings(d, res.ResourceOauthSettings)
}

func resourceResourceOAuthSettingsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceOAuthSettings(ctx, &corev1.DeleteResourceOAuthSettingsRequest{ResourceId: d.Id()})
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func flattenResourceOAuthSettings(d *schema.ResourceData, settings *corev1.ResourceOAuthSettings) diag.Diagnostics {
	d.Set("oauth_settings_id", settings.Id)
	d.Set("resource_id", settings.ResourceId)
	d.Set("upstream_oauth_mode", settings.UpstreamOauthMode)
	d.Set("requested_scopes", settings.RequestedScopes)
	d.Set("client_id", settings.ClientId)
	d.Set("client_secret_set", settings.ClientSecretSet)
	d.Set("issuer", settings.GetIssuer())
	if settings.CreatedAt != nil {
		d.Set("created_at", settings.CreatedAt.AsTime().UTC().Format(time.RFC3339))
	}
	if settings.UpdatedAt != nil {
		d.Set("updated_at", settings.UpdatedAt.AsTime().UTC().Format(time.RFC3339))
	}
	return nil
}
