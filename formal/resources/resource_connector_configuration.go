package resource

import (
	"context"
	"errors"
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

func ResourceConnectorConfiguration() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Registering a Connector Configuration with Formal.",
		CreateContext: resourceConnectorConfigurationCreate,
		ReadContext:   resourceConnectorConfigurationRead,
		UpdateContext: resourceConnectorConfigurationUpdate,
		DeleteContext: resourceConnectorConfigurationDelete,
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
				Description: "The ID of this Connector Configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector_id": {
				// This description is used by the documentation generator and the language server.
				Description: "The ID of the Connector this configuration is linked to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"log_level": {
				// This description is used by the documentation generator and the language server.
				Description: "The log level to be configured for this Connector.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "info",
			},
			"health_check_port": {
				// This description is used by the documentation generator and the language server.
				Description: "The port to be used for this Connector's health check.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     8080,
			},
		},
	}
}

func resourceConnectorConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	req := &corev1.CreateConnectorConfigurationRequest{
		ConnectorId:     d.Get("connector_id").(string),
		LogLevel:        d.Get("log_level").(string),
		HealthCheckPort: int32(d.Get("health_check_port").(int)),
	}

	res, err := c.Grpc.Sdk.ConnectorServiceClient.CreateConnectorConfiguration(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.ConnectorConfiguration.Id)
	resourceConnectorConfigurationRead(ctx, d, meta)
	return diags
}

func resourceConnectorConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	var diags diag.Diagnostics

	connectorConfigurationId := d.Id()

	req := connect.NewRequest(&corev1.GetConnectorConfigurationRequest{
		Id: connectorConfigurationId,
	})

	res, err := c.Grpc.Sdk.ConnectorServiceClient.GetConnectorConfiguration(ctx, req)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Connector configuration was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.ConnectorConfiguration.Id)
	d.Set("connector_id", res.Msg.ConnectorConfiguration.ConnectorId)
	d.Set("log_level", res.Msg.ConnectorConfiguration.LogLevel)
	d.Set("health_check_port", res.Msg.ConnectorConfiguration.HealthCheckPort)

	d.SetId(res.Msg.ConnectorConfiguration.Id)

	return diags
}

func resourceConnectorConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorConfigurationId := d.Id()

	fieldsThatCanChange := []string{"log_level", "health_check_port"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.FromErr(errors.New(err))
	}

	logLevel := d.Get("log_level").(string)
	healthCheckPort := d.Get("health_check_port").(int32)

	req := connect.NewRequest(&corev1.UpdateConnectorConfigurationRequest{
		Id:              connectorConfigurationId,
		LogLevel:        &logLevel,
		HealthCheckPort: &healthCheckPort,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.UpdateConnectorConfiguration(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	resourceConnectorConfigurationRead(ctx, d, meta)

	return diags
}

func resourceConnectorConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorConfigurationId := d.Id()

	req := connect.NewRequest(&corev1.DeleteConnectorConfigurationRequest{
		Id: connectorConfigurationId,
	})

	_, err := c.Grpc.Sdk.ConnectorServiceClient.DeleteConnectorConfiguration(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}
