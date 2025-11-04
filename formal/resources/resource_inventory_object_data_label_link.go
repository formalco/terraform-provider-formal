package resource

import (
	"context"
	"strings"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceInventoryObjectDataLabelLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Data Label with Formal.",
		CreateContext: resourceInventoryObjectDataLabelLinkCreate,
		ReadContext:   resourceInventoryObjectDataLabelLinkRead,
		UpdateContext: resourceInventoryObjectDataLabelLinkUpdate,
		DeleteContext: resourceInventoryObjectDataLabelLinkDelete,
		SchemaVersion: 1,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Resource ID to which the inventory object belongs.",
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
			},
			"path": {
				// This description is used by the documentation generator and the language server.
				Description: "Path of the inventory object.",
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
			},
			"data_label": {
				// This description is used by the documentation generator and the language server.
				Description: "Data label to link to the inventory object.",
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
			},
			"locked": {
				// This description is used by the documentation generator and the language server.
				Description: "Whether the inventory object is locked.",
				Type:        schema.TypeBool,
				Required:    true,
				Optional:    false,
			},
		},
	}
}

func resourceInventoryObjectDataLabelLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dataLabel := d.Get("data_label").(string)
	lockStatusValidated := d.Get("locked").(bool)

	createReq := &corev1.UpdateColumnRequest{
		DatastoreId:         d.Get("resource_id").(string),
		Path:                d.Get("path").(string),
		DataLabel:           &dataLabel,
		LockStatusValidated: &lockStatusValidated,
	}

	res, err := c.Grpc.Sdk.InventoryServiceClient.UpdateColumn(ctx, connect.NewRequest(createReq))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Id)
	resourceInventoryObjectDataLabelLinkRead(ctx, d, meta)

	d.Set("resource_id", d)

	return diags
}

func resourceInventoryObjectDataLabelLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.InventoryServiceClient.GetInventoryObject(ctx, connect.NewRequest(&corev1.GetInventoryObjectRequest{Id: d.Id()}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Data Label was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if res.Msg.Object.GetColumn() != nil {
		d.Set("resource_id", res.Msg.Object.GetColumn().ResourceId)
		d.Set("path", res.Msg.Object.GetColumn().Path)
		d.Set("data_label", res.Msg.Object.GetColumn().DataLabel)
		d.Set("locked", res.Msg.Object.GetColumn().DataLabelLockedForSidecar)
	}

	return diags
}

func resourceInventoryObjectDataLabelLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	fieldsThatCanChange := []string{"data_label", "locked"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	dataLabel := d.Get("data_label").(string)
	lockStatusValidated := d.Get("locked").(bool)

	_, err := c.Grpc.Sdk.InventoryServiceClient.UpdateColumn(ctx, connect.NewRequest(&corev1.UpdateColumnRequest{
		DatastoreId:         d.Get("resource_id").(string),
		Path:                d.Get("path").(string),
		DataLabel:           &dataLabel,
		LockStatusValidated: &lockStatusValidated,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceInventoryObjectDataLabelLinkRead(ctx, d, meta)

	return diags
}

func resourceInventoryObjectDataLabelLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	emptyDataLabel := ""
	emptyLockStatusValidated := false

	_, err := c.Grpc.Sdk.InventoryServiceClient.UpdateColumn(ctx, connect.NewRequest(&corev1.UpdateColumnRequest{
		DatastoreId:         d.Get("resource_id").(string),
		Path:                d.Get("path").(string),
		DataLabel:           &emptyDataLabel,
		LockStatusValidated: &emptyLockStatusValidated,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
