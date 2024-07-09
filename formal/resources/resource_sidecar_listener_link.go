package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceSidecarListenerLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Sidecar Listener Link with Formal.",
		CreateContext: resourceSidecarListenerLinkCreate,
		ReadContext:   resourceSidecarListenerLinkRead,
		UpdateContext: resourceSidecarListenerLinkUpdate,
		DeleteContext: resourceSidecarListenerLinkDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the listener link.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"sidecar_listener_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the sidecar listener.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"sidecar_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the sidecar.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceSidecarListenerLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	sidecarListenerId := d.Get("sidecar_listener_id").(string)
	sidecarId := d.Get("sidecar_id").(string)

	res, err := c.Grpc.Sdk.ListenersServiceClient.CreateSidecarListenerLink(ctx, connect.NewRequest(&corev1.CreateSidecarListenerLinkRequest{
		SidecarListenerId: sidecarListenerId,
		SidecarId:         sidecarId,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Id)

	resourceSidecarListenerLinkRead(ctx, d, meta)

	return diags
}

func resourceSidecarListenerLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarListenerLinkId := d.Id()

	res, err := c.Grpc.Sdk.ListenersServiceClient.GetSidecarListenerLink(ctx, connect.NewRequest(&corev1.GetSidecarListenerLinkRequest{Id: sidecarListenerLinkId}))
	if err != nil {
		return diag.FromErr(err)
	}
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Sidecar listener rule was deleted
			tflog.Warn(ctx, "The sidecar listener was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.Id)
	d.Set("sidecar_listener_id", res.Msg.SidecarListenerId)
	d.Set("sidecar_id", res.Msg.SidecarId)
	d.SetId(res.Msg.Id)

	return diags
}

func resourceSidecarListenerLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarListenerLinkId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"sidecar_listener_id", "sidecar_id"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	sidecarListenerId := d.Get("sidecar_listener_id").(string)
	sidecarId := d.Get("sidecar_id").(string)

	req := connect.NewRequest(&corev1.UpdateSidecarListenerLinkRequest{
		Id:                sidecarListenerLinkId,
		SidecarListenerId: sidecarListenerId,
		SidecarId:         sidecarId,
	})
	_, err := c.Grpc.Sdk.ListenersServiceClient.UpdateSidecarListenerLink(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceSidecarListenerLinkRead(ctx, d, meta)

	return diags
}

func resourceSidecarListenerLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	sidecarListenerLinkId := d.Id()
	req := connect.NewRequest(&corev1.DeleteSidecarListenerLinkRequest{
		Id: sidecarListenerLinkId,
	})

	_, err := c.Grpc.Sdk.ListenersServiceClient.DeleteSidecarListenerLink(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
