terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "deployment_type" {
  type        = string
}

resource "formal_sidecar" "name" {
  deployment_type        = var.deployment_type
  global_kms_decrypt     = false
  name                   = "terraform-test-local-formal-sidecar-with-termination-protection"
  technology             = "postgres"
  network_type           = "internal"
}
