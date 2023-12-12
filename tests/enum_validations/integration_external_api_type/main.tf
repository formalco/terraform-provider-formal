terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "integration_external_api_type" {
  type        = string
}

resource "formal_integration_external_api" "name" {
  auth_type = "basic"
  name      = "terraform-test-local-formal_integration_external_api-with-termination-protection"
  type      = var.integration_external_api_type
  url       = "https://zendesk.com"
}
