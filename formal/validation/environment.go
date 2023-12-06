package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Environment() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"DEV", "TEST", "QA", "UAT", "EI", "PRE", "STG", "NON_PROD", "PROD", "CORP"}, false)
}
