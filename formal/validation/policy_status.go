package validation

import (
	"strings"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1/types/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func PolicyStatus() schema.SchemaValidateFunc {
	return validation.StringInSlice(getPolicyStatusEnumValues(), false)
}

func getPolicyStatusEnumValues() []string {
	validEnumValues := []string{}

	enumValues := adminv1.File_admin_v1_types_v1_policy_proto.Enums().ByName("PolicyStatus").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {

		enumValue := enumValues.Get(i)
		enumName := string(enumValue.Name())
		enumName = strings.ReplaceAll(enumName, "POLICY_STATUS_", "")
		enumName = strings.ToLower(enumName)

		if enumName == "unspecified" {
			continue
		}
		if enumName == "dry_run" {
			enumName = "dry-run"
		}

		validEnumValues = append(validEnumValues, enumName)
	}

	return validEnumValues
}
