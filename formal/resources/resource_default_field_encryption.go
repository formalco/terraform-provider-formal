package resource

import (
	"context"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"github.com/bufbuild/connect-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/formalco/terraform-provider-formal/formal/validation"
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
				Description:  "How the encrypted data key that encrypts the data should be stored. Use `control_plane_and_with_data` if the encrypted data key should be stored in the database alongside the encrypted data. Use `control_plane_only` if the encrypted data key should only be stored in the Formal Control Plane. In both cases, the data key is encrypted by the encryption key.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.KeyStorage(),
			},
			"kms_key_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Formal ID of the encryption key to be used for this field encryption. This encryption key will be used to encrypt the data key. Read about envelope encryption here: https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#enveloping ",
				Type:        schema.TypeString,
				Required:    true,
			},
			"encryption_alg": {
				// This description is used by the documentation generator and the language server.
				Description:  "Encryption Algorithm to use. Supported values are `aes_random` and `aes_deterministic`. For highest security, `aes_random` is recommended, but `aes_deterministic` is required to enable search (WHERE clauses) over underlying data in encrypted fields.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.EncryptionAlgorithm(),
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this field encryption policy cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
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
	DataKeyStorage := d.Get("data_key_storage").(string)
	KmsKeyID := d.Get("kms_key_id").(string)
	EncryptionAlg := d.Get("encryption_alg").(string)
	TerminationProtection := d.Get("termination_protection").(bool)

	_, err := c.Grpc.Sdk.FieldEncryptionPolicyServiceClient.CreateOrUpdateDefaultFieldEncryptionPolicy(ctx, connect.NewRequest(&adminv1.CreateOrUpdateDefaultFieldEncryptionPolicyRequest{KmsKeyId: KmsKeyID, DataKeyStorage: DataKeyStorage, EncryptionAlg: EncryptionAlg, TerminationProtection: TerminationProtection}))
	if err != nil {
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

	defaultFieldEncryption, err := c.Grpc.Sdk.FieldEncryptionPolicyServiceClient.GetDefaultFieldEncryptionPolicy(ctx, connect.NewRequest(&adminv1.GetDefaultFieldEncryptionPolicyRequest{}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
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
	d.Set("data_key_storage", defaultFieldEncryption.Msg.DefaultFieldEncryptionPolicy.DataKeyStorage)
	d.Set("kms_key_id", defaultFieldEncryption.Msg.DefaultFieldEncryptionPolicy.KmsKeyId)
	d.Set("encryption_alg", defaultFieldEncryption.Msg.DefaultFieldEncryptionPolicy.EncryptionAlg)
	d.Set("termination_protection", defaultFieldEncryption.Msg.DefaultFieldEncryptionPolicy.TerminationProtection)

	d.SetId(resourceId)

	return diags
}

func resourceDefaultFieldEncryptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("At the moment you can't delete a default field encryption policy. Thank you!")
}
