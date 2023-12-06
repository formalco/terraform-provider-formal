package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func RoleAppType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"metabase", "tableau", "popsql"}, false)
}
