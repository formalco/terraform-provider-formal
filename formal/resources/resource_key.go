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

func ResourceKey() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Creating a key in Formal.",

		CreateContext: resourceKeyCreate,
		ReadContext:   resourceKeyRead,
		UpdateContext: resourceKeyUpdate,
		DeleteContext: resourceKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for this key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "The friendly name for this key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"org_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for your organisation.",
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated:  "field is deprecated, and will be removed on a future release",
			},
			"key_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the key as referenced in the key management service. Required only if the `managed_by` field is `customer_managed`; otherwise Formal creates the key and retrieves this value.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"cloud_region": {
				// This description is used by the documentation generator and the language server.
				Description: "The cloud region that the key should be created in.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"arn": {
				// This description is used by the documentation generator and the language server.
				Description: "ARN of the created key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"active": {
				// This description is used by the documentation generator and the language server.
				Description: "Active status of the key. For data accessibility, Formal does not delete its record of created keys.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"key_type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of key based on cloud provider. Supported values at the moment are `aws_kms`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"managed_by": {
				// This description is used by the documentation generator and the language server.
				Description: "How the key is managed. Supported values are `saas_managed`, `managed_cloud`, or `customer_managed`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"cloud_account_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Formal ID of the managed Cloud Account to be used to create the key. Required if managed_by is `managed_cloud`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

// Done
func resourceKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	KeyName := d.Get("name").(string)
	KeyId := d.Get("key_id").(string)
	CloudRegion := d.Get("cloud_region").(string)
	KeyType := d.Get("key_type").(string)
	ManagedBy := d.Get("managed_by").(string)
	CloudAccountID := d.Get("cloud_account_id").(string)

	res, err := c.Grpc.Sdk.KmsServiceClient.CreateKeyRegistration(ctx, connect.NewRequest(&adminv1.CreateKeyRegistrationRequest{
		CloudRegion:    CloudRegion,
		KeyId:          KeyId,
		ManagedBy:      ManagedBy,
		KeyType:        KeyType,
		KeyName:        KeyName,
		CloudAccountId: CloudAccountID,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Key.Id)

	resourceKeyRead(ctx, d, meta)

	return diags
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.KmsServiceClient.GetKey(ctx, connect.NewRequest(&adminv1.GetKeyRequest{Id: d.Id()}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Key not found
			tflog.Warn(ctx, "The key was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Should map to all fields of KeyOrgItem
	d.Set("id", res.Msg.Key.Id)
	d.Set("arn", res.Msg.Key.KeyArn)
	d.Set("name", res.Msg.Key.Name)
	d.Set("key_id", res.Msg.Key.KeyId)
	d.Set("cloud_region", res.Msg.Key.CloudRegion)
	d.Set("arn", res.Msg.Key.KeyArn)
	d.Set("active", res.Msg.Key.Active)
	d.Set("key_type", res.Msg.Key.KeyType)
	d.Set("managed_by", res.Msg.Key.ManagedBy)
	d.Set("cloud_account_id", res.Msg.Key.CloudAccountId)

	d.SetId(res.Msg.Key.Id)

	return diags
}

func resourceKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Keys are immutable at the moment, but will be updateable soon. Please create a new key. Thank you!")

	// 	c := meta.(*clients.Clients)

	// keyId := d.Id()

	// keyUpdate := KeyOrgItem{
	// 	Name:        d.Get("name").(string),
	// 	Description: d.Get("description").(string),
	// 	Module:      d.Get("module").(string),
	// }

	// client.UpdateKey(keyId, keyUpdate)
	// return resourceKeyRead(ctx, d, meta)
}

func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	keyId := d.Id()

	_, err := c.Grpc.Sdk.KmsServiceClient.DeactivateFieldEncryptionKey(ctx, connect.NewRequest(&adminv1.DeactivateFieldEncryptionKeyRequest{Id: keyId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
