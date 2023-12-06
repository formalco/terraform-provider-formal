package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func IntegrationExternalAPIType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"zendesk", "custom"}, false)
}
