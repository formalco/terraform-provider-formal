terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.0.18"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~>3.0"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_datastore" "main" {
  technology              = "snowflake"
  name                    = var.name
  hostname                = var.snowflake_hostname
  port                    = var.snowflake_port
  default_access_behavior = "allow"
}

resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "onprem"
  datastore_id       = formal_datastore.main.id
  fail_open          = false
  global_kms_decrypt = false
  network_type       = "internet-facing"
  cloud_provider     = "aws"
  formal_hostname    = var.sidecar_hostname // hostname of the instance running the formal sidecar
}

# Native Role
resource "formal_native_role" "main_postgres" {
  datastore_id       = formal_datastore.main.id
  native_role_id     = var.snowflake_username
  native_role_secret = var.snowflake_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}
