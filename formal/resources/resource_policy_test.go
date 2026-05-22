package resource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourcePolicySchemaIncludesOwnerNotificationRemovalUpgrader(t *testing.T) {
	r := ResourcePolicy()

	require.Equal(t, 2, r.SchemaVersion)
	require.Len(t, r.StateUpgraders, 2)
	require.Equal(t, 1, r.StateUpgraders[1].Version)
}

func TestResourcePolicyStateUpgradeV1RemovesDeprecatedFields(t *testing.T) {
	rawState := map[string]any{
		"id":           "policy_123",
		"name":         "test-policy",
		"owner":        "john@company.com",
		"notification": "all",
	}

	upgraded, err := resourcePolicyStateUpgradeV1(t.Context(), rawState, nil)
	require.NoError(t, err)
	require.NotContains(t, upgraded, "owner")
	require.NotContains(t, upgraded, "notification")
	require.Equal(t, "test-policy", upgraded["name"])
}

func TestResourcePolicyStateUpgradeV1RejectsNilState(t *testing.T) {
	_, err := resourcePolicyStateUpgradeV1(t.Context(), nil, nil)
	require.Error(t, err)
}
