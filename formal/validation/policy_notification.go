package validation

import (
	"strings"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1/types/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func PolicyNotification() schema.SchemaValidateFunc {
	return validation.StringInSlice(getPolicyNotificationEnumValues(), false)
}

func getPolicyNotificationEnumValues() []string {
	validEnumValues := []string{}

	enumValues := adminv1.File_admin_v1_types_v1_policy_proto.Enums().ByName("PolicyNotification").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {
		enumValue := enumValues.Get(i)
		policyNotificationName := strings.ToLower(string(enumValue.Name()))
		validEnumValues = append(validEnumValues, policyNotificationName)
	}

	return validEnumValues
}
