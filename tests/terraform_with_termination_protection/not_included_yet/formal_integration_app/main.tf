terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = ">= 3.2.11"
    }
  }
}

provider "formal" {}

variable "termination_protection" {
  type        = bool
  description = "Whether termination protection is enabled for the resource."
}

resource "formal_integration_app" "name" {
  name                   = "terraform-test-local-formal_integration_app-with-termination-protection"
  type                   = "metabase"
  linked_db_user_id      = "postgres"
  metabase_hostname      = "https://metabase.com"
  metabase_password      = "metabasepassword"
  metabase_username      = "metabaseusername"
  termination_protection = var.termination_protection
}
