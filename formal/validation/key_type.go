package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func KeyType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"aws_kms"}, false)
}
