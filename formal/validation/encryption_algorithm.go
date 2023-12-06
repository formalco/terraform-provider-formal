package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func EncryptionAlgorithm() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"aes_random", "aes_deterministic"}, false)
}
