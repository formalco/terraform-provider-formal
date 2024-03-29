terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "termination_protection" {
  type        = bool
  description = "Whether termination protection is enabled for the resource."
}

resource "formal_integration_external_api" "name" {
  auth_type = "basic"
  name      = "terraform-test-local-formal_integration_external_api-with-termination-protection"
  type      = "custom"
  url       = "https://zendesk.com"
  # termination_protection = var.termination_protection
}
