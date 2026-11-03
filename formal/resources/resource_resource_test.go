package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestResourceEnvironmentRemovedFromConfigHasNoDiff(t *testing.T) {
	state := &terraform.InstanceState{
		ID: "resource_test",
		Attributes: map[string]string{
			"id":                     "resource_test",
			"name":                   "postgres-prod",
			"hostname":               "db.internal",
			"technology":             "postgres",
			"port":                   "5432",
			"environment":            "PROD",
			"termination_protection": "false",
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name":       "postgres-prod",
		"hostname":   "db.internal",
		"technology": "postgres",
		"port":       5432,
	})

	diff, err := ResourceResource().Diff(t.Context(), state, config, nil)
	require.NoError(t, err)
	if diff != nil {
		require.NotContains(t, diff.Attributes, "environment")
	}
}
