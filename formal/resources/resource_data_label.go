package resource

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceDataLabel() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Data Label with Formal.",
		CreateContext: resourceDataLabelCreate,
		ReadContext:   resourceDataLabelRead,
		UpdateContext: resourceDataLabelUpdate,
		DeleteContext: resourceDataLabelDelete,
		SchemaVersion: 1,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of this data label.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this data label.",
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
			},
			"classifier_type": {
				// This description is used by the documentation generator and the language server.
				Description: "Type of classifier for the data label (regex or prompt)",
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				ValidateFunc: validation.StringInSlice([]string{
					"regex",
					"prompt",
				}, false),
			},
			"classifier_data": {
				// This description is used by the documentation generator and the language server.
				Description: "Data for the classifier (pattern for regex or label name for prompt).",
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
			},
		},
	}
}

func resourceDataLabelCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	createReq := &corev1.CreateDataLabelRequest{
		Name:           d.Get("name").(string),
		ClassifierType: d.Get("classifier_type").(string),
		ClassifierData: d.Get("classifier_data").(string),
	}

	res, err := c.Grpc.Sdk.InventoryServiceClient.CreateDataLabel(ctx, createReq)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.DataLabel.Id)

	resourceDataLabelRead(ctx, d, meta)

	return diags
}

func resourceDataLabelRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	dataLabelId := d.Id()

	res, err := c.Grpc.Sdk.InventoryServiceClient.GetDataLabel(ctx, &corev1.GetDataLabelRequest{Id: dataLabelId})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Data Label was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.DataLabel.Id)
	d.Set("name", res.DataLabel.Name)
	d.Set("classifier_type", res.DataLabel.ClassifierType)
	d.Set("classifier_data", res.DataLabel.ClassifierData)

	d.SetId(res.DataLabel.Id)

	return diags
}

func resourceDataLabelUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	dataLabelId := d.Id()

	fieldsThatCanChange := []string{"name", "classifier_type", "classifier_data"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	name := d.Get("name").(string)
	classifier_type := d.Get("classifier_type").(string)
	classifier_data := d.Get("classifier_data").(string)

	_, err := c.Grpc.Sdk.InventoryServiceClient.UpdateDataLabel(ctx, &corev1.UpdateDataLabelRequest{
		Id:             dataLabelId,
		Name:           name,
		ClassifierType: classifier_type,
		ClassifierData: classifier_data,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	resourceDataLabelRead(ctx, d, meta)

	return diags
}

func resourceDataLabelDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	dataLabelId := d.Id()

	_, err := c.Grpc.Sdk.InventoryServiceClient.DeleteDataLabel(ctx, &corev1.DeleteDataLabelRequest{Id: dataLabelId})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
