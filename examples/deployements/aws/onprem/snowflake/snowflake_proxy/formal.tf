terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.2.3"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "onprem"
  technology         = "snowflake"
  global_kms_decrypt = false
  formal_hostname    = var.snowflake_sidecar_hostname
}

resource "formal_datastore" "main" {
  technology = "snowflake"
  name       = var.name
  hostname   = var.snowflake_hostname
  port       = var.main_port
}

resource "formal_sidecar_datastore_link" "main" {
  datastore_id = formal_datastore.main.id
  sidecar_id   = formal_sidecar.main.id
  port         = 443
}

# Native Role
resource "formal_native_role" "main_snowflake" {
  datastore_id       = formal_datastore.main.id
  native_role_id     = var.snowflake_username
  native_role_secret = var.snowflake_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}
