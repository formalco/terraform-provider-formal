package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func KeyManagedBy() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"saas_managed", "managed_cloud", "customer_managed"}, false)
}
