package validation

import (
	"strings"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1/types/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func KeyStorage() schema.SchemaValidateFunc {
	return validation.StringInSlice(getKeyStorageEnumValues(), false)
}

func getKeyStorageEnumValues() []string {
	validEnumValues := []string{}

	enumValues := adminv1.File_admin_v1_types_v1_default_field_encryption_policy_proto.Enums().ByName("KeyStorage").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {
		enumValue := enumValues.Get(i)
		validEnumValues = append(validEnumValues, strings.ToLower(string(enumValue.Name())))
	}

	return validEnumValues
}
