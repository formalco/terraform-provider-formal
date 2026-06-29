package resource

import (
	"context"
	"strings"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceTlsConfiguration() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a TLS Configuration of a Resource in Formal.",

		CreateContext: resourceTlsConfigurationCreate,
		ReadContext:   resourceTlsConfigurationRead,
		UpdateContext: resourceTlsConfigurationUpdate,
		DeleteContext: resourceTlsConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourcePolicyInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourcePolicyStateUpgradeV0,
			},
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "ID of the TLS Configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Resource ID for which the TLS configuration is applied to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"tls_config": {
				// This description is used by the documentation generator and the language server.
				Description: "Validation mode for the TLS configuration. Supported values are: `disable` (no TLS), `insecure-skip-verify` (TLS without verification), `insecure-verify-ca-only` (verify CA only), `verify-full` (full certificate verification).",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"disable",
					"insecure-skip-verify",
					"insecure-verify-ca-only",
					"verify-full",
				}, false),
			},
			"tls_min_version": {
				// This description is used by the documentation generator and the language server.
				Description: "Minimum TLS version to be used for connections.",
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"TLSv1.2",
					"TLSv1.3",
				}, false),
				Default: "TLSv1.3",
			},
			"tls_ca_truststore": {
				// This description is used by the documentation generator and the language server.
				Description: "PEM encoded CA certificate to verify resource certificates. Only required if resource certificates are not trusted by the root CA truststore.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tls_client_cert": {
				// This description is used by the documentation generator and the language server.
				Description: "Client certificate the connector presents to the resource for mutual TLS. Either the PEM, or the name of an environment variable read on the connector when `tls_client_cert_is_env` is set.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tls_client_key": {
				// This description is used by the documentation generator and the language server.
				Description: "Private key paired with `tls_client_cert`. Either the PEM, or the name of an environment variable read on the connector when `tls_client_key_is_env` is set.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"tls_client_cert_is_env": {
				// This description is used by the documentation generator and the language server.
				Description: "When true, `tls_client_cert` is the name of an environment variable read on the connector rather than the literal PEM.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"tls_client_key_is_env": {
				// This description is used by the documentation generator and the language server.
				Description: "When true, `tls_client_key` is the name of an environment variable read on the connector rather than the literal PEM.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func resourceTlsConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	resourceId := d.Get("resource_id").(string)
	tlsConfig := d.Get("tls_config").(string)
	tlsMinVersion := d.Get("tls_min_version").(string)
	tlsCaTrustStore := d.Get("tls_ca_truststore").(string)

	msg := &corev1.CreateResourceTlsConfigurationRequest{
		ResourceId:         resourceId,
		TlsConfig:          tlsConfig,
		TlsMinVersion:      tlsMinVersion,
		TlsCaTruststore:    tlsCaTrustStore,
		TlsClientCert:      d.Get("tls_client_cert").(string),
		TlsClientKey:       d.Get("tls_client_key").(string),
		TlsClientCertIsEnv: d.Get("tls_client_cert_is_env").(bool),
		TlsClientKeyIsEnv:  d.Get("tls_client_key_is_env").(bool),
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err = v.Validate(msg); err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateResourceTlsConfiguration(ctx, connect.NewRequest(msg))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ResourceTlsConfiguration.Id)

	resourceTlsConfigurationRead(ctx, d, meta)
	return diags
}

func resourceTlsConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	id := corev1.GetResourceTlsConfigurationRequest_ResourceTlsConfigurationId{
		ResourceTlsConfigurationId: d.Id(),
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetResourceTlsConfiguration(ctx, connect.NewRequest(&corev1.GetResourceTlsConfigurationRequest{Id: &id}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("resource_id", res.Msg.ResourceTlsConfiguration.ResourceId)
	d.Set("tls_config", res.Msg.ResourceTlsConfiguration.TlsConfig)
	d.Set("tls_min_version", res.Msg.ResourceTlsConfiguration.TlsMinVersion)
	d.Set("tls_ca_truststore", res.Msg.ResourceTlsConfiguration.TlsCaTruststore)
	d.Set("tls_client_cert", res.Msg.ResourceTlsConfiguration.TlsClientCert)
	d.Set("tls_client_key", res.Msg.ResourceTlsConfiguration.TlsClientKey)
	d.Set("tls_client_cert_is_env", res.Msg.ResourceTlsConfiguration.TlsClientCertIsEnv)
	d.Set("tls_client_key_is_env", res.Msg.ResourceTlsConfiguration.TlsClientKeyIsEnv)

	d.SetId(res.Msg.ResourceTlsConfiguration.Id)

	return diags
}

func resourceTlsConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	resourceTlsConfig := d.Id()

	fieldsThatCanChange := []string{"tls_config", "tls_min_version", "tls_ca_truststore", "tls_client_cert", "tls_client_key", "tls_client_cert_is_env", "tls_client_key_is_env"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	tlsConfig := d.Get("tls_config").(string)
	tlsMinVersion := d.Get("tls_min_version").(string)
	tlsCaTrustStore := d.Get("tls_ca_truststore").(string)
	tlsClientCert := d.Get("tls_client_cert").(string)
	tlsClientKey := d.Get("tls_client_key").(string)
	tlsClientCertIsEnv := d.Get("tls_client_cert_is_env").(bool)
	tlsClientKeyIsEnv := d.Get("tls_client_key_is_env").(bool)

	_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResourceTlsConfiguration(ctx, connect.NewRequest(&corev1.UpdateResourceTlsConfigurationRequest{
		Id:                 resourceTlsConfig,
		TlsConfig:          tlsConfig,
		TlsMinVersion:      tlsMinVersion,
		TlsCaTruststore:    tlsCaTrustStore,
		TlsClientCert:      &tlsClientCert,
		TlsClientKey:       &tlsClientKey,
		TlsClientCertIsEnv: &tlsClientCertIsEnv,
		TlsClientKeyIsEnv:  &tlsClientKeyIsEnv,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceTlsConfigurationRead(ctx, d, meta)

	return diags
}

func resourceTlsConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceTlsConfigurationId := d.Id()

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceTlsConfiguration(ctx, connect.NewRequest(&corev1.DeleteResourceTlsConfigurationRequest{Id: resourceTlsConfigurationId}))
	if err != nil {
		tflog.Warn(ctx, err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
