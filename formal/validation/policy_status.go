package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func PolicyStatus() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"draft", "dry-run", "active"}, false)
}
