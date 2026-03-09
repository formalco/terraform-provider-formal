package resource

import "testing"

func TestResourceConnectorConfigurationSchemaIncludesRenameUpgrader(t *testing.T) {
	r := ResourceConnectorConfiguration()

	if r.SchemaVersion != 2 {
		t.Fatalf("expected SchemaVersion to be 2, got %d", r.SchemaVersion)
	}

	if len(r.StateUpgraders) != 1 {
		t.Fatalf("expected exactly one state upgrader, got %d", len(r.StateUpgraders))
	}

	if r.StateUpgraders[0].Version != 1 {
		t.Fatalf("expected upgrader version to be 1, got %d", r.StateUpgraders[0].Version)
	}
}

func TestResourceConnectorConfigurationStateUpgradeV1MigratesLegacyKey(t *testing.T) {
	rawState := map[string]interface{}{
		"resources_health_checks_frequency_seconds": 300,
	}

	upgraded, err := resourceConnectorConfigurationStateUpgradeV1(t.Context(), rawState, nil)
	if err != nil {
		t.Fatalf("unexpected upgrade error: %v", err)
	}

	if got, ok := upgraded["resources_health_checks_frequency"]; !ok || got != 300 {
		t.Fatalf("expected migrated value 300 on new key, got %#v", upgraded["resources_health_checks_frequency"])
	}

	if _, exists := upgraded["resources_health_checks_frequency_seconds"]; exists {
		t.Fatalf("expected legacy key to be removed after upgrade")
	}
}

func TestResourceConnectorConfigurationStateUpgradeV1KeepsExistingNewKey(t *testing.T) {
	rawState := map[string]interface{}{
		"resources_health_checks_frequency_seconds": 300,
		"resources_health_checks_frequency":         120,
	}

	upgraded, err := resourceConnectorConfigurationStateUpgradeV1(t.Context(), rawState, nil)
	if err != nil {
		t.Fatalf("unexpected upgrade error: %v", err)
	}

	if got := upgraded["resources_health_checks_frequency"]; got != 120 {
		t.Fatalf("expected existing new key value to be preserved, got %#v", got)
	}

	if _, exists := upgraded["resources_health_checks_frequency_seconds"]; exists {
		t.Fatalf("expected legacy key to be removed after upgrade")
	}
}

func TestResourceConnectorConfigurationStateUpgradeV1RejectsNilState(t *testing.T) {
	_, err := resourceConnectorConfigurationStateUpgradeV1(t.Context(), nil, nil)
	if err == nil {
		t.Fatalf("expected error when upgrading nil state")
	}
}
