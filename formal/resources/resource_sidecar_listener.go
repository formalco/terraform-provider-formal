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

func ResourceSidecarListener() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Sidecar Listener with Formal.",
		CreateContext: resourceSidecarListenerCreate,
		ReadContext:   resourceSidecarListenerRead,
		UpdateContext: resourceSidecarListenerUpdate,
		DeleteContext: resourceSidecarListenerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the listener.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "The port your listener is listening on.",
				Type:        schema.TypeInt,
				Required:    true,
			},
		},
	}
}

func resourceSidecarListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	portInt, ok := d.Get("port").(int)
	if !ok {
		return diag.FromErr(fmt.Errorf("error reading port"))
	}

	res, err := c.Grpc.Sdk.ListenersServiceClient.CreateSidecarListener(ctx, connect.NewRequest(&corev1.CreateSidecarListenerRequest{
		Port: int32(portInt),
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Id)

	resourceSidecarListenerRead(ctx, d, meta)

	return diags
}

func resourceSidecarListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarListenerId := d.Id()

	res, err := c.Grpc.Sdk.ListenersServiceClient.GetSidecarListener(ctx, connect.NewRequest(&corev1.GetSidecarListenerRequest{Id: sidecarListenerId}))
	if err != nil {
		return diag.FromErr(err)
	}
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			// Sidecar listener was deleted
			tflog.Warn(ctx, "The sidecar listener was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.Id)
	d.Set("port", res.Msg.Port)
	d.SetId(res.Msg.Id)

	return diags
}

func resourceSidecarListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	sidecarListenerId := d.Id()

	// Only enable updates to these fields, err otherwise

	fieldsThatCanChange := []string{"port"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	port := d.Get("port").(int32)

	req := connect.NewRequest(&corev1.UpdateSidecarListenerRequest{
		Id:   sidecarListenerId,
		Port: port,
	})
	_, err := c.Grpc.Sdk.ListenersServiceClient.UpdateSidecarListener(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceSidecarListenerRead(ctx, d, meta)

	return diags
}

func resourceSidecarListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	sidecarListenerId := d.Id()
	req := connect.NewRequest(&corev1.DeleteSidecarListenerRequest{
		Id: sidecarListenerId,
	})

	_, err := c.Grpc.Sdk.ListenersServiceClient.DeleteSidecarListener(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
