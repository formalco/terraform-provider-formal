terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "network_type" {
  type        = string
}

resource "formal_sidecar" "name" {
  deployment_type        = "onprem"
  global_kms_decrypt     = false
  name                   = "terraform-test-local-formal-sidecar-with-termination-protection"
  technology             = "postgres"
  network_type           = var.network_type
}
