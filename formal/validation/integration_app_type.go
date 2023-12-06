package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func IntegrationAppType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"metabase", "custom"}, false)
}
