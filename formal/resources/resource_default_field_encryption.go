package resource

import (
	"context"
	"fmt"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"strings"

	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDefaultFieldEncryption() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Creating Field Encryptions in Formal.",
		CreateContext: resourceDefaultFieldEncryptionCreateOrUpdate,
		ReadContext:   resourceDefaultFieldEncryptionRead,
		UpdateContext: resourceDefaultFieldEncryptionCreateOrUpdate,
		DeleteContext: resourceDefaultFieldEncryptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"data_key_storage": {
				// This description is used by the documentation generator and the language server.
				Description: "How the encrypted data key that encrypts the data should be stored. Use `control_plane_and_with_data` if the encrypted data key should be stored in the database alongside the encrypted data. Use `control_plane_only` if the encrypted data key should only be stored in the Formal Control Plane. In both cases, the data key is encrypted by the encryption key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"kms_key_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Formal ID of the encryption key to be used for this field encryption. This encryption key will be used to encrypt the data key. Read about envelope encryption here: https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#enveloping ",
				Type:        schema.TypeString,
				Required:    true,
			},
			"encryption_alg": {
				// This description is used by the documentation generator and the language server.
				Description: "Encryption Algorithm to use. Supported values are `aes_random` and `aes_deterministic`. For highest security, `aes_random` is recommended, but `aes_deterministic` is required to enable search (WHERE clauses) over underlying data in encrypted fields.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// Done
func resourceDefaultFieldEncryptionCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newDefaultFieldEncryption := api.DefaultFieldEncryptionStruct{
		DataKeyStorage: d.Get("data_key_storage").(string),
		KmsKeyID:       d.Get("kms_key_id").(string),
		EncryptionAlg:  d.Get("encryption_alg").(string),
	}

	defaultFieldEncryption, err := c.Http.CreateOrUpdateDefaultFieldEncryption(newDefaultFieldEncryption)
	if err != nil {
		return diag.FromErr(err)
	}
	if defaultFieldEncryption == nil {
		return diag.FromErr(err)
	}

	d.SetId("default")

	resourceDefaultFieldEncryptionRead(ctx, d, meta)

	return diags
}

func resourceDefaultFieldEncryptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	resourceId := d.Id()

	defaultFieldEncryption, err := c.Http.GetDefaultFieldEncryption()
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Was deleted
			tflog.Warn(ctx, "The Default Field Encryption was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if defaultFieldEncryption == nil {
		return diags
	}

	// Should map to all tracked fields of DefaultFieldEncryptionOrgItem
	d.Set("data_key_storage", defaultFieldEncryption.DataKeyStorage)
	d.Set("kms_key_id", defaultFieldEncryption.KmsKeyID)
	d.Set("encryption_alg", defaultFieldEncryption.EncryptionAlg)

	d.SetId(resourceId)

	return diags
}

// DONE
func resourceDefaultFieldEncryptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	err := c.Http.DeleteDefaultFieldEncryption()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
