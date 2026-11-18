package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func connectorListenerState() *terraform.InstanceState {
	return &terraform.InstanceState{
		ID: "connector_listener_test",
		Attributes: map[string]string{
			"id":                     "connector_listener_test",
			"name":                   "postgres",
			"port":                   "5432",
			"connector_id":           "connector_test",
			"termination_protection": "false",
		},
	}
}

func TestConnectorListenerConnectorIDFromAPIWithoutConfigHasNoDiff(t *testing.T) {
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name": "postgres",
		"port": 5432,
	})

	diff, err := ResourceConnectorListener().Diff(t.Context(), connectorListenerState(), config, nil)
	require.NoError(t, err)
	if diff != nil {
		require.NotContains(t, diff.Attributes, "connector_id")
		require.False(t, diff.RequiresNew())
	}
}

func TestConnectorListenerConnectorIDMatchingStateHasNoDiff(t *testing.T) {
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name":         "postgres",
		"port":         5432,
		"connector_id": "connector_test",
	})

	diff, err := ResourceConnectorListener().Diff(t.Context(), connectorListenerState(), config, nil)
	require.NoError(t, err)
	if diff != nil {
		require.NotContains(t, diff.Attributes, "connector_id")
		require.False(t, diff.RequiresNew())
	}
}

func TestConnectorListenerConnectorIDSetOnUnattachedListenerUpdatesInPlace(t *testing.T) {
	state := connectorListenerState()
	delete(state.Attributes, "connector_id")
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name":         "postgres",
		"port":         5432,
		"connector_id": "connector_test",
	})

	diff, err := ResourceConnectorListener().Diff(t.Context(), state, config, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)
	require.Equal(t, "connector_test", diff.Attributes["connector_id"].New)
	require.False(t, diff.RequiresNew())
}

func TestConnectorListenerConnectorIDChangeRequiresReplace(t *testing.T) {
	config := terraform.NewResourceConfigRaw(map[string]any{
		"name":         "postgres",
		"port":         5432,
		"connector_id": "connector_other",
	})

	diff, err := ResourceConnectorListener().Diff(t.Context(), connectorListenerState(), config, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)
	require.Contains(t, diff.Attributes, "connector_id")
	require.True(t, diff.RequiresNew())
}
