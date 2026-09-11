package resource

import (
	"testing"

	validate "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

func TestResourceTechnologiesMatchAPI(t *testing.T) {
	fd := (&corev1.CreateResourceRequest{}).ProtoReflect().Descriptor().Fields().ByName("technology")
	require.NotNil(t, fd, "technology field not found")

	rules, ok := proto.GetExtension(fd.Options(), validate.E_Field).(*validate.FieldRules)
	require.True(t, ok, "technology field has no buf.validate rules")
	require.Equal(t, rules.GetString().GetIn(), resourceTechnologies)
}

func TestResourceTechnologyValidationAcceptsKubernetes(t *testing.T) {
	validateTechnology := ResourceResource().Schema["technology"].ValidateFunc
	require.NotNil(t, validateTechnology)

	warnings, errors := validateTechnology("kubernetes", "technology")
	require.Empty(t, warnings)
	require.Empty(t, errors)
}

func TestResourceNativeUsersV3EnabledUsesProviderDefault(t *testing.T) {
	field := ResourceResource().Schema["native_users_v3_enabled"]

	require.True(t, field.Optional)
	require.True(t, field.Computed)
	require.Nil(t, field.Default)

	tests := []struct {
		name   string
		config map[string]any
		want   bool
	}{
		{name: "omitted defaults to enabled", config: map[string]any{}, want: true},
		{name: "explicit false", config: map[string]any{"native_users_v3_enabled": false}, want: false},
		{name: "explicit true", config: map[string]any{"native_users_v3_enabled": true}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, ResourceResource().Schema, tt.config)

			got, err := nativeUsersV3EnabledOnCreate(data)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
