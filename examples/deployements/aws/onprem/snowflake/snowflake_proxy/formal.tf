terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>4.0.0"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name       = var.name
  technology = "snowflake"
  hostname   = var.snowflake_sidecar_hostname
}

resource "formal_resource" "main" {
  technology = "snowflake"
  name       = var.name
  hostname   = var.snowflake_hostname
  port       = var.main_port
}

resource "formal_sidecar_resource_link" "main" {
  resource_id = formal_resource.main.id
  sidecar_id  = formal_sidecar.main.id
  port        = 443
}

# Native Role
resource "formal_native_user" "main_snowflake" {
  resource_id        = formal_resource.main.id
  native_user_id     = var.snowflake_username
  native_user_secret = "ENV:SNOWFLAKE_PASSWORD"
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}
