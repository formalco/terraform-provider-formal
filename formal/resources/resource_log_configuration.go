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
	"google.golang.org/protobuf/types/known/durationpb"

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
				Description: "The name of this log configuration.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"encryption_key_id": {
				Description: "The ID of the encryption key to use for this log configuration.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"scope": {
				Description: "The scope configuration for this log configuration.",
				Type:        schema.TypeSet,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "The type of scope (resource, connector, space, org).",
							Type:        schema.TypeString,
							Required:    true,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								valid := []string{"resource", "connector", "space", "org"}
								for _, validVal := range valid {
									if v == validVal {
										return
									}
								}
								errs = append(errs, fmt.Errorf("%q must be one of %v", key, valid))
								return
							},
						},
						"resource_id": {
							Description: "The ID of the resource (required when type is resource).",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"connector_id": {
							Description: "The ID of the connector (required when type is connector).",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"space_id": {
							Description: "The ID of the space (required when type is space).",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"request": {
				Description: "Request logging configuration.",
				Type:        schema.TypeSet,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encrypt": {
							Description: "Whether to encrypt request payloads.",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"max_payload_size": {
							Description: "Maximum size of request payloads to log.",
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     -1,
						},
						"sql": {
							Description: "SQL logging configuration for requests.",
							Type:        schema.TypeSet,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"strip_values": {
										Description: "Whether to obfuscate SQL queries in logs.",
										Type:        schema.TypeBool,
										Required:    true,
									},
									"encrypt": {
										Description: "Whether to encrypt SQL queries in logs.",
										Type:        schema.TypeBool,
										Required:    true,
									},
								},
							},
						},
						"policy_eval_input_retention": {
							Description: "Duration to retain policy evaluation inputs for requests. Valid values: 1d, 2d, 3d, 7d, 14d, 21d, 30d.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"response": {
				Description: "Response logging configuration.",
				Type:        schema.TypeSet,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encrypt": {
							Description: "Whether to encrypt response payloads.",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"max_payload_size": {
							Description: "Maximum size of response payloads to log.",
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     -1,
						},
						"policy_eval_input_retention": {
							Description: "Duration to retain policy evaluation inputs for responses. Valid values: 1d, 2d, 3d, 7d, 14d, 21d, 30d.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"stream": {
				Description: "Stream logging configuration.",
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encrypt": {
							Description: "Whether to encrypt stream data.",
							Type:        schema.TypeBool,
							Required:    true,
						},
					},
				},
			},
			"session": {
				Description: "Session logging configuration.",
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy_eval_input_retention": {
							Description: "Duration to retain policy evaluation inputs for sessions. Valid values: 1d, 2d, 3d, 7d, 14d, 21d, 30d.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
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

// parseDuration converts a duration string in format "%dd" (e.g., "1d", "7d", "30d") to a protobuf Duration.
// Only accepts values: 1d, 2d, 3d, 7d, 14d, 21d, 30d
// Returns nil if the input string is empty.
func parseDuration(durationStr string) (*durationpb.Duration, error) {
	if durationStr == "" {
		return nil, nil
	}

	// Valid duration values in days
	validDurations := map[string]int{
		"1d":  1,
		"2d":  2,
		"3d":  3,
		"7d":  7,
		"14d": 14,
		"21d": 21,
		"30d": 30,
	}

	days, ok := validDurations[durationStr]
	if !ok {
		return nil, fmt.Errorf("invalid duration '%s': must be one of 1d, 2d, 3d, 7d, 14d, 21d, 30d", durationStr)
	}

	// Convert days to time.Duration (days * 24 hours)
	duration := time.Duration(days) * 24 * time.Hour
	return durationpb.New(duration), nil
}

// formatDuration converts a protobuf Duration to a string representation in "%dd" format
func formatDuration(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}

	duration := d.AsDuration()
	days := int(duration.Hours() / 24)

	// Return in "%dd" format
	return fmt.Sprintf("%dd", days)
}

func resourceLogConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)

	req := &corev1.CreateLogConfigurationRequest{
		Name: d.Get("name").(string),
	}

	// Handle encryption_key_id if provided
	if encryptionKeyId, ok := d.GetOk("encryption_key_id"); ok {
		keyId := encryptionKeyId.(string)
		req.EncryptionKeyId = &keyId
	}

	// Handle scope
	if scopeSet := d.Get("scope").(*schema.Set); scopeSet.Len() > 0 {
		scopeData := scopeSet.List()[0].(map[string]interface{})
		scopeType := scopeData["type"].(string)
		req.Scope = &corev1.LogConfigurationScope{}

		switch scopeType {
		case "resource":
			req.Scope.Scope = corev1.LogConfigurationScopeType_LOG_CONFIGURATION_SCOPE_TYPE_RESOURCE
			if resourceId, ok := scopeData["resource_id"].(string); ok && resourceId != "" {
				req.Scope.ResourceId = &resourceId
			} else {
				return diag.Errorf("resource_id is required when scope type is 'resource'")
			}
		case "connector":
			req.Scope.Scope = corev1.LogConfigurationScopeType_LOG_CONFIGURATION_SCOPE_TYPE_CONNECTOR
			if connectorId, ok := scopeData["connector_id"].(string); ok && connectorId != "" {
				req.Scope.ConnectorId = &connectorId
			} else {
				return diag.Errorf("connector_id is required when scope type is 'connector'")
			}
		case "space":
			req.Scope.Scope = corev1.LogConfigurationScopeType_LOG_CONFIGURATION_SCOPE_TYPE_SPACE
			if spaceId, ok := scopeData["space_id"].(string); ok && spaceId != "" {
				req.Scope.SpaceId = &spaceId
			} else {
				return diag.Errorf("space_id is required when scope type is 'space'")
			}
		case "org":
			req.Scope.Scope = corev1.LogConfigurationScopeType_LOG_CONFIGURATION_SCOPE_TYPE_ORG
		default:
			return diag.Errorf("invalid scope type: %s", scopeType)
		}
	}

	// Handle request
	if requestSet := d.Get("request").(*schema.Set); requestSet.Len() > 0 {
		requestData := requestSet.List()[0].(map[string]interface{})
		req.Request = &corev1.LogConfigurationRequest{
			Encrypt:        requestData["encrypt"].(bool),
			MaxPayloadSize: int64(requestData["max_payload_size"].(int)),
		}

		// Handle SQL config if present
		if sqlSetRaw, ok := requestData["sql"]; ok {
			if sqlSet := sqlSetRaw.(*schema.Set); sqlSet != nil && sqlSet.Len() > 0 {
				sqlData := sqlSet.List()[0].(map[string]interface{})
				req.Request.Sql = &corev1.LogConfigurationSql{
					StripValues: sqlData["strip_values"].(bool),
					Encrypt:     sqlData["encrypt"].(bool),
				}
			}
		}

		// Handle policy_eval_input_retention if present
		if retentionStr, ok := requestData["policy_eval_input_retention"].(string); ok && retentionStr != "" {
			retention, err := parseDuration(retentionStr)
			if err != nil {
				return diag.FromErr(fmt.Errorf("invalid request policy_eval_input_retention: %w", err))
			}
			req.Request.PolicyEvalInputRetention = retention
		}
	}

	// Handle response
	if responseSet := d.Get("response").(*schema.Set); responseSet.Len() > 0 {
		responseData := responseSet.List()[0].(map[string]interface{})
		req.Response = &corev1.LogConfigurationResponse{
			Encrypt:        responseData["encrypt"].(bool),
			MaxPayloadSize: int64(responseData["max_payload_size"].(int)),
		}

		// Handle policy_eval_input_retention if present
		if retentionStr, ok := responseData["policy_eval_input_retention"].(string); ok && retentionStr != "" {
			retention, err := parseDuration(retentionStr)
			if err != nil {
				return diag.FromErr(fmt.Errorf("invalid response policy_eval_input_retention: %w", err))
			}
			req.Response.PolicyEvalInputRetention = retention
		}
	}

	// Handle stream
	if streamSet := d.Get("stream").(*schema.Set); streamSet.Len() > 0 {
		streamData := streamSet.List()[0].(map[string]interface{})
		req.Stream = &corev1.LogConfigurationStream{
			Encrypt: streamData["encrypt"].(bool),
		}
	}

	// Handle session
	if sessionSet := d.Get("session").(*schema.Set); sessionSet.Len() > 0 {
		sessionData := sessionSet.List()[0].(map[string]interface{})
		req.Session = &corev1.LogConfigurationSession{}

		// Handle policy_eval_input_retention if present
		if retentionStr, ok := sessionData["policy_eval_input_retention"].(string); ok && retentionStr != "" {
			retention, err := parseDuration(retentionStr)
			if err != nil {
				return diag.FromErr(fmt.Errorf("invalid session policy_eval_input_retention: %w", err))
			}
			req.Session.PolicyEvalInputRetention = retention
		}
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

	logConfig := res.Msg.LogConfiguration
	d.Set("id", logConfig.Id)
	d.Set("name", logConfig.Name)
	d.Set("encryption_key_id", logConfig.EncryptionKeyId)

	// Set scope
	if logConfig.Scope != nil {
		scopeData := map[string]interface{}{}
		switch logConfig.Scope.Scope {
		case corev1.LogConfigurationScopeType_LOG_CONFIGURATION_SCOPE_TYPE_RESOURCE:
			scopeData["type"] = "resource"
			if logConfig.Scope.ResourceId != nil {
				scopeData["resource_id"] = *logConfig.Scope.ResourceId
			}
		case corev1.LogConfigurationScopeType_LOG_CONFIGURATION_SCOPE_TYPE_CONNECTOR:
			scopeData["type"] = "connector"
			if logConfig.Scope.ConnectorId != nil {
				scopeData["connector_id"] = *logConfig.Scope.ConnectorId
			}
		case corev1.LogConfigurationScopeType_LOG_CONFIGURATION_SCOPE_TYPE_SPACE:
			scopeData["type"] = "space"
			if logConfig.Scope.SpaceId != nil {
				scopeData["space_id"] = *logConfig.Scope.SpaceId
			}
		case corev1.LogConfigurationScopeType_LOG_CONFIGURATION_SCOPE_TYPE_ORG:
			scopeData["type"] = "org"
		}
		d.Set("scope", []interface{}{scopeData})
	}

	// Set request
	if logConfig.Request != nil {
		requestData := map[string]interface{}{
			"encrypt":          logConfig.Request.Encrypt,
			"max_payload_size": logConfig.Request.MaxPayloadSize,
		}

		// Set SQL config if present
		if logConfig.Request.Sql != nil {
			sqlData := map[string]interface{}{
				"strip_values": logConfig.Request.Sql.StripValues,
				"encrypt":      logConfig.Request.Sql.Encrypt,
			}
			requestData["sql"] = []interface{}{sqlData}
		}

		// Set policy_eval_input_retention if present
		if logConfig.Request.PolicyEvalInputRetention != nil {
			requestData["policy_eval_input_retention"] = formatDuration(logConfig.Request.PolicyEvalInputRetention)
		}

		d.Set("request", []interface{}{requestData})
	}

	// Set response
	if logConfig.Response != nil {
		responseData := map[string]interface{}{
			"encrypt":          logConfig.Response.Encrypt,
			"max_payload_size": logConfig.Response.MaxPayloadSize,
		}

		// Set policy_eval_input_retention if present
		if logConfig.Response.PolicyEvalInputRetention != nil {
			responseData["policy_eval_input_retention"] = formatDuration(logConfig.Response.PolicyEvalInputRetention)
		}

		d.Set("response", []interface{}{responseData})
	}

	// Set stream
	if logConfig.Stream != nil {
		streamData := map[string]interface{}{
			"encrypt": logConfig.Stream.Encrypt,
		}
		d.Set("stream", []interface{}{streamData})
	}

	// Set session
	if logConfig.Session != nil {
		sessionData := map[string]interface{}{}

		// Set policy_eval_input_retention if present
		if logConfig.Session.PolicyEvalInputRetention != nil {
			sessionData["policy_eval_input_retention"] = formatDuration(logConfig.Session.PolicyEvalInputRetention)
		}

		d.Set("session", []interface{}{sessionData})
	}

	d.Set("created_at", logConfig.CreatedAt.AsTime().String())
	d.Set("updated_at", logConfig.UpdatedAt.AsTime().String())

	d.SetId(logConfig.Id)
	return diags
}

func resourceLogConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*clients.Clients)
	configId := d.Id()

	fieldsThatCanChange := []string{
		"name", "encryption_key_id", "request", "response", "stream", "session",
	}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	req := &corev1.UpdateLogConfigurationRequest{
		Id: configId,
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}

	if d.HasChange("encryption_key_id") {
		encryptionKeyId := d.Get("encryption_key_id").(string)
		req.EncryptionKeyId = &encryptionKeyId
	}

	// Handle request changes
	if d.HasChange("request") {
		if requestSet := d.Get("request").(*schema.Set); requestSet.Len() > 0 {
			requestData := requestSet.List()[0].(map[string]interface{})
			req.Request = &corev1.LogConfigurationRequest{
				Encrypt:        requestData["encrypt"].(bool),
				MaxPayloadSize: int64(requestData["max_payload_size"].(int)),
			}

			// Handle SQL config if present
			if sqlSet := requestData["sql"].(*schema.Set); sqlSet.Len() > 0 {
				sqlData := sqlSet.List()[0].(map[string]interface{})
				req.Request.Sql = &corev1.LogConfigurationSql{
					StripValues: sqlData["strip_values"].(bool),
					Encrypt:     sqlData["encrypt"].(bool),
				}
			}

			// Handle policy_eval_input_retention if present
			if retentionStr, ok := requestData["policy_eval_input_retention"].(string); ok && retentionStr != "" {
				retention, err := parseDuration(retentionStr)
				if err != nil {
					return diag.FromErr(fmt.Errorf("invalid request policy_eval_input_retention: %w", err))
				}
				req.Request.PolicyEvalInputRetention = retention
			}
		}
	}

	// Handle response changes
	if d.HasChange("response") {
		if responseSet := d.Get("response").(*schema.Set); responseSet.Len() > 0 {
			responseData := responseSet.List()[0].(map[string]interface{})
			req.Response = &corev1.LogConfigurationResponse{
				Encrypt:        responseData["encrypt"].(bool),
				MaxPayloadSize: int64(responseData["max_payload_size"].(int)),
			}

			// Handle policy_eval_input_retention if present
			if retentionStr, ok := responseData["policy_eval_input_retention"].(string); ok && retentionStr != "" {
				retention, err := parseDuration(retentionStr)
				if err != nil {
					return diag.FromErr(fmt.Errorf("invalid response policy_eval_input_retention: %w", err))
				}
				req.Response.PolicyEvalInputRetention = retention
			}
		}
	}

	// Handle stream changes
	if d.HasChange("stream") {
		if streamSet := d.Get("stream").(*schema.Set); streamSet.Len() > 0 {
			streamData := streamSet.List()[0].(map[string]interface{})
			req.Stream = &corev1.LogConfigurationStream{
				Encrypt: streamData["encrypt"].(bool),
			}
		}
	}

	// Handle session changes
	if d.HasChange("session") {
		if sessionSet := d.Get("session").(*schema.Set); sessionSet.Len() > 0 {
			sessionData := sessionSet.List()[0].(map[string]interface{})
			req.Session = &corev1.LogConfigurationSession{}

			// Handle policy_eval_input_retention if present
			if retentionStr, ok := sessionData["policy_eval_input_retention"].(string); ok && retentionStr != "" {
				retention, err := parseDuration(retentionStr)
				if err != nil {
					return diag.FromErr(fmt.Errorf("invalid session policy_eval_input_retention: %w", err))
				}
				req.Session.PolicyEvalInputRetention = retention
			}
		}
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
