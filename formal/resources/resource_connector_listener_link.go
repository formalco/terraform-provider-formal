package resource

import (
	"context"
	"strings"
	"time"

	corev1 "buf.build/gen/go/formal/core/protocolbuffers/go/core/v1"
	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceConnectorListenerLink() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:        "Registering a Connector Listener Link with Formal.",
		DeprecationMessage: "This resource is deprecated. Declare the connector ID on the connector listener resource instead.",
		CreateContext:      resourceConnectorListenerLinkCreate,
		ReadContext:        resourceConnectorListenerLinkRead,
		UpdateContext:      resourceConnectorListenerLinkUpdate,
		DeleteContext:      resourceConnectorListenerLinkDelete,
		SchemaVersion:      1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of this connector listener link.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector_listener_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Connector Listener you want to link to a connector.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"connector_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Connector Listener you want to link to a connector.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"termination_protection": {
				// This description is used by the documentation generator and the language server.
				Description: "If set to true, this connector cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceConnectorListenerLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	req := &corev1.CreateConnectorListenerLinkRequest{
		ConnectorListenerId:   d.Get("connector_listener_id").(string),
		ConnectorId:           d.Get("connector_id").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorListenerLink(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ConnectorListenerLink.Id)

	resourceConnectorListenerLinkRead(ctx, d, meta)

	return diags
}

func resourceConnectorListenerLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorListenerLinkId := d.Id()

	req := connect.NewRequest(&corev1.GetConnectorListenerLinkRequest{
		Id: connectorListenerLinkId,
	})

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorListenerLink(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector listener was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.ConnectorListenerLink.Id)
	d.Set("connector_listener_id", res.Msg.ConnectorListenerLink.Listener.Id)
	d.Set("connector_id", res.Msg.ConnectorListenerLink.Connector.Id)
	d.Set("termination_protection", res.Msg.ConnectorListenerLink.TerminationProtection)

	d.SetId(res.Msg.ConnectorListenerLink.Id)

	return diags
}

func resourceConnectorListenerLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorListenerLinkId := d.Id()

	fieldsThatCanChange := []string{"termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	terminationProtection := d.Get("termination_protection").(bool)

	req := connect.NewRequest(&corev1.UpdateConnectorListenerLinkRequest{
		Id:                    connectorListenerLinkId,
		TerminationProtection: &terminationProtection,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnectorListenerLink(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceConnectorListenerLinkRead(ctx, d, meta)

	return diags
}

func resourceConnectorListenerLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorListenerLinkId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Connector listener link cannot be deleted because termination_protection is set to true")
	}

	req := connect.NewRequest(&corev1.DeleteConnectorListenerLinkRequest{
		Id: connectorListenerLinkId,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorListenerLink(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
