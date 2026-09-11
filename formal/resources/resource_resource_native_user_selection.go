package resource

import (
	"context"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceResourceNativeUserSelection() *schema.Resource {
	return &schema.Resource{
		Description: "Controls which Native User V3 a session connects to a Resource as using a CEL expression. Enable Native Users V3 on the Resource separately with `native_users_v3_enabled`.",

		CreateContext: resourceNativeUserSelectionCreate,
		ReadContext:   resourceNativeUserSelectionRead,
		UpdateContext: resourceNativeUserSelectionUpdate,
		DeleteContext: resourceNativeUserSelectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the Resource this selection applies to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				Description: "The ID of the Resource this selection applies to. Only one selection can exist per Resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"cel": {
				Description:  "The CEL expression that returns the ID of the Native User a session connects as, or an empty string to refuse the session.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCelSyntaxAttribute,
			},
		},
	}
}

func resourceNativeUserSelectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	resourceId := d.Get("resource_id").(string)
	expression := d.Get("cel").(string)

	// Nothing is created server-side: this writes the expression onto the Resource.
	_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResource(ctx, &corev1.UpdateResourceRequest{
		Id:                     resourceId,
		NativeUserSelectionCel: &expression,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceId)

	return resourceNativeUserSelectionRead(ctx, d, meta)
}

func resourceNativeUserSelectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	resourceId := d.Id()

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetResource(ctx, &corev1.GetResourceRequest{Id: resourceId})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Resource "+resourceId+" was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if res.Resource.NativeUserSelectionCel == nil {
		tflog.Warn(ctx, "The Resource "+resourceId+" does not have a Native User V3 selection expression.", nil)
		d.SetId("")
		return diags
	}

	d.Set("resource_id", res.Resource.Id)
	d.Set("cel", res.Resource.GetNativeUserSelectionCel())
	d.SetId(res.Resource.Id)

	return diags
}

func resourceNativeUserSelectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	expression := d.Get("cel").(string)

	_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResource(ctx, &corev1.UpdateResourceRequest{
		Id:                     d.Id(),
		NativeUserSelectionCel: &expression,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNativeUserSelectionRead(ctx, d, meta)
}

func resourceNativeUserSelectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceNativeUserSelection(ctx, &corev1.DeleteResourceNativeUserSelectionRequest{
		Id: d.Id(),
	})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
