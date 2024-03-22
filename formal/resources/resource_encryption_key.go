package resource

import (
	"context"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceEncryptionKey() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Creating an Encryption Key in Formal.",
		CreateContext: resourceEncryptionKeyCreate,
		ReadContext:   resourceEncryptionKeyRead,
		UpdateContext: resourceEncryptionKeyUpdate,
		DeleteContext: resourceEncryptionKeyDelete,
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
			"cloud_region": {
				// This description is used by the documentation generator and the language server.
				Description: "Aws cloud region where the encryption key will be created.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key_name": {
				// This description is used by the documentation generator and the language server.
				Description: "Name of the encryption key to be created.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Aws key id of the encryption key to be created.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceEncryptionKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Maps to user-defined fields
	CloudRegion := d.Get("cloud_region").(string)
	KeyName := d.Get("key_name").(string)
	KeyId := d.Get("key_id").(string)

	res, err := c.Grpc.Sdk.KmsServiceClient.CreateKeyRegistration(ctx, connect.NewRequest(&adminv1.CreateKeyRegistrationRequest{
		CloudRegion: CloudRegion,
		KeyName:     KeyName,
		KeyId:       KeyId,
		KeyType:     "aws_kms",
		ManagedBy:   "customer_managed",
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Key.Id)

	resourceEncryptionKeyRead(ctx, d, meta)

	return diags
}

func resourceEncryptionKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.KmsServiceClient.GetKey(ctx, connect.NewRequest(&adminv1.GetKeyRequest{Id: d.Id()}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Was deleted
			tflog.Warn(ctx, "The Encryption key was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if res.Msg.Key == nil {
		return diags
	}

	d.Set("cloud_region", res.Msg.Key.CloudRegion)
	d.Set("key_name", res.Msg.Key.Name)
	d.Set("key_id", res.Msg.Key.KeyId)

	d.SetId(d.Id())

	return diags
}

func resourceEncryptionKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Encryption keys are immutable. Please create a new Encryption key. Thank you!")
}

func resourceEncryptionKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	id := d.Id()

	_, err := c.Grpc.Sdk.KmsServiceClient.DeactivateFieldEncryptionKey(ctx, connect.NewRequest(&adminv1.DeactivateFieldEncryptionKeyRequest{Id: id}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
