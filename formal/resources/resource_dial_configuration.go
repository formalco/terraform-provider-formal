package resource

import (
	"context"
	"fmt"
	"strings"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceDialConfiguration() *schema.Resource {
	return &schema.Resource{
		Description: "Creating a Dial Configuration of a Resource in Formal.",

		CreateContext: resourceDialConfigurationCreate,
		ReadContext:   resourceDialConfigurationRead,
		UpdateContext: resourceDialConfigurationUpdate,
		DeleteContext: resourceDialConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
			if d.Get("dial_method").(string) != "gcp_cloudsql" {
				return nil
			}
			if !d.NewValueKnown("dial_target") {
				return nil
			}
			if d.Get("dial_target").(string) == "" {
				return fmt.Errorf("dial_target is required when dial_method is gcp_cloudsql")
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the Dial Configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				Description: "Resource ID for which the dial configuration is applied to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"dial_method": {
				Description: "How the connector dials this resource's upstream. Supported values are: `tcp` (direct TCP via the resource's hostname and port), `gcp_cloudsql` (dial via the GCP Cloud SQL connector library — `dial_target` must be set to the `project:region:instance` connection name).",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"tcp",
					"gcp_cloudsql",
				}, false),
			},
			"dial_target": {
				Description: "Method-specific dial target. For `gcp_cloudsql`, the `project:region:instance` connection name. Leave empty for `tcp`.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
		},
	}
}

func resourceDialConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	resourceId := d.Get("resource_id").(string)
	dialMethod := d.Get("dial_method").(string)
	dialTarget := d.Get("dial_target").(string)

	msg := &corev1.CreateResourceDialConfigurationRequest{
		ResourceId: resourceId,
		DialMethod: dialMethod,
		DialTarget: dialTarget,
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err = v.Validate(msg); err != nil {
		return diag.FromErr(err)
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.CreateResourceDialConfiguration(ctx, connect.NewRequest(msg))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ResourceDialConfiguration.Id)

	return resourceDialConfigurationRead(ctx, d, meta)
}

func resourceDialConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	id := corev1.GetResourceDialConfigurationRequest_ResourceDialConfigurationId{
		ResourceDialConfigurationId: d.Id(),
	}

	res, err := c.Grpc.Sdk.ResourceServiceClient.GetResourceDialConfiguration(ctx, connect.NewRequest(&corev1.GetResourceDialConfigurationRequest{Id: &id}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The dial configuration was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("resource_id", res.Msg.ResourceDialConfiguration.ResourceId)
	d.Set("dial_method", res.Msg.ResourceDialConfiguration.DialMethod)
	d.Set("dial_target", res.Msg.ResourceDialConfiguration.DialTarget)

	d.SetId(res.Msg.ResourceDialConfiguration.Id)

	return diags
}

func resourceDialConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	fieldsThatCanChange := []string{"dial_method", "dial_target"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	dialMethod := d.Get("dial_method").(string)
	dialTarget := d.Get("dial_target").(string)

	_, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResourceDialConfiguration(ctx, connect.NewRequest(&corev1.UpdateResourceDialConfigurationRequest{
		ResourceDialConfiguration: &corev1.ResourceDialConfiguration{
			Id:         d.Id(),
			ResourceId: d.Get("resource_id").(string),
			DialMethod: dialMethod,
			DialTarget: dialTarget,
		},
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDialConfigurationRead(ctx, d, meta)
}

func resourceDialConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceDialConfigurationId := d.Id()

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceDialConfiguration(ctx, connect.NewRequest(&corev1.DeleteResourceDialConfigurationRequest{Id: resourceDialConfigurationId}))
	if err != nil {
		tflog.Warn(ctx, err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
