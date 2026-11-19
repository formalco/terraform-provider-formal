package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func TestLogSchemaPathsRoundTrip(t *testing.T) {
	resource := ResourceLogSchema()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"path": []any{
			map[string]any{
				"name":      "log.request.http.body.received",
				"truncate":  []any{map[string]any{"max_size_bytes": 1024}},
				"encrypt":   true,
				"strip_sql": false,
			},
			map[string]any{
				"name": "log.response.http.body.received",
				"drop": true,
			},
		},
	})

	paths, err := expandLogSchemaPaths(data.Get("path").(*schema.Set))
	require.NoError(t, err)
	require.Equal(t, &corev1.LogSchemaPath{
		Encrypt:  true,
		Truncate: &corev1.TruncateOperation{MaxSizeBytes: 1024},
	}, paths.Paths["log.request.http.body.received"])
	require.Equal(t, &corev1.LogSchemaPath{Drop: true}, paths.Paths["log.response.http.body.received"])

	require.NoError(t, data.Set("path", flattenLogSchemaPaths(paths)))
	roundTripped, err := expandLogSchemaPaths(data.Get("path").(*schema.Set))
	require.NoError(t, err)
	require.Equal(t, paths, roundTripped)
}

func TestLogSchemaPathHashTreatsOmittedActionsAsFalse(t *testing.T) {
	pathSchema := ResourceLogSchema().Schema["path"]
	omitted := map[string]any{
		"name": "log.request.http.method",
	}
	explicit := map[string]any{
		"name":        "log.request.http.method",
		"drop":        false,
		"encrypt":     false,
		"strip_sql":   false,
		"encrypt_sql": false,
	}

	require.Equal(t, pathSchema.Set(omitted), pathSchema.Set(explicit))
}

func TestLogSchemaPathRequiresAnAction(t *testing.T) {
	resource := ResourceLogSchema()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"path": []any{map[string]any{"name": "log.request.http.method"}},
	})

	_, err := expandLogSchemaPaths(data.Get("path").(*schema.Set))
	require.ErrorContains(t, err, "must specify at least one action")
}

func TestLogSchemaPathNamesMustBeUnique(t *testing.T) {
	resource := ResourceLogSchema()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"path": []any{
			map[string]any{"name": "log.request.http.method", "drop": true},
			map[string]any{"name": "log.request.http.method", "encrypt": true},
		},
	})

	_, err := expandLogSchemaPaths(data.Get("path").(*schema.Set))
	require.ErrorContains(t, err, "path names must be unique")
}

func TestLogSchemaTruncateSizeMustBePositive(t *testing.T) {
	truncateSchema := ResourceLogSchema().Schema["path"].Elem.(*schema.Resource).
		Schema["truncate"].Elem.(*schema.Resource).Schema["max_size_bytes"]

	_, errors := truncateSchema.ValidateFunc(0, "max_size_bytes")
	require.NotEmpty(t, errors)
	_, errors = truncateSchema.ValidateFunc(1, "max_size_bytes")
	require.Empty(t, errors)
}

func TestLogSchemaPlanAcceptsTruncate(t *testing.T) {
	r := ResourceLogSchema()
	raw := map[string]any{
		"name":      "truncate-bodies",
		"scope_cel": "true",
		"path": []any{map[string]any{
			"name":     "log.request.http.body.received",
			"truncate": []any{map[string]any{"max_size_bytes": 1024}},
		}},
	}
	diff, err := r.Diff(t.Context(), nil, terraform.NewResourceConfigRaw(raw), nil)
	require.NoError(t, err)

	d, err := schema.InternalMap(r.SchemaMap()).Data(nil, diff)
	require.NoError(t, err)
	paths, err := expandLogSchemaPaths(d.Get("path").(*schema.Set))
	require.NoError(t, err)
	require.Equal(t, &corev1.LogSchemaPath{
		Truncate: &corev1.TruncateOperation{MaxSizeBytes: 1024},
	}, paths.Paths["log.request.http.body.received"])
}

func TestLogSchemaPlanRequiresAnAction(t *testing.T) {
	raw := map[string]any{
		"name":      "no-action",
		"scope_cel": "true",
		"path":      []any{map[string]any{"name": "log.request.http.body.received"}},
	}
	_, err := ResourceLogSchema().Diff(t.Context(), nil, terraform.NewResourceConfigRaw(raw), nil)
	require.ErrorContains(t, err, "must specify at least one action")
}
