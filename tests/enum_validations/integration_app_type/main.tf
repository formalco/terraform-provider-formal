terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "integration_app_type" {
  type        = string
}

resource "formal_integration_app" "name" {
  name                   = "terraform-test-local-formal_integration_app-with-termination-protection"
  type                   = var.integration_app_type
  linked_db_user_id      = "postgres"
  metabase_hostname      = "https://metabase.com"
  metabase_password      = "metabasepassword"
  metabase_username      = "metabaseusername"
}
