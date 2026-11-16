package resource

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func protoHasFieldPath(md protoreflect.MessageDescriptor, path string) bool {
	parts := strings.Split(path, ".")
	for i, part := range parts {
		field := md.Fields().ByName(protoreflect.Name(part))
		if field == nil {
			return false
		}
		if i < len(parts)-1 {
			if field.Message() == nil {
				return false
			}
			md = field.Message()
		}
	}
	return true
}

// Every field a Terraform Update func accepts must be sent to the API, otherwise
// Terraform accepts the change, Read restores the server value and the diff never converges.
func TestFieldsThatCanChangeExistInUpdateRequest(t *testing.T) {
	tests := []struct {
		name    string
		fields  []string
		request proto.Message
		// Terraform field name -> proto field path, when they differ.
		protoPaths map[string]string
		// Fields sent through a different RPC than the main Update request.
		sentElsewhere []string
	}{
		{name: "formal_connector", fields: connectorFieldsThatCanChange, request: &corev1.UpdateConnectorRequest{}},
		{name: "formal_connector_configuration", fields: connectorConfigurationFieldsThatCanChange, request: &corev1.UpdateConnectorConfigurationRequest{}},
		{name: "formal_connector_hostname", fields: connectorHostnameFieldsThatCanChange, request: &corev1.UpdateConnectorHostnameRequest{}},
		{name: "formal_connector_listener", fields: connectorListenerFieldsThatCanChange, request: &corev1.UpdateConnectorListenerRequest{}},
		{name: "formal_connector_listener_link", fields: connectorListenerLinkFieldsThatCanChange, request: &corev1.UpdateConnectorListenerLinkRequest{}},
		{name: "formal_connector_listener_rule", fields: connectorListenerRuleFieldsThatCanChange, request: &corev1.UpdateConnectorListenerRuleRequest{}},
		{name: "formal_data_discovery", fields: dataDiscoveryFieldsThatCanChange, request: &corev1.UpdateDataDiscoveryConfigurationRequest{}},
		{name: "formal_data_label", fields: dataLabelFieldsThatCanChange, request: &corev1.UpdateDataLabelRequest{}},
		{
			name:    "formal_resource_dial_configuration",
			fields:  dialConfigurationFieldsThatCanChange,
			request: &corev1.UpdateResourceDialConfigurationRequest{},
			protoPaths: map[string]string{
				"dial_method": "resource_dial_configuration.dial_method",
				"dial_target": "resource_dial_configuration.dial_target",
			},
		},
		{
			name:       "formal_encryption_key",
			fields:     encryptionKeyFieldsThatCanChange,
			request:    &corev1.UpdateEncryptionKeyRequest{},
			protoPaths: map[string]string{"key_provider": "provider"},
		},
		{name: "formal_resource_health_check", fields: healthCheckFieldsThatCanChange, request: &corev1.UpdateResourceHealthCheckRequest{}},
		{name: "formal_integration_cloud", fields: integrationCloudFieldsThatCanBeUpdated, request: &corev1.UpdateCloudIntegrationRequest{}},
		{
			name:       "formal_inventory_object_data_label_link",
			fields:     inventoryObjectDataLabelLinkFieldsThatCanChange,
			request:    &corev1.UpdateColumnRequest{},
			protoPaths: map[string]string{"locked": "lock_status_validated"},
		},
		{name: "formal_log_configuration", fields: logConfigurationFieldsThatCanChange, request: &corev1.UpdateLogConfigurationRequest{}},
		{
			name:       "formal_log_schema",
			fields:     logSchemaFieldsThatCanChange,
			request:    &corev1.UpdateLogSchemaRequest{},
			protoPaths: map[string]string{"path": "paths"},
		},
		{name: "formal_policy_data_loader", fields: policyDataLoaderFieldsThatCanChange, request: &corev1.UpdatePolicyDataLoaderRequest{}},
		{
			name:          "formal_resource",
			fields:        resourceFieldsThatCanChange,
			request:       &corev1.UpdateResourceRequest{},
			sentElsewhere: []string{"native_users_v3_enabled"},
		},
		{name: "formal_resource_hostname", fields: resourceHostnameFieldsThatCanChange, request: &corev1.UpdateResourceHostnameRequest{}},
		{name: "formal_satellite_hostname", fields: satelliteHostnameFieldsThatCanChange, request: &corev1.UpdateSatelliteHostnameRequest{}},
		{name: "formal_space", fields: spaceFieldsThatCanChange, request: &corev1.UpdateSpaceRequest{}},
		{name: "formal_resource_ssh_host_key", fields: sshHostKeyFieldsThatCanChange, request: &corev1.UpdateResourceSshHostKeyRequest{}},
		{name: "formal_resource_tls_configuration", fields: tlsConfigurationFieldsThatCanChange, request: &corev1.UpdateResourceTlsConfigurationRequest{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := tt.request.ProtoReflect().Descriptor()
			for _, field := range tt.fields {
				if slices.Contains(tt.sentElsewhere, field) {
					continue
				}
				path := field
				if p, ok := tt.protoPaths[field]; ok {
					path = p
				}
				require.Truef(t, protoHasFieldPath(md, path), "%q is in fieldsThatCanChange but %s has no field %q", field, md.FullName(), path)
			}
		})
	}
}
