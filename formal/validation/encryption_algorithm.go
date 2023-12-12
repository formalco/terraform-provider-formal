package validation

import (
	"strings"

	adminv1 "buf.build/gen/go/formal/admin/protocolbuffers/go/admin/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func EncryptionAlgorithm() schema.SchemaValidateFunc {
	return validation.StringInSlice(getEncryptionAlgorithmEnumValues(), false)
}

func getEncryptionAlgorithmEnumValues() []string {
	validEnumValues := []string{}

	enumValues := adminv1.File_admin_v1_encryption_proto.Enums().ByName("EncryptionAlgorithm").Values()
	len := enumValues.Len()
	for i := 0; i < len; i++ {
		enumValue := enumValues.Get(i)
		enumName := string(enumValue.Name())
		enumName = strings.ReplaceAll(enumName, "ENCRYPTION_ALGORITHM_", "")
		enumName = strings.ToLower(enumName)

		if enumName == "unspecified" {
			continue
		}

		validEnumValues = append(validEnumValues, enumName)
	}

	return validEnumValues
}
