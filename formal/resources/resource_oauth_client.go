package resource

import (
	"context"
	"errors"
	"maps"
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

func ResourceOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Registers an OAuth client, such as Cursor or Devin, that signs users in to the organization's OAuth-enabled Resources (see `formal_resource_oauth_settings`) with a static client ID instead of a client ID metadata document. Destroying or replacing it revokes every consent and refresh token issued to the client.",

		CreateContext: resourceOAuthClientCreate,
		ReadContext:   resourceOAuthClientRead,
		UpdateContext: resourceOAuthClientUpdate,
		DeleteContext: resourceOAuthClientDelete,
		CustomizeDiff: resourceOAuthClientCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The OAuth `client_id` to configure in the client.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description:  "Name of the client, shown to users on the consent screen.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"redirect_uris": {
				Description: "Exact redirect URIs the client may use: `https`, `http` on a loopback host (any port), or a private-use scheme containing a period.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    20,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 2048),
				},
			},
			"confidential": {
				Description: "Whether the client is confidential and authenticates with a Formal-generated secret, using `client_secret_basic` or `client_secret_post`. Public clients use PKCE alone. Changing it forces replacement.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"client_secret": {
				Description: "The client secret of a confidential client. The API returns it only when the client is created or its secret is rotated, so it is empty after import, and it is not stored when the provider sets `retrieve_sensitive_values = false`.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"client_secret_rotation_triggers": {
				Description: "Arbitrary map of values that, when changed, rotates the secret of a confidential client. The old secret stops working immediately; existing grants are kept. Adding the map, for example after an import, or removing it doesn't rotate the secret.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"created_at": {
				Description: "When the client was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "When the client was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// rotatesClientSecret reports whether a change of client_secret_rotation_triggers
// rotates the secret. Adding the triggers, as on the first apply after an
// import, records them without revoking the live secret, and removing them
// doesn't rotate either.
func rotatesClientSecret(oldTriggers, newTriggers any) bool {
	oldMap, newMap := oldTriggers.(map[string]any), newTriggers.(map[string]any)
	return len(oldMap) > 0 && len(newMap) > 0 && !maps.Equal(oldMap, newMap)
}

func resourceOAuthClientCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if d.Id() == "" || d.HasChange("confidential") || !d.HasChange("client_secret_rotation_triggers") {
		return nil
	}
	oldTriggers, newTriggers := d.GetChange("client_secret_rotation_triggers")
	// Unknown triggers may still rotate on apply, so the plan must allow a new secret.
	// The SDK reports a map with unknown values through its count key.
	unknown := !d.NewValueKnown("client_secret_rotation_triggers.%") && len(oldTriggers.(map[string]any)) > 0
	if !unknown && !rotatesClientSecret(oldTriggers, newTriggers) {
		return nil
	}
	if !d.Get("confidential").(bool) {
		return errors.New("client_secret_rotation_triggers can only rotate the secret of a confidential client")
	}
	return d.SetNewComputed("client_secret")
}

func resourceOAuthClientCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	req := &corev1.CreateOAuthClientRequest{
		Name:         d.Get("name").(string),
		RedirectUris: lo.Map(d.Get("redirect_uris").([]any), func(v any, _ int) string { return v.(string) }),
		Confidential: d.Get("confidential").(bool),
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err := v.Validate(req); err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.OAuthServiceClient.CreateOAuthClient(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.OauthClient.Id)
	if c.Grpc.ReturnSensitiveValue {
		d.Set("client_secret", res.ClientSecret)
	}
	return flattenOAuthClient(d, res.OauthClient)
}

func resourceOAuthClientRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	id := d.Id()
	res, err := c.Grpc.Sdk.OAuthServiceClient.GetOAuthClient(ctx, &corev1.GetOAuthClientRequest{Id: id})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The OAuth client with ID "+id+" was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return flattenOAuthClient(d, res.OauthClient)
}

func resourceOAuthClientUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	if d.HasChanges("name", "redirect_uris") {
		name := d.Get("name").(string)
		req := &corev1.UpdateOAuthClientRequest{
			Id:           d.Id(),
			Name:         &name,
			RedirectUris: lo.Map(d.Get("redirect_uris").([]any), func(v any, _ int) string { return v.(string) }),
		}

		v, err := protovalidate.New()
		if err != nil {
			return diag.FromErr(err)
		}
		if err := v.Validate(req); err != nil {
			return diag.FromErr(err)
		}

		res, err := c.Grpc.Sdk.OAuthServiceClient.UpdateOAuthClient(ctx, req)
		if err != nil {
			return diag.FromErr(err)
		}
		flattenOAuthClient(d, res.OauthClient)
	}

	if rotatesClientSecret(d.GetChange("client_secret_rotation_triggers")) {
		res, err := c.Grpc.Sdk.OAuthServiceClient.CreateOAuthClientSecret(ctx, &corev1.CreateOAuthClientSecretRequest{Id: d.Id()})
		if err != nil {
			return diag.FromErr(err)
		}
		if c.Grpc.ReturnSensitiveValue {
			d.Set("client_secret", res.ClientSecret)
		}
		flattenOAuthClient(d, res.OauthClient)
	}

	return nil
}

func resourceOAuthClientDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.OAuthServiceClient.DeleteOAuthClient(ctx, &corev1.DeleteOAuthClientRequest{Id: d.Id()})
	if err != nil && connect.CodeOf(err) != connect.CodeNotFound {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func flattenOAuthClient(d *schema.ResourceData, client *corev1.OAuthClient) diag.Diagnostics {
	d.Set("name", client.Name)
	d.Set("redirect_uris", client.RedirectUris)
	d.Set("confidential", client.ClientSecretSet)
	if client.CreatedAt != nil {
		d.Set("created_at", client.CreatedAt.AsTime().UTC().Format(time.RFC3339))
	}
	if client.UpdatedAt != nil {
		d.Set("updated_at", client.UpdatedAt.AsTime().UTC().Format(time.RFC3339))
	}
	return nil
}
