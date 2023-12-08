package validation

import (
	"strings"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1/types/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func IntegrationLogType() schema.SchemaValidateFunc {
	return validation.StringInSlice(getIntegrationLogTypeEnumValues(), false)
}

func getIntegrationLogTypeEnumValues() []string {
	validEnumValues := []string{}

	enumValues := adminv1.File_admin_v1_types_v1_integration_log_proto.Enums().ByName("IntegrationLogType").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {
		enumValue := enumValues.Get(i)
		enumName := string(enumValue.Name())
		enumName = strings.ReplaceAll(enumName, "INTEGRATION_LOG_TYPE_", "")
		enumName = strings.ToLower(enumName)

		if enumName == "unspecified" {
			continue
		}

		validEnumValues = append(validEnumValues, enumName)
	}

	return validEnumValues
}
