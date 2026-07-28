package resource

import (
	"testing"

	validate "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	corev1 "github.com/formalco/go-sdk/v3/core/v1"
)

// The provider's rule pattern must match the API's buf.validate constraint,
// which is pinned to the backend set of technologies.
func TestConnectorListenerRulePatternMatchesAPI(t *testing.T) {
	fd := (&corev1.CreateConnectorListenerRuleRequest{}).ProtoReflect().Descriptor().Fields().ByName("rule")
	require.NotNil(t, fd, "rule field not found")

	rules, ok := proto.GetExtension(fd.Options(), validate.E_Field).(*validate.FieldRules)
	require.True(t, ok, "rule field has no buf.validate rules")

	want := rules.GetString().GetPattern()
	require.NotEmpty(t, want, "rule field has no buf.validate pattern")
	require.Equal(t, want, connectorListenerRuleValuePattern.String())
}
