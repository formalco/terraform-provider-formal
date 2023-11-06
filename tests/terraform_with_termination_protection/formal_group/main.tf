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

resource "formal_group" "name" {
  description            = "terraform-test-local-formal_group-with-termination-protection"
  name                   = "terraform-test-local-formal_group-with-termination-protection"
  termination_protection = var.termination_protection
}
