package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestNativeUserSelectionSchemaRequiresRawCel(t *testing.T) {
	resource := ResourceResourceNativeUserSelection()

	require.Len(t, resource.Schema, 3)
	require.Contains(t, resource.Schema, "cel")
	require.Contains(t, resource.Schema, "id")
	require.Contains(t, resource.Schema, "resource_id")
	require.True(t, resource.Schema["cel"].Required)

	diags := resource.Validate(terraform.NewResourceConfigRaw(map[string]any{
		"resource_id": "res-1",
	}))
	require.True(t, diags.HasError())

	diags = resource.Validate(terraform.NewResourceConfigRaw(map[string]any{
		"resource_id": "res-1",
		"cel":         `user.type == "human" ? "nu-human" : "nu-machine"`,
	}))
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
}
