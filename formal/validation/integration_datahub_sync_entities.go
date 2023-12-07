package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// TODO: Add validation for lists
func IntegrationDatahubSyncEntity() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"tags", "data_labels"}, false)
}
