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
	"google.golang.org/protobuf/types/known/durationpb"

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
			"otel_endpoint_hostname": {
				// This description is used by the documentation generator and the language server.
				Description: "The OpenTelemetry endpoint hostname for this Connector. Defaults to 'localhost'.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "localhost",
			},
			"otel_endpoint_port": {
				// This description is used by the documentation generator and the language server.
				Description: "The OpenTelemetry endpoint port for this Connector. Defaults to 4317.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     4317,
			},
			"resources_health_checks_frequency_seconds": {
				// This description is used by the documentation generator and the language server.
				Description: "The frequency in seconds for resource health checks. Must be between 10 and 3600 seconds. Defaults to 60.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     60,
			},
		},
	}
}

func resourceConnectorConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	otelHostname := d.Get("otel_endpoint_hostname").(string)
	otelPort := int32(d.Get("otel_endpoint_port").(int))
	resourcesHealthChecksFrequencySeconds := int32(d.Get("resources_health_checks_frequency_seconds").(int))

	req := &corev1.CreateConnectorConfigurationRequest{
		ConnectorId:                    d.Get("connector_id").(string),
		LogLevel:                       d.Get("log_level").(string),
		OtelEndpointHostname:           &otelHostname,
		OtelEndpointPort:               &otelPort,
		ResourcesHealthChecksFrequency: durationpb.New(time.Duration(resourcesHealthChecksFrequencySeconds) * time.Second),
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
	d.Set("otel_endpoint_hostname", res.Msg.ConnectorConfiguration.OtelEndpointHostname)
	d.Set("otel_endpoint_port", res.Msg.ConnectorConfiguration.OtelEndpointPort)
	d.Set("resources_health_checks_frequency_seconds", int(res.Msg.ConnectorConfiguration.ResourcesHealthChecksFrequency.AsDuration().Seconds()))

	d.SetId(res.Msg.ConnectorConfiguration.Id)

	return diags
}

func resourceConnectorConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	c := meta.(*clients.Clients)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	connectorConfigurationId := d.Id()

	fieldsThatCanChange := []string{"log_level", "otel_endpoint_hostname", "otel_endpoint_port", "resources_health_checks_frequency_seconds"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.FromErr(errors.New(err))
	}

	logLevel := d.Get("log_level").(string)
	otelHostname := d.Get("otel_endpoint_hostname").(string)
	otelPort := int32(d.Get("otel_endpoint_port").(int))
	resourcesHealthChecksFrequencySeconds := int32(d.Get("resources_health_checks_frequency_seconds").(int))

	req := connect.NewRequest(&corev1.UpdateConnectorConfigurationRequest{
		Id:                             connectorConfigurationId,
		LogLevel:                       &logLevel,
		OtelEndpointHostname:           &otelHostname,
		OtelEndpointPort:               &otelPort,
		ResourcesHealthChecksFrequency: durationpb.New(time.Duration(resourcesHealthChecksFrequencySeconds) * time.Second),
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
