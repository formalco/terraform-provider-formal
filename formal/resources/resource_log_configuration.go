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

func ResourceLogConfiguration() *schema.Resource {
	return &schema.Resource{
		Description:   "Managing Log Configuration with Formal.",
		CreateContext: resourceLogConfigurationCreate,
		ReadContext:   resourceLogConfigurationRead,
		UpdateContext: resourceLogConfigurationUpdate,
		DeleteContext: resourceLogConfigurationDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this log configuration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of this log configuration.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"connector_id": {
				Description: "The ID of the connector this configuration applies to.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"resource_id": {
				Description: "The ID of the resource this configuration applies to.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"space_id": {
				Description: "The ID of the space this configuration applies to.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"request_payload_max_size": {
				Description: "Maximum size of request payloads to log.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"response_payload_max_size": {
				Description: "Maximum size of response payloads to log.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"encrypt_request_payload": {
				Description: "Whether to encrypt request payloads.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"encrypt_response_payload": {
				Description: "Whether to encrypt response payloads.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"encryption_key_id": {
				Description: "ID of the encryption key to use for encryption.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"strip_values_from_sql_queries": {
				Description: "Whether to obfuscate SQL queries in logs.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"encrypt_values_from_sql_queries": {
				Description: "Whether to encrypt SQL queries in logs.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"encrypt_exec_stream_logs": {
				Description: "Whether to encrypt exec stream logs.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"created_at": {
				Description: "When the log configuration was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "Last update time.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceLogConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	req := &corev1.CreateLogConfigurationRequest{
		Name:                        d.Get("name").(string),
		LogConfigurationLevel:       corev1.LogConfigurationLevel_LOG_CONFIGURATION_LEVEL_ACCOUNT,
		RequestPayloadMaxSize:       int64(d.Get("request_payload_max_size").(int)),
		ResponsePayloadMaxSize:      int64(d.Get("response_payload_max_size").(int)),
		EncryptRequestPayload:       d.Get("encrypt_request_payload").(bool),
		EncryptResponsePayload:      d.Get("encrypt_response_payload").(bool),
		StripValuesFromSqlQueries:   d.Get("strip_values_from_sql_queries").(bool),
		EncryptValuesFromSqlQueries: d.Get("encrypt_values_from_sql_queries").(bool),
		EncryptExecStreamLogs:       d.Get("encrypt_exec_stream_logs").(bool),
		EncryptionKeyId:             d.Get("encryption_key_id").(string),
	}

	spaceId, spaceIdOk := d.GetOk("space_id")
	connectorId, connectorIdOk := d.GetOk("connector_id")
	resourceId, resourceIdOk := d.GetOk("resource_id")

	if (spaceIdOk && connectorIdOk) || (spaceIdOk && resourceIdOk) || (connectorIdOk && resourceIdOk) {
		return diag.Errorf("You can only specify one of space_id, connector_id, or resource_id")
	}

	if spaceIdOk {
		spaceId := spaceId.(string)
		req.SpaceId = &spaceId
		req.LogConfigurationLevel = corev1.LogConfigurationLevel_LOG_CONFIGURATION_LEVEL_SPACE
	}

	if connectorIdOk {
		connectorId := connectorId.(string)
		req.ConnectorId = &connectorId
		req.LogConfigurationLevel = corev1.LogConfigurationLevel_LOG_CONFIGURATION_LEVEL_CONNECTOR
	}

	if resourceIdOk {
		resourceId := resourceId.(string)
		req.ResourceId = &resourceId
		req.LogConfigurationLevel = corev1.LogConfigurationLevel_LOG_CONFIGURATION_LEVEL_RESOURCE
	}

	res, err := c.Grpc.Sdk.LogsServiceClient.CreateLogConfiguration(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(res.Msg.LogConfiguration.Id)
	return resourceLogConfigurationRead(ctx, d, meta)
}

func resourceLogConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	configId := d.Id()

	res, err := c.Grpc.Sdk.LogsServiceClient.GetLogConfiguration(ctx, connect.NewRequest(&corev1.GetLogConfigurationRequest{Id: configId}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Log Configuration was not found, which means it may have been deleted without using this Terraform config.", map[string]interface{}{"err": err})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	d.Set("id", res.Msg.LogConfiguration.Id)
	d.Set("name", res.Msg.LogConfiguration.Name)
	d.Set("log_configuration_level", res.Msg.LogConfiguration.LogConfigurationLevel.String())
	d.Set("connector_id", res.Msg.LogConfiguration.ConnectorId)
	d.Set("resource_id", res.Msg.LogConfiguration.ResourceId)
	d.Set("space_id", res.Msg.LogConfiguration.SpaceId)
	d.Set("request_payload_max_size", res.Msg.LogConfiguration.RequestPayloadMaxSize)
	d.Set("response_payload_max_size", res.Msg.LogConfiguration.ResponsePayloadMaxSize)
	d.Set("encrypt_request_payload", res.Msg.LogConfiguration.EncryptRequestPayload)
	d.Set("encrypt_response_payload", res.Msg.LogConfiguration.EncryptResponsePayload)
	d.Set("strip_values_from_sql_queries", res.Msg.LogConfiguration.StripValuesFromSqlQueries)
	d.Set("encrypt_values_from_sql_queries", res.Msg.LogConfiguration.EncryptValuesFromSqlQueries)
	d.Set("encrypt_exec_stream_logs", res.Msg.LogConfiguration.EncryptExecStreamLogs)
	d.Set("encryption_key_id", res.Msg.LogConfiguration.EncryptionKeyId)
	d.Set("created_at", res.Msg.LogConfiguration.CreatedAt.AsTime().String())
	d.Set("updated_at", res.Msg.LogConfiguration.UpdatedAt.AsTime().String())

	d.SetId(res.Msg.LogConfiguration.Id)
	return diags
}

func resourceLogConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	configId := d.Id()

	fieldsThatCanChange := []string{
		"name",
		"request_payload_max_size", "response_payload_max_size",
		"encrypt_request_payload", "encrypt_response_payload",
		"strip_values_from_sql_queries", "encrypt_values_from_sql_queries",
		"encrypt_exec_stream_logs", "encryption_key_id",
	}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		err := fmt.Sprintf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
		return diag.Errorf(err)
	}

	req := &corev1.UpdateLogConfigurationRequest{
		Id: configId,
	}

	if d.HasChange("request_payload_max_size") {
		size := int64(d.Get("request_payload_max_size").(int))
		req.RequestPayloadMaxSize = &size
	}
	if d.HasChange("response_payload_max_size") {
		size := int64(d.Get("response_payload_max_size").(int))
		req.ResponsePayloadMaxSize = &size
	}
	if d.HasChange("encrypt_request_payload") {
		encrypt := d.Get("encrypt_request_payload").(bool)
		req.EncryptRequestPayload = &encrypt
	}
	if d.HasChange("encrypt_response_payload") {
		encrypt := d.Get("encrypt_response_payload").(bool)
		req.EncryptResponsePayload = &encrypt
	}
	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}
	if d.HasChange("strip_values_from_sql_queries") {
		strip := d.Get("strip_values_from_sql_queries").(bool)
		req.StripValuesFromSqlQueries = &strip
	}
	if d.HasChange("encrypt_values_from_sql_queries") {
		encrypt := d.Get("encrypt_values_from_sql_queries").(bool)
		req.EncryptValuesFromSqlQueries = &encrypt
	}
	if d.HasChange("encrypt_exec_stream_logs") {
		val := d.Get("encrypt_exec_stream_logs").(bool)
		req.EncryptExecStreamLogs = &val
	}
	if d.HasChange("encryption_key_id") {
		keyId := d.Get("encryption_key_id").(string)
		req.EncryptionKeyId = &keyId
	}

	_, err := c.Grpc.Sdk.LogsServiceClient.UpdateLogConfiguration(ctx, connect.NewRequest(req))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceLogConfigurationRead(ctx, d, meta)
}

func resourceLogConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	var diags diag.Diagnostics

	configId := d.Id()

	_, err := c.Grpc.Sdk.LogsServiceClient.DeleteLogConfiguration(ctx, connect.NewRequest(&corev1.DeleteLogConfigurationRequest{Id: configId}))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}
