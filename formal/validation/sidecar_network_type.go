package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func SidecarNetworkType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"internet-facing", "internal", "internet-and-internal"}, false)
}
