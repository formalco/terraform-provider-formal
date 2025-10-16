package resource

import (
	"context"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this encryption key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"key_provider": {
				Description: "The provider of the encryption key. Currently only 'aws' is supported.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"aws",
				}, false),
			},
			"key_id": {
				Description: "The ID of the key in the provider's system (e.g., key ARN for AWS KMS).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"algorithm": {
				Description: "The algorithm used for encryption. Can be either 'aes_random' or 'aes_deterministic'.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"aes_random",
					"aes_deterministic",
				}, false),
			},
			"decryptor_uri": {
				Description: "The URI of the decryptor (e.g., a URL to a Lambda function, either directly or via API Gateway). This is used to decrypt the data on the frontend only (and is never called by the Formal Control Plane backend).",
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

func resourceEncryptionKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	decryptorUri := d.Get("decryptor_uri").(string)
	req := &corev1.CreateEncryptionKeyRequest{
		Provider:     d.Get("key_provider").(string),
		KeyId:        d.Get("key_id").(string),
		Algorithm:    d.Get("algorithm").(string),
		DecryptorUri: &decryptorUri,
	}

	res, err := c.Grpc.Sdk.LogsServiceClient.CreateEncryptionKey(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.EncryptionKey.Id)
	return resourceEncryptionKeyRead(ctx, d, meta)
}

func resourceEncryptionKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	keyId := d.Id()

	res, err := c.Grpc.Sdk.LogsServiceClient.GetEncryptionKey(ctx, connect.NewRequest(&corev1.GetEncryptionKeyRequest{Id: keyId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Encryption Key was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.EncryptionKey.Id)
	d.Set("key_provider", res.Msg.EncryptionKey.Provider)
	d.Set("key_id", res.Msg.EncryptionKey.KeyId)
	d.Set("algorithm", res.Msg.EncryptionKey.Algorithm)
	d.Set("decryptor_uri", res.Msg.EncryptionKey.DecryptorUri)
	d.Set("created_at", res.Msg.EncryptionKey.CreatedAt.AsTime().String())
	d.Set("updated_at", res.Msg.EncryptionKey.UpdatedAt.AsTime().String())

	d.SetId(res.Msg.EncryptionKey.Id)
	return diags
}

func resourceEncryptionKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	keyId := d.Id()

	fieldsThatCanChange := []string{"key_provider", "key_id", "algorithm", "decryptor_uri"}
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

	_, err := c.Grpc.Sdk.LogsServiceClient.UpdateEncryptionKey(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceEncryptionKeyRead(ctx, d, meta)
}

func resourceEncryptionKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	keyId := d.Id()

	_, err := c.Grpc.Sdk.LogsServiceClient.DeleteEncryptionKey(ctx, connect.NewRequest(&corev1.DeleteEncryptionKeyRequest{Id: keyId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
