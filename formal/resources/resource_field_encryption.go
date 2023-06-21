package resource

import (
	"context"
	"errors"
	"fmt"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"strings"

	"github.com/formalco/terraform-provider-formal/formal/api"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceFieldEncryption() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Creating Field Encryptions in Formal.",
		CreateContext: resourceFieldEncryptionCreate,
		ReadContext:   resourceFieldEncryptionRead,
		UpdateContext: resourceFieldEncryptionUpdate,
		DeleteContext: resourceFieldEncryptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "Formal ID for this resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"datastore_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Formal ID of the datastore that this Field Encryption should be applied to.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"path": {
				// This description is used by the documentation generator and the language server.
				Description: "The full path of the field that should be encrypted, separated by `.` (eg `{databaseName}.{schemaName}.{tableName}.{columnName}`)",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key_storage": {
				// This description is used by the documentation generator and the language server.
				Description: "How the encrypted data key that encrypts the data should be stored. Use `control_plane_and_with_data` if the encrypted data key should be stored in the database alongside the encrypted data. Use `control_plane_only` if the encrypted data key should only be stored in the Formal Control Plane. In both cases, the data key is encrypted by the encryption key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Formal ID of the encryption key to be used for this field encryption. This encryption key will be used to encrypt the data key. Read about envelope encryption here: https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#enveloping ",
				Type:        schema.TypeString,
				Required:    true,
			},
			"alg": {
				// This description is used by the documentation generator and the language server.
				Description: "Encryption Algorithm to use. Supported values are `aes_random` and `aes_deterministic`. For highest security, `aes_random` is recommended, but `aes_deterministic` is required to enable search (WHERE clauses) over underlying data in encrypted fields.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

const fieldEncryptionTerraformIdDelimiter = "#_#"

// Done
func resourceFieldEncryptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	newFieldEncryption := api.FieldEncryptionStruct{
		DsId:       d.Get("datastore_id").(string),
		Path:       d.Get("path").(string),
		KeyStorage: d.Get("key_storage").(string),
		KeyId:      d.Get("key_id").(string),
		Alg:        d.Get("alg").(string),
	}

	fieldEncryption, err := c.Http.CreateFieldEncryption(newFieldEncryption)
	if err != nil {
		return diag.FromErr(err)
	}
	if fieldEncryption == nil {
		return diag.FromErr(err)
	}

	d.SetId(fieldEncryption.DsId + fieldEncryptionTerraformIdDelimiter + fieldEncryption.Path)

	resourceFieldEncryptionRead(ctx, d, meta)

	return diags
}

func resourceFieldEncryptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	terraformFieldEncryptionId := d.Id()
	// Split
	terraformFieldEncryptionIdSplit := strings.Split(terraformFieldEncryptionId, fieldEncryptionTerraformIdDelimiter)
	if len(terraformFieldEncryptionIdSplit) != 2 {
		return diag.FromErr(errors.New("id for encryption field is malformatted"))
	}
	dsId := terraformFieldEncryptionIdSplit[0]
	path := terraformFieldEncryptionIdSplit[1]

	fieldEncryption, err := c.Http.GetFieldEncryption(dsId, path)
	if err != nil {
		if strings.Contains(fmt.Sprint(err), "status: 404") {
			// Was deleted
			tflog.Warn(ctx, "The Field Encryption was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	if fieldEncryption == nil {
		return diags
	}

	// Should map to all tracked fields of FieldEncryptionOrgItem
	d.Set("datastore_id", fieldEncryption.DsId)
	d.Set("path", fieldEncryption.Path)
	// d.Set("name", fieldEncryption.FieldName)
	d.Set("key_storage", fieldEncryption.KeyStorage)
	d.Set("key_id", fieldEncryption.KeyId)
	d.Set("alg", fieldEncryption.Alg)

	d.SetId(terraformFieldEncryptionId)

	return diags
}

// Doesn't exist
func resourceFieldEncryptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Field Encryptions are immutable. Please create a new Field Encryption. Thank you!")
	// client := meta.(*Client)

	// fieldEncryptionId := d.Id()

	// fieldEncryptionUpdate := FieldEncryption{
	// 	Name:        d.Get("name").(string),
	// 	Description: d.Get("description").(string),
	// 	Module:      d.Get("module").(string),
	// }

	// client.UpdateFieldEncryption(fieldEncryptionId, fieldEncryptionUpdate)
	// return resourceFieldEncryptionRead(ctx, d, meta)
}

// DONE
func resourceFieldEncryptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	terraformFieldEncryptionId := d.Id()
	terraformFieldEncryptionIdSplit := strings.Split(terraformFieldEncryptionId, fieldEncryptionTerraformIdDelimiter)
	if len(terraformFieldEncryptionIdSplit) != 2 {
		return diag.FromErr(errors.New("id for terraform ID for this encryption field is malformatted"))

	}
	dsId := terraformFieldEncryptionIdSplit[0]
	path := terraformFieldEncryptionIdSplit[1]

	err := c.Http.DeleteFieldEncryption(dsId, path)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
