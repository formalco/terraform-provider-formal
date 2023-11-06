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

resource "formal_group" "name" {
  description            = "terraform-test-local-formal_group-with-termination-protection"
  name                   = "terraform-test-local-formal_group-with-termination-protection"
  termination_protection = var.termination_protection
}
