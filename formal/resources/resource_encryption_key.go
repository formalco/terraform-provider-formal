package resource

import (
	"context"
	"fmt"
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

func ResourceEncryptionKey() *schema.Resource {
	return &schema.Resource{
		Description:   "Registering an Encryption Key with Formal.",
		CreateContext: resourceEncryptionKeyCreate,
		ReadContext:   resourceEncryptionKeyRead,
		UpdateContext: resourceEncryptionKeyUpdate,
		DeleteContext: resourceEncryptionKeyDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ any) error {
			// Only constrain new keys: existing keys keep their stored (possibly
			// deprecated) provider/algorithm so they stay planable.
			if d.Id() != "" {
				return nil
			}
			if provider, _ := d.Get("key_provider").(string); provider == "aws" {
				return fmt.Errorf("key_provider %q is deprecated; create encryption keys with aws-kms", provider)
			}
			if alg, _ := d.Get("algorithm").(string); alg != "" && alg != "rsaes_oaep_sha256" {
				return fmt.Errorf("algorithm %q is no longer supported; create encryption keys with rsaes_oaep_sha256", alg)
			}
			return nil
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this encryption key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"key_provider": {
				Description: "The provider of the encryption key. One of 'aws-kms' or 'gcp-kms' ('aws' is a deprecated alias for 'aws-kms').",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"aws-kms",
					"gcp-kms",
					"aws",
				}, false),
			},
			"key_id": {
				Description: "The ID of the key in the provider's system (key ARN for AWS KMS, or the crypto key version resource name for GCP KMS).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"algorithm": {
				Description: "Deprecated. Symmetric and deterministic algorithms ('aes_random', 'aes_deterministic') are no longer supported. Encryption keys use asymmetric RSA ('rsaes_oaep_sha256'), which is the default.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Deprecated:  "Symmetric and deterministic algorithms are no longer supported. Encryption keys now use asymmetric RSA (rsaes_oaep_sha256) by default; this field can be removed.",
				ValidateFunc: validation.StringInSlice([]string{
					"aes_random",
					"aes_deterministic",
					"rsaes_oaep_sha256",
				}, false),
			},
			"decryptor_uri": {
				Description: "The URI of the decryptor (e.g., a URL to a Lambda function, either directly or via API Gateway). This is used to decrypt the data on the frontend only (and is never called by the Formal Control Plane backend).",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"public_key_pem": {
				Description: "PEM-encoded RSA public key for client-side encryption. Required for all encryption keys. Typically wired from another resource, e.g. `data.aws_kms_public_key.<name>.public_key_pem` for an asymmetric AWS KMS key.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"created_at": {
				Description: "When the encryption key was created.",
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

func resourceEncryptionKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	decryptorUri := d.Get("decryptor_uri").(string)
	algorithm := "rsaes_oaep_sha256"
	if v, ok := d.GetOk("algorithm"); ok {
		algorithm = v.(string)
	}
	req := &corev1.CreateEncryptionKeyRequest{
		Provider:     d.Get("key_provider").(string),
		KeyId:        d.Get("key_id").(string),
		Algorithm:    algorithm,
		DecryptorUri: &decryptorUri,
	}
	if pem := d.Get("public_key_pem").(string); pem != "" {
		req.PublicKeyPem = &pem
	}

	res, err := c.Grpc.Sdk.LogsServiceClient.CreateEncryptionKey(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.EncryptionKey.Id)
	return resourceEncryptionKeyRead(ctx, d, meta)
}

func resourceEncryptionKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	keyId := d.Id()

	res, err := c.Grpc.Sdk.LogsServiceClient.GetEncryptionKey(ctx, &corev1.GetEncryptionKeyRequest{Id: keyId})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Encryption Key was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.EncryptionKey.Id)
	d.Set("key_provider", res.EncryptionKey.Provider)
	d.Set("key_id", res.EncryptionKey.KeyId)
	d.Set("algorithm", res.EncryptionKey.Algorithm)
	d.Set("decryptor_uri", res.EncryptionKey.DecryptorUri)
	d.Set("public_key_pem", res.EncryptionKey.PublicKeyPem)
	d.Set("created_at", res.EncryptionKey.CreatedAt.AsTime().String())
	d.Set("updated_at", res.EncryptionKey.UpdatedAt.AsTime().String())

	d.SetId(res.EncryptionKey.Id)
	return diags
}

func resourceEncryptionKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	keyId := d.Id()

	fieldsThatCanChange := []string{"key_provider", "key_id", "algorithm", "decryptor_uri", "public_key_pem"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	req := &corev1.UpdateEncryptionKeyRequest{
		Id: keyId,
	}

	if d.HasChange("key_provider") {
		keyProvider := d.Get("key_provider").(string)
		req.Provider = &keyProvider
	}
	if d.HasChange("key_id") {
		keyId := d.Get("key_id").(string)
		req.KeyId = &keyId
	}
	if d.HasChange("algorithm") {
		algorithm := d.Get("algorithm").(string)
		req.Algorithm = &algorithm
	}
	if d.HasChange("decryptor_uri") {
		decryptorUri := d.Get("decryptor_uri").(string)
		req.DecryptorUri = &decryptorUri
	}
	if d.HasChange("public_key_pem") {
		pem := d.Get("public_key_pem").(string)
		req.PublicKeyPem = &pem
	}

	_, err := c.Grpc.Sdk.LogsServiceClient.UpdateEncryptionKey(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceEncryptionKeyRead(ctx, d, meta)
}

func resourceEncryptionKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	keyId := d.Id()

	_, err := c.Grpc.Sdk.LogsServiceClient.DeleteEncryptionKey(ctx, &corev1.DeleteEncryptionKeyRequest{Id: keyId})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
