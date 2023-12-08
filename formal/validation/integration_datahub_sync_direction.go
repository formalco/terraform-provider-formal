package validation

import (
	"strings"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1/types/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func IntegrationDatahubSyncDirection() schema.SchemaValidateFunc {
	return validation.StringInSlice(getIntegrationDatahubSyncDirectionEnumValues(), false)
}

func getIntegrationDatahubSyncDirectionEnumValues() []string {
	validEnumValues := []string{}

	enumValues := adminv1.File_admin_v1_types_v1_datahub_proto.Enums().ByName("IntegrationDatahubSyncDirection").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {
		enumValue := enumValues.Get(i)
		enumName := string(enumValue.Name())
		enumName = strings.ReplaceAll(enumName, "INTEGRATION_DATAHUB_SYNC_DIRECTION_", "")
		enumName = strings.ToLower(enumName)

		if enumName == "unspecified" {
			continue
		}

		validEnumValues = append(validEnumValues, enumName)
	}

	return validEnumValues
}
