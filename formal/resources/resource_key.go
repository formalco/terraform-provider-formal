package resource

import (
	"context"

	"github.com/formalco/terraform-provider-formal/formal/api"
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

		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for this key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "The friendly name for this key. NOTE: for data recovery purposes, we do not enable keys to be deleted or updated -- please consider this before applying your terraform configuration.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"org_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The Formal ID for your organisation.",
				Type:        schema.TypeString,
				Computed:    true,
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
	client := meta.(*api.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Right: field names according to KeyStruct. Using these for consistency
	// with KeyStruct, after phasing out createkeypayload struct on both here and admin api
	newKey := api.KeyStruct{
		KeyName:        d.Get("name").(string),
		KeyId:          d.Get("key_id").(string),
		CloudRegion:    d.Get("cloud_region").(string),
		KeyType:        d.Get("key_type").(string),
		ManagedBy:      d.Get("managed_by").(string),
		CloudAccountID: d.Get("cloud_account_id").(string),
	}

	key, err := client.CreateKey(newKey)
	if err != nil {
		return diag.FromErr(err)
	}
	if key == nil {
		return diag.FromErr(err)
	}

	// we use our own terraform key instead of real id bc it helps us encode values when READing by this terraformid
	d.SetId(key.Id)

	resourceKeyRead(ctx, d, meta)

	return diags
}

func resourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.Client)
	var diags diag.Diagnostics

	key, err := client.GetKey(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if key == nil {
		return diags
	}


	// Should map to all fields of KeyOrgItem
	d.Set("id", key.Id)
	d.Set("arn", key.KeyArn)
	d.Set("name", key.KeyName)
	d.Set("org_id", key.OrgId)
	d.Set("key_id", key.KeyId)
	d.Set("cloud_region", key.CloudRegion)
	d.Set("arn", key.KeyArn)
	d.Set("active", key.Active)
	d.Set("key_type", key.KeyType)
	d.Set("managed_by", key.ManagedBy)
	d.Set("cloud_account_id", key.CloudAccountID)

	d.SetId(key.Id)

	return diags
}

func resourceKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("Keys are immutable at the moment, but will be updateable soon. Please create a new key. Thank you!")

	// client := meta.(*Client)

	// keyId := d.Id()

	// keyUpdate := KeyOrgItem{
	// 	Name:        d.Get("name").(string),
	// 	Description: d.Get("description").(string),
	// 	Module:      d.Get("module").(string),
	// }

	// client.UpdateKey(keyId, keyUpdate)
	// return resourceKeyRead(ctx, d, meta)
}

// TODO delete
func resourceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
	// return diag.Errorf("Keys are not deleteable at the moment to ensure data recovery.")
}
