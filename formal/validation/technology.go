package validation

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Technology() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{"snowflake", "postgres", "redshift", "mysql", "mariadb", "s3", "http", "ssh"}, false)
}
