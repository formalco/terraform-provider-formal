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

resource "formal_sidecar" "name" {
  deployment_type        = "onprem"
  global_kms_decrypt     = false
  name                   = "terraform-test-local-formal-sidecar-with-termination-protection"
  technology             = "postgres"
  network_type           = "internal"
  termination_protection = var.termination_protection
}
