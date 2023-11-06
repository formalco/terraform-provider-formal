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

resource "formal_satellite" "name" {
  name                   = "terraform-test-local-formal_satellite-with-termination-protection"
  termination_protection = var.termination_protection
}
