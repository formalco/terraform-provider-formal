package resource

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceConnectorTokenEncryptionKey() *schema.Resource {
	return &schema.Resource{
		Description:   "Register a KMS key encryption key (KEK) for HTTP policy token encryption on a Connector.",
		CreateContext: resourceConnectorTokenEncryptionKeyCreate,
		ReadContext:   resourceConnectorTokenEncryptionKeyRead,
		DeleteContext: resourceConnectorTokenEncryptionKeyDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this connector token encryption key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector_id": {
				Description: "The ID of the Connector this token encryption key is linked to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"key_provider": {
				Description: "The KMS provider. One of 'aws-kms', 'gcp-kms', or 'azure-key-vault'.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"aws-kms",
					"gcp-kms",
					"azure-key-vault",
				}, false),
			},
			"key_id": {
				Description: "Provider-native key identifier (AWS KMS ARN, GCP CryptoKey resource name, or Azure Key Vault key URI).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"created_at": {
				Description: "When the connector token encryption key was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "Last update time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceConnectorTokenEncryptionKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorTokenEncryptionKey(ctx, &corev1.CreateConnectorTokenEncryptionKeyRequest{
		ConnectorId: d.Get("connector_id").(string),
		Provider:    d.Get("key_provider").(string),
		KeyId:       d.Get("key_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.ConnectorTokenEncryptionKey.Id)
	return resourceConnectorTokenEncryptionKeyRead(ctx, d, meta)
}

func resourceConnectorTokenEncryptionKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorTokenEncryptionKey(ctx, &corev1.GetConnectorTokenEncryptionKeyRequest{
		Identifier: &corev1.GetConnectorTokenEncryptionKeyRequest_Id{
			Id: d.Id(),
		},
	})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector Token Encryption Key was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	key := res.ConnectorTokenEncryptionKey
	d.Set("id", key.Id)
	d.Set("connector_id", key.ConnectorId)
	d.Set("key_provider", key.Provider)
	d.Set("key_id", key.KeyId)
	d.Set("created_at", key.CreatedAt.AsTime().String())
	d.Set("updated_at", key.UpdatedAt.AsTime().String())
	d.SetId(key.Id)
	return diags
}

func resourceConnectorTokenEncryptionKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorTokenEncryptionKey(ctx, &corev1.DeleteConnectorTokenEncryptionKeyRequest{Id: d.Id()})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector Token Encryption Key was not found, so remove it from state.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
