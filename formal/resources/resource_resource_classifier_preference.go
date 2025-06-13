package resource

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceResourceClassifierPreference() *schema.Resource {
	return &schema.Resource{
		Description:   "A Resource Classifier Preference is a preference for a resource classifier.",
		CreateContext: resourceResourceClassifierPreferenceCreate,
		ReadContext:   resourceResourceClassifierPreferenceRead,
		UpdateContext: resourceResourceClassifierPreferenceUpdate,
		DeleteContext: resourceResourceClassifierPreferenceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the Resource Classifier Preference.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				Description: "The ID of the Resource.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"preference": {
				Description: "The preference. Supported values are `nlp`, `llm`, `both`, and `none`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"created_at": {
				Description: "The timestamp of the Resource Classifier Preference creation.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"updated_at": {
				Description: "The timestamp of the Resource Classifier Preference update.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func resourceResourceClassifierPreferenceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceId := d.Get("resource_id").(string)
	preference := d.Get("preference").(string)

	msg := &corev1.CreateResourceClassifierPreferenceRequest{
		ResourceId: resourceId,
		Preference: preference,
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err = v.Validate(msg); err != nil {
		return diag.FromErr(err)
	}

	response, err := c.Grpc.Sdk.ResourceServiceClient.CreateResourceClassifierPreference(ctx, connect.NewRequest(msg))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.Msg.ResourceClassifierPreference.Id)
	d.Set("resource_id", response.Msg.ResourceClassifierPreference.ResourceId)
	d.Set("preference", response.Msg.ResourceClassifierPreference.Preference)
	d.Set("created_at", response.Msg.ResourceClassifierPreference.CreatedAt)
	d.Set("updated_at", response.Msg.ResourceClassifierPreference.UpdatedAt)

	return diags
}

func resourceResourceClassifierPreferenceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceId := d.Get("resource_id").(string)

	response, err := c.Grpc.Sdk.ResourceServiceClient.GetResourceClassifierPreference(ctx, connect.NewRequest(&corev1.GetResourceClassifierPreferenceRequest{ResourceId: resourceId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("id", response.Msg.ResourceClassifierPreference.Id)
	d.Set("resource_id", response.Msg.ResourceClassifierPreference.ResourceId)
	d.Set("preference", response.Msg.ResourceClassifierPreference.Preference)
	d.Set("created_at", response.Msg.ResourceClassifierPreference.CreatedAt)
	d.Set("updated_at", response.Msg.ResourceClassifierPreference.UpdatedAt)

	return diags
}

func resourceResourceClassifierPreferenceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceClassifierPreferenceId := d.Id()
	preference := d.Get("preference").(string)

	msg := &corev1.UpdateResourceClassifierPreferenceRequest{
		Id:         resourceClassifierPreferenceId,
		Preference: preference,
	}

	v, err := protovalidate.New()
	if err != nil {
		return diag.FromErr(err)
	}
	if err = v.Validate(msg); err != nil {
		return diag.FromErr(err)
	}

	response, err := c.Grpc.Sdk.ResourceServiceClient.UpdateResourceClassifierPreference(ctx, connect.NewRequest(msg))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("id", response.Msg.ResourceClassifierPreference.Id)
	d.Set("resource_id", response.Msg.ResourceClassifierPreference.ResourceId)
	d.Set("preference", response.Msg.ResourceClassifierPreference.Preference)
	d.Set("created_at", response.Msg.ResourceClassifierPreference.CreatedAt)
	d.Set("updated_at", response.Msg.ResourceClassifierPreference.UpdatedAt)

	return diags
}

func resourceResourceClassifierPreferenceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	resourceClassifierPreferenceId := d.Id()

	_, err := c.Grpc.Sdk.ResourceServiceClient.DeleteResourceClassifierPreference(ctx, connect.NewRequest(&corev1.DeleteResourceClassifierPreferenceRequest{Id: resourceClassifierPreferenceId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
