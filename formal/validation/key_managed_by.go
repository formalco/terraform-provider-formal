package validation

import (
	"strings"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1/types/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func KeyManagedBy() schema.SchemaValidateFunc {
	return validation.StringInSlice(getKeyManagedByEnumValues(), false)
}

func getKeyManagedByEnumValues() []string {
	validEnumValues := []string{}

	enumValues := adminv1.File_admin_v1_types_v1_key_proto.Enums().ByName("KeyManagedBy").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {
		enumValue := enumValues.Get(i)
		validEnumValues = append(validEnumValues, strings.ToLower(string(enumValue.Name())))
	}

	return validEnumValues
}
