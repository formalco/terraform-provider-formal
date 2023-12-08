package validation

import (
	"strings"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func IntegrationExternalAPIAuthType() schema.SchemaValidateFunc {
	return validation.StringInSlice(getIntegrationExternalAPIAuthTypeEnumValues(), false)
}

func getIntegrationExternalAPIAuthTypeEnumValues() []string {
	validEnumValues := []string{}

	enumValues := adminv1.File_admin_v1_integration_external_api_proto.Enums().ByName("IntegrationExternalAPIAuthType").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {
		enumValue := enumValues.Get(i)
		validEnumValues = append(validEnumValues, strings.ToLower(string(enumValue.Name())))
	}

	return validEnumValues
}
