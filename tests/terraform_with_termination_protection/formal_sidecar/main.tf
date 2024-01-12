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

resource "formal_sidecar" "name" {
  deployment_type        = "onprem"
  global_kms_decrypt     = false
  name                   = "terraform-test-local-formal-sidecar-with-termination-protection"
  technology             = "postgres"
  termination_protection = var.termination_protection
}
