terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "app_type" {
  type        = string
}

resource "formal_role" "role" {
  name       = "terraform-test-local-formal_role-with-termination-protection"
  type       = "machine"
  app_type   = var.app_type
}
