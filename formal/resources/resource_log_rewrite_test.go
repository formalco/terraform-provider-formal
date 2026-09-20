package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func TestLogRewritePathsRoundTrip(t *testing.T) {
	resource := ResourceLogRewrite()
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

	paths, err := expandLogRewritePaths(data.Get("path").(*schema.Set))
	require.NoError(t, err)
	require.Equal(t, &corev1.LogRewritePath{
		Encrypt:  true,
		Truncate: &corev1.TruncateOperation{MaxSizeBytes: 1024},
	}, paths.Paths["log.request.http.body.received"])
	require.Equal(t, &corev1.LogRewritePath{Drop: true}, paths.Paths["log.response.http.body.received"])

	require.NoError(t, data.Set("path", flattenLogRewritePaths(paths)))
	roundTripped, err := expandLogRewritePaths(data.Get("path").(*schema.Set))
	require.NoError(t, err)
	require.Equal(t, paths, roundTripped)
}

func TestLogRewritePathHashTreatsOmittedActionsAsFalse(t *testing.T) {
	pathSchema := ResourceLogRewrite().Schema["path"]
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

func TestLogRewritePathRequiresAnAction(t *testing.T) {
	resource := ResourceLogRewrite()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"path": []any{map[string]any{"name": "log.request.http.method"}},
	})

	_, err := expandLogRewritePaths(data.Get("path").(*schema.Set))
	require.ErrorContains(t, err, "must specify at least one action")
}

func TestLogRewritePathNamesMustBeUnique(t *testing.T) {
	resource := ResourceLogRewrite()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{
		"path": []any{
			map[string]any{"name": "log.request.http.method", "drop": true},
			map[string]any{"name": "log.request.http.method", "encrypt": true},
		},
	})

	_, err := expandLogRewritePaths(data.Get("path").(*schema.Set))
	require.ErrorContains(t, err, "path names must be unique")
}

func TestLogRewriteTruncateSizeMustBePositive(t *testing.T) {
	truncateSchema := ResourceLogRewrite().Schema["path"].Elem.(*schema.Resource).
		Schema["truncate"].Elem.(*schema.Resource).Schema["max_size_bytes"]

	_, errors := truncateSchema.ValidateFunc(0, "max_size_bytes")
	require.NotEmpty(t, errors)
	_, errors = truncateSchema.ValidateFunc(1, "max_size_bytes")
	require.Empty(t, errors)
}
