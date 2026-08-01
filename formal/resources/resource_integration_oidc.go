package resource

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

const oidcAudiencePrefix = "oidc.formal.ai/"

func ResourceIntegrationOIDC() *schema.Resource {
	return &schema.Resource{
		Description: "Registers an OIDC trust configuration used to authenticate externally signed JWTs to the Formal control plane as a machine user.",

		CreateContext: resourceIntegrationOIDCCreate,
		ReadContext:   resourceIntegrationOIDCRead,
		UpdateContext: resourceIntegrationOIDCUpdate,
		DeleteContext: resourceIntegrationOIDCDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique identifier of the OIDC integration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"audience": {
				Description: "The Formal audience value that tokens must present: `oidc.formal.ai/{id}`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description:  "Friendly name for the OIDC integration. Must be unique within the organization.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"issuer": {
				Description:  "OIDC issuer URL. Must be absolute HTTPS (path preserved). In local `ENV=dev`, the configured fake OIDC provider origin may use HTTP.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateOIDCIssuerURL,
			},
			"jwks_uri": {
				Description:  "Optional JWKS URI. When unset, Formal uses OIDC Discovery from the issuer. Must be absolute HTTPS when set.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateOIDCJWKSURI,
			},
			"machine_user_id": {
				Description: "ID of the Formal machine user that authenticated OIDC tokens map to.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"claim_condition": {
				Description: "CEL expression evaluated against verified token claims. Must return a boolean. Defaults to `true`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "true",
			},
			"status": {
				Description: "Integration status. Accepted values are `active` and `draft`. Draft disables authentication.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "active",
				ValidateFunc: validation.StringInSlice([]string{
					"active",
					"draft",
				}, false),
			},
			"created_at": {
				Description: "When the integration was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "When the integration was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func validateOIDCIssuerURL(v any, key string) (warns []string, errs []error) {
	raw, ok := v.(string)
	if !ok || strings.TrimSpace(raw) == "" {
		errs = append(errs, fmt.Errorf("%q must be a non-empty string", key))
		return warns, errs
	}
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		errs = append(errs, fmt.Errorf("%q must be a valid absolute URL", key))
		return warns, errs
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		errs = append(errs, fmt.Errorf("%q scheme must be https (http only for local fake OIDC provider)", key))
		return warns, errs
	}
	if parsed.User != nil {
		errs = append(errs, fmt.Errorf("%q must not include user credentials", key))
		return warns, errs
	}
	if parsed.Fragment != "" {
		errs = append(errs, fmt.Errorf("%q must not include a fragment", key))
		return warns, errs
	}
	if parsed.RawQuery != "" {
		errs = append(errs, fmt.Errorf("%q must not include a query", key))
		return warns, errs
	}
	return warns, errs
}

func validateOIDCJWKSURI(v any, key string) (warns []string, errs []error) {
	raw, ok := v.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("%q must be a string", key))
		return warns, errs
	}
	if strings.TrimSpace(raw) == "" {
		return warns, errs
	}
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		errs = append(errs, fmt.Errorf("%q must be a valid absolute URL", key))
		return warns, errs
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		errs = append(errs, fmt.Errorf("%q scheme must be https (http only for local fake OIDC provider)", key))
		return warns, errs
	}
	if parsed.User != nil {
		errs = append(errs, fmt.Errorf("%q must not include user credentials", key))
		return warns, errs
	}
	if parsed.Fragment != "" {
		errs = append(errs, fmt.Errorf("%q must not include a fragment", key))
		return warns, errs
	}
	return warns, errs
}

func oidcAudience(id string) string {
	return oidcAudiencePrefix + id
}

func resourceIntegrationOIDCCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	req := &corev1.CreateIntegrationOIDCRequest{
		Name:           d.Get("name").(string),
		Issuer:         d.Get("issuer").(string),
		MachineUserId:  d.Get("machine_user_id").(string),
		ClaimCondition: d.Get("claim_condition").(string),
		Status:         d.Get("status").(string),
	}
	if v, ok := d.GetOk("jwks_uri"); ok {
		jwksURI := v.(string)
		req.JwksUri = &jwksURI
	}

	res, err := c.Grpc.Sdk.IntegrationOIDCServiceClient.CreateIntegrationOIDC(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Integration.Id)
	return resourceIntegrationOIDCRead(ctx, d, meta)
}

func resourceIntegrationOIDCRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	id := d.Id()
	res, err := c.Grpc.Sdk.IntegrationOIDCServiceClient.GetIntegrationOIDC(ctx, &corev1.GetIntegrationOIDCRequest{Id: id})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The OIDC integration with ID "+id+" was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return flattenIntegrationOIDC(d, res.Integration)
}

func resourceIntegrationOIDCUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	integration := &corev1.IntegrationOIDC{
		Id:             d.Id(),
		Name:           d.Get("name").(string),
		Issuer:         d.Get("issuer").(string),
		MachineUserId:  d.Get("machine_user_id").(string),
		ClaimCondition: d.Get("claim_condition").(string),
		Status:         d.Get("status").(string),
	}
	if v, ok := d.GetOk("jwks_uri"); ok {
		jwksURI := v.(string)
		integration.JwksUri = &jwksURI
	}

	_, err := c.Grpc.Sdk.IntegrationOIDCServiceClient.UpdateIntegrationOIDC(ctx, &corev1.UpdateIntegrationOIDCRequest{
		Integration: integration,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceIntegrationOIDCRead(ctx, d, meta)
}

func resourceIntegrationOIDCDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	_, err := c.Grpc.Sdk.IntegrationOIDCServiceClient.DeleteIntegrationOIDC(ctx, &corev1.DeleteIntegrationOIDCRequest{Id: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func flattenIntegrationOIDC(d *schema.ResourceData, integration *corev1.IntegrationOIDC) diag.Diagnostics {
	d.Set("id", integration.Id)
	d.Set("audience", oidcAudience(integration.Id))
	d.Set("name", integration.Name)
	d.Set("issuer", integration.Issuer)
	d.Set("machine_user_id", integration.MachineUserId)
	d.Set("claim_condition", integration.ClaimCondition)
	d.Set("status", integration.Status)
	if integration.JwksUri != nil {
		d.Set("jwks_uri", *integration.JwksUri)
	} else {
		d.Set("jwks_uri", nil)
	}
	if integration.CreatedAt != nil {
		d.Set("created_at", integration.CreatedAt.AsTime().UTC().Format(time.RFC3339))
	}
	if integration.UpdatedAt != nil {
		d.Set("updated_at", integration.UpdatedAt.AsTime().UTC().Format(time.RFC3339))
	}
	return nil
}
