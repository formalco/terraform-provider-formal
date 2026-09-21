package resource

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
	"github.com/formalco/terraform-provider-formal/formal/clients"
)

func ResourceLogSchema() *schema.Resource {
	pathResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The protobuf field path, starting with `log`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"drop": {
				Description: "Whether to remove the field from the log.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"encrypt": {
				Description: "Whether to encrypt the string field.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"truncate": {
				Description: "Truncates the string field to a maximum UTF-8 byte length.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_size_bytes": {
							Description:  "The maximum field size in bytes.",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"strip_sql": {
				Description: "Whether to remove literal values from a SQL query.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"encrypt_sql": {
				Description: "Whether to encrypt literal values in a SQL query.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
	pathHash := schema.HashResource(pathResource)

	return &schema.Resource{
		Description:   "Defines field-level actions for matching Formal logs before they are exported or stored.",
		CreateContext: resourceLogSchemaCreate,
		ReadContext:   resourceLogSchemaRead,
		UpdateContext: resourceLogSchemaUpdate,
		DeleteContext: resourceLogSchemaDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ any) error {
			if !d.NewValueKnown("path") {
				return nil
			}
			pathSet, ok := d.Get("path").(*schema.Set)
			if !ok {
				return nil
			}
			_, err := expandLogSchemaPaths(pathSet)
			return err
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of this log schema.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of this log schema.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"encryption_key_id": {
				Description: "The ID of the asymmetric encryption key used by encrypt actions.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"scope_cel": {
				Description: "A CEL expression that determines which logs this schema applies to.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"path": {
				Description: "A log field path and the actions to apply to it.",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem:        pathResource,
				Set: func(value any) int {
					pathData, ok := value.(map[string]any)
					if !ok {
						return pathHash(value)
					}
					normalized := make(map[string]any, len(pathData)+4)
					for key, value := range pathData {
						normalized[key] = value
					}
					for _, action := range []string{"drop", "encrypt", "strip_sql", "encrypt_sql"} {
						if _, ok := normalized[action]; !ok {
							normalized[action] = false
						}
					}
					return pathHash(normalized)
				},
			},
			"created_at": {
				Description: "When the log schema was created.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated_at": {
				Description: "When the log schema was last updated.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func expandLogSchemaPaths(pathSet *schema.Set) (*corev1.LogSchemaPaths, error) {
	paths := make(map[string]*corev1.LogSchemaPath, pathSet.Len())
	for _, rawPath := range pathSet.List() {
		pathData := rawPath.(map[string]any)
		name := pathData["name"].(string)
		if _, exists := paths[name]; exists {
			return nil, fmt.Errorf("path names must be unique: %q is repeated", name)
		}

		path := &corev1.LogSchemaPath{
			Drop:       pathData["drop"].(bool),
			Encrypt:    pathData["encrypt"].(bool),
			StripSql:   pathData["strip_sql"].(bool),
			EncryptSql: pathData["encrypt_sql"].(bool),
		}
		if truncateData := pathData["truncate"].([]any); len(truncateData) > 0 {
			path.Truncate = &corev1.TruncateOperation{
				MaxSizeBytes: uint32(truncateData[0].(map[string]any)["max_size_bytes"].(int)),
			}
		}
		if !path.Drop && !path.Encrypt && path.Truncate == nil && !path.StripSql && !path.EncryptSql {
			return nil, fmt.Errorf("path %q must specify at least one action", name)
		}
		paths[name] = path
	}
	return &corev1.LogSchemaPaths{Paths: paths}, nil
}

func flattenLogSchemaPaths(paths *corev1.LogSchemaPaths) []any {
	if paths == nil {
		return nil
	}

	names := make([]string, 0, len(paths.Paths))
	for name := range paths.Paths {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]any, 0, len(names))
	for _, name := range names {
		path := paths.Paths[name]
		pathData := map[string]any{
			"name":        name,
			"drop":        path.GetDrop(),
			"encrypt":     path.GetEncrypt(),
			"strip_sql":   path.GetStripSql(),
			"encrypt_sql": path.GetEncryptSql(),
		}
		if path.GetTruncate() != nil {
			pathData["truncate"] = []any{map[string]any{
				"max_size_bytes": int(path.GetTruncate().GetMaxSizeBytes()),
			}}
		}
		result = append(result, pathData)
	}
	return result
}

func resourceLogSchemaCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	paths, err := expandLogSchemaPaths(d.Get("path").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &corev1.CreateLogSchemaRequest{
		Name:     d.Get("name").(string),
		ScopeCel: d.Get("scope_cel").(string),
		Paths:    paths,
	}
	if encryptionKeyID, ok := d.GetOk("encryption_key_id"); ok {
		value := encryptionKeyID.(string)
		req.EncryptionKeyId = &value
	}

	res, err := c.Grpc.Sdk.LogsServiceClient.CreateLogSchema(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(res.LogSchema.Id)
	return resourceLogSchemaRead(ctx, d, meta)
}

func resourceLogSchemaRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	res, err := c.Grpc.Sdk.LogsServiceClient.GetLogSchema(ctx, &corev1.GetLogSchemaRequest{Id: d.Id()})
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "The Log Schema was not found, which means it may have been deleted without using this Terraform config.", map[string]any{"err": err})
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	schema := res.LogSchema
	values := map[string]any{
		"id":                schema.Id,
		"name":              schema.Name,
		"scope_cel":         schema.ScopeCel,
		"path":              flattenLogSchemaPaths(schema.Paths),
		"encryption_key_id": "",
		"created_at":        schema.CreatedAt.AsTime().String(),
		"updated_at":        schema.UpdatedAt.AsTime().String(),
	}
	if schema.EncryptionKeyId != nil {
		values["encryption_key_id"] = *schema.EncryptionKeyId
	}
	for key, value := range values {
		if err := d.Set(key, value); err != nil {
			return diag.Errorf("set log schema %s: %v", key, err)
		}
	}
	d.SetId(schema.Id)
	return nil
}

func resourceLogSchemaUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	fieldsThatCanChange := []string{"name", "encryption_key_id", "scope_cel", "path"}
	if d.HasChangesExcept(fieldsThatCanChange...) {
		return diag.Errorf("At the moment you can only update the following fields: %s. If you'd like to update other fields, please message the Formal team and we're happy to help.", strings.Join(fieldsThatCanChange, ", "))
	}

	req := &corev1.UpdateLogSchemaRequest{Id: d.Id()}
	if d.HasChange("name") {
		value := d.Get("name").(string)
		req.Name = &value
	}
	if d.HasChange("encryption_key_id") {
		value := d.Get("encryption_key_id").(string)
		req.EncryptionKeyId = &value
	}
	if d.HasChange("scope_cel") {
		value := d.Get("scope_cel").(string)
		req.ScopeCel = &value
	}
	if d.HasChange("path") {
		paths, err := expandLogSchemaPaths(d.Get("path").(*schema.Set))
		if err != nil {
			return diag.FromErr(err)
		}
		req.Paths = paths
	}

	if _, err := c.Grpc.Sdk.LogsServiceClient.UpdateLogSchema(ctx, req); err != nil {
		return diag.FromErr(err)
	}
	return resourceLogSchemaRead(ctx, d, meta)
}

func resourceLogSchemaDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*clients.Clients)
	if _, err := c.Grpc.Sdk.LogsServiceClient.DeleteLogSchema(ctx, &corev1.DeleteLogSchemaRequest{Id: d.Id()}); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
