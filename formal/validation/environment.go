package validation

import (
	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1/types/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Environment() schema.SchemaValidateFunc {
	return validation.StringInSlice(getEnumValues(), false)
}

func getEnumValues() []string {
	validEnumValues := []string{}
	enumValues := adminv1.File_admin_v1_types_v1_datastore_proto.Enums().ByName("Environment").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {
		enumValue := enumValues.Get(i)
		validEnumValues = append(validEnumValues, string(enumValue.Name()))
	}
	return validEnumValues
}
