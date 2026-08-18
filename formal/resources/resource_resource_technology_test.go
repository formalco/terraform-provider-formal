package resource

import (
	"testing"

	validate "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
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
