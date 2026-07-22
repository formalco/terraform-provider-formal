package resource

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
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
				Description: "The ID of this Connector.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				// This description is used by the documentation generator and the language server.
				Description: "Friendly name for this Connector.",
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
				Description: "If set to true, this Connector cannot be deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"space_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Space to create the Connector in.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
		},
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorReq := &corev1.CreateConnectorRequest{
		Name:                  d.Get("name").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}
	if d.Get("space_id").(string) != "" {
		spaceId := d.Get("space_id").(string)
		connectorReq.SpaceId = &spaceId
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnector(ctx, connectorReq)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Connector.Id)

	resourceConnectorRead(ctx, d, meta)

	return diags
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorId := d.Id()

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnector(ctx, &corev1.GetConnectorRequest{Id: connectorId})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	resApiKey, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorApiKey(ctx, &corev1.GetConnectorApiKeyRequest{Id: connectorId})
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("id", res.Connector.Id)
	d.Set("name", res.Connector.Name)
	d.Set("api_key", resApiKey.Secret)
	d.Set("termination_protection", res.Connector.TerminationProtection)
	if res.Connector.Space != nil {
		d.Set("space_id", res.Connector.Space.Id)
	}
	d.SetId(res.Connector.Id)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorId := d.Id()

	fieldsThatCanChange := []string{"name", "termination_protection", "space_id"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	terminationProtection := d.Get("termination_protection").(bool)
	spaceId := d.Get("space_id").(string)

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnector(ctx, &corev1.UpdateConnectorRequest{
		Id:                    connectorId,
		TerminationProtection: &terminationProtection,
		SpaceId:               &spaceId,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	resourceConnectorRead(ctx, d, meta)

	return diags
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorId := d.Id()

	terminationProtection := d.Get("termination_protection").(bool)
	if terminationProtection {
		return diag.Errorf("Connector cannot be deleted because termination_protection is set to true")
	}

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnector(ctx, &corev1.DeleteConnectorRequest{Id: connectorId})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
