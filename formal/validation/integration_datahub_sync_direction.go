package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func IntegrationDatahubSyncDirection() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"bidirectional", "formal_to_datahub", "datahub_to_formal"}, false)
}
