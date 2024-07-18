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

func ResourceConnectorListener() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Connector Listener with Formal.",
		CreateContext: resourceConnectorListenerCreate,
		ReadContext:   resourceConnectorListenerRead,
		UpdateContext: resourceConnectorListenerUpdate,
		DeleteContext: resourceConnectorListenerDelete,
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of this connector listener.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"port": {
				// This description is used by the documentation generator and the language server.
				Description: "The listening port for this connector listener.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this connector listener cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceConnectorListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	req := &corev1.CreateConnectorListenerRequest{
		Port:                  d.Get("port").(int32),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorListener(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ConnectorListener.Id)

	resourceConnectorListenerRead(ctx, d, meta)

	return diags
}

func resourceConnectorListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorListenerId := d.Id()

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorListener(ctx, connect.NewRequest(&corev1.GetConnectorListenerRequest{Id: connectorListenerId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector listener was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.ConnectorListener.Id)
	d.Set("port", res.Msg.ConnectorListener.Port)
	d.Set("termination_protection", res.Msg.ConnectorListener.TerminationProtection)

	d.SetId(res.Msg.ConnectorListener.Id)

	return diags
}

func resourceConnectorListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorListenerId := d.Id()

	fieldsThatCanChange := []string{"port", "termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	port := d.Get("port").(int32)
	terminationProtection := d.Get("termination_protection").(bool)

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnectorListener(ctx, connect.NewRequest(&corev1.UpdateConnectorListenerRequest{
		Id:                    connectorListenerId,
		Port:                  &port,
		TerminationProtection: &terminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceConnectorListenerRead(ctx, d, meta)

	return diags
}

func resourceConnectorListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorListenerId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Connector listener cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorListener(ctx, connect.NewRequest(&corev1.DeleteConnectorListenerRequest{Id: connectorListenerId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
