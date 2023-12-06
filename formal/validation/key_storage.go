package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func KeyStorage() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"control_plane_and_with_data", "control_plane_only"}, false)
}
