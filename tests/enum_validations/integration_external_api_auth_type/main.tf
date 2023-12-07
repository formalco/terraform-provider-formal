terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "auth_type" {
  type        = string
}

resource "formal_integration_external_api" "name" {
  auth_type = var.auth_type
  name      = "terraform-test-local-formal_integration_external_api-with-termination-protection"
  type      = "custom"
  url       = "https://zendesk.com"
}
