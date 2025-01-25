package resource

import (
	"context"
	"fmt"
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
				Description: "The ID of the key in the provider's system.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"region": {
				Description: "The provider's region where the key is located.",
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

	req := &corev1.CreateEncryptionKeyRequest{
		Provider:  d.Get("key_provider").(string),
		KeyId:     d.Get("key_id").(string),
		Region:    d.Get("region").(string),
		Algorithm: d.Get("algorithm").(string),
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
	d.Set("region", res.Msg.EncryptionKey.Region)
	d.Set("algorithm", res.Msg.EncryptionKey.Algorithm)
	d.Set("created_at", res.Msg.EncryptionKey.CreatedAt.AsTime().String())
	d.Set("updated_at", res.Msg.EncryptionKey.UpdatedAt.AsTime().String())

	d.SetId(res.Msg.EncryptionKey.Id)
	return diags
}

func resourceEncryptionKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	keyId := d.Id()

	fieldsThatCanChange := []string{"key_provider", "key_id", "region", "algorithm"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
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
	if d.HasChange("region") {
		region := d.Get("region").(string)
		req.Region = &region
	}
	if d.HasChange("algorithm") {
		algorithm := d.Get("algorithm").(string)
		req.Algorithm = &algorithm
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
