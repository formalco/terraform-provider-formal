package resource

import (
	"context"
	"fmt"
	"strings"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceConnectorTlsCertificate() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Manages a self-supplied TLS certificate for a Formal Connector. Note: this resource should not be used for Formal-managed certificates.",
		CreateContext: resourceConnectorTlsCertificateCreate,
		ReadContext:   resourceConnectorTlsCertificateRead,
		UpdateContext: resourceConnectorTlsCertificateUpdate,
		DeleteContext: resourceConnectorTlsCertificateDelete,
		SchemaVersion: 1,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of this TLS certificate.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector_hostname_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Connector hostname this certificate is for.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"domain": {
				// This description is used by the documentation generator and the language server.
				Description: "The domain for this TLS certificate.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"certificate": {
				// This description is used by the documentation generator and the language server.
				Description: "The TLS certificate in PEM format (full chain).",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"private_key": {
				// This description is used by the documentation generator and the language server.
				Description: "The private key for the TLS certificate in PEM format.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"managed": {
				// This description is used by the documentation generator and the language server.
				Description: "Whether this certificate is managed by Formal.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"expires_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When this certificate expires.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When this certificate was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				// This description is used by the documentation generator and the language server.
				Description: "When this certificate was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceConnectorTlsCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	req := &corev1.CreateConnectorTlsCertificateRequest{
		ConnectorHostnameId: d.Get("connector_hostname_id").(string),
		Domain:              d.Get("domain").(string),
		Certificate:         []byte(d.Get("certificate").(string)),
		PrivateKey:          []byte(d.Get("private_key").(string)),
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorTlsCertificate(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ConnectorTlsCertificate.Id)
	resourceConnectorTlsCertificateRead(ctx, d, meta)
	return diags
}

func resourceConnectorTlsCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorTlsCertificateId := d.Id()

	req := connect.NewRequest(&corev1.GetConnectorTlsCertificateRequest{
		Id: &corev1.GetConnectorTlsCertificateRequest_ConnectorTlsCertificateId{
			ConnectorTlsCertificateId: connectorTlsCertificateId,
		},
	})

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorTlsCertificate(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector TLS Certificate was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.ConnectorTlsCertificate.Id)
	d.Set("connector_hostname_id", res.Msg.ConnectorTlsCertificate.ConnectorHostnameId)
	d.Set("domain", res.Msg.ConnectorTlsCertificate.Domain)
	d.Set("managed", res.Msg.ConnectorTlsCertificate.Managed)

	if res.Msg.ConnectorTlsCertificate.ExpiresAt != nil {
		d.Set("expires_at", res.Msg.ConnectorTlsCertificate.ExpiresAt.AsTime().Format("2006-01-02T15:04:05Z"))
	}
	if res.Msg.ConnectorTlsCertificate.CreatedAt != nil {
		d.Set("created_at", res.Msg.ConnectorTlsCertificate.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z"))
	}
	if res.Msg.ConnectorTlsCertificate.UpdatedAt != nil {
		d.Set("updated_at", res.Msg.ConnectorTlsCertificate.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z"))
	}

	d.SetId(res.Msg.ConnectorTlsCertificate.Id)

	return diags
}

func resourceConnectorTlsCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorTlsCertificateId := d.Id()

	fieldsThatCanChange := []string{"certificate", "private_key"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	req := connect.NewRequest(&corev1.UpdateConnectorTlsCertificateRequest{
		Id: connectorTlsCertificateId,
	})

	if d.HasChange("certificate") {
		certificate := []byte(d.Get("certificate").(string))
		req.Msg.Certificate = certificate
	}

	if d.HasChange("private_key") {
		privateKey := []byte(d.Get("private_key").(string))
		req.Msg.PrivateKey = privateKey
	}

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnectorTlsCertificate(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceConnectorTlsCertificateRead(ctx, d, meta)
	return diags
}

func resourceConnectorTlsCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorTlsCertificateId := d.Id()

	req := connect.NewRequest(&corev1.DeleteConnectorTlsCertificateRequest{
		Id: connectorTlsCertificateId,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorTlsCertificate(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
