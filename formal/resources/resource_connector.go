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

func ResourceConnector() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Connector with Formal.",
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		UpdateContext: resourceConnectorUpdate,
		DeleteContext: resourceConnectorDelete,
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
				Description: "The ID of this connector.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this connector.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"api_key": {
				// This description is used by the documentation generator and the language server.
				Description: "Api key for the deployed Connector.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
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

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	sidecarReq := &corev1.CreateConnectorRequest{
		Name:                  d.Get("name").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnector(ctx, connect.NewRequest(sidecarReq))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.Connector.Id)

	resourceConnectorRead(ctx, d, meta)

	return diags
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorId := d.Id()

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnector(ctx, connect.NewRequest(&corev1.GetConnectorRequest{Id: connectorId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	resApiKey, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorApiKey(ctx, connect.NewRequest(&corev1.GetConnectorApiKeyRequest{Id: connectorId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.Connector.Id)
	d.Set("name", res.Msg.Connector.Name)
	d.Set("api_key", resApiKey.Msg.Secret)
	d.Set("termination_protection", res.Msg.Connector.TerminationProtection)

	d.SetId(res.Msg.Connector.Id)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorId := d.Id()

	fieldsThatCanChange := []string{"name", "termination_protection"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	terminationProtection := d.Get("termination_protection").(bool)

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnector(ctx, connect.NewRequest(&corev1.UpdateConnectorRequest{
		Id:                    connectorId,
		TerminationProtection: &terminationProtection,
	}))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceConnectorRead(ctx, d, meta)

	return diags
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Connector cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnector(ctx, connect.NewRequest(&corev1.DeleteConnectorRequest{Id: connectorId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
