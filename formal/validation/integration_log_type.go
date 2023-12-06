package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func IntegrationLogType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"datadog", "splunk"}, false)
}
