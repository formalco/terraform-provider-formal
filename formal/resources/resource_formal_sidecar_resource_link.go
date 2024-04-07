package resource

import (
	"context"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/formalco/terraform-provider-formal/formal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSidecarResourceLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Linking a Resource to a Sidecar in Formal.",

		CreateContext: resourceSidecarResourceLinkCreate,
		ReadContext:   resourceSidecarResourceLinkRead,
		DeleteContext: resourceSidecarResourceLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "Resource ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"resource_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Resource ID to be linked.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"sidecar_id": {
				// This description is used by the documentation generator and the language server.
				Description: "Sidecar ID that should be linked.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "Port.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this Sidecar Datastore Link cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
		},
	}
}

func resourceSidecarResourceLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	res, err := c.Grpc.Sdk.SidecarServiceClient.CreateSidecarResourceLink(ctx, connect.NewRequest(&corev1.CreateSidecarResourceLinkRequest{
		ResourceId: d.Get("resource_id").(string),
		SidecarId:  d.Get("sidecar_id").(string),
		Port:       int32(d.Get("port").(int)),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Link.Id)

	resourceSidecarResourceLinkRead(ctx, d, meta)
	return diags
}

func resourceSidecarResourceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarResourceLinkId := d.Id()

	res, err := c.Grpc.Sdk.SidecarServiceClient.GetSidecarResourceLink(ctx, connect.NewRequest(&corev1.GetSidecarResourceLinkRequest{Id: sidecarResourceLinkId}))
	if err != nil {
		return diag.FromErr(err)
	}

	// Should map to all fields of
	d.Set("id", res.Msg.Link.Id)
	d.Set("resource", res.Msg.Link.Resource)
	d.Set("sidecar", res.Msg.Link.Sidecar)
	d.Set("port", res.Msg.Link.Port)

	d.SetId(res.Msg.Link.Id)

	return diags
}

func resourceSidecarResourceLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Sidecar Resource Link cannot be deleted because termination_protection is set to true")
	}

	sidecarResourceLinkId := d.Id()

	_, err := c.Grpc.Sdk.SidecarServiceClient.DeleteSidecarResourceLink(ctx, connect.NewRequest(&corev1.DeleteSidecarResourceLinkRequest{Id: sidecarResourceLinkId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}