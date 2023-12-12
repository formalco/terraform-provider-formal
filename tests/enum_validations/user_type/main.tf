terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "user_type" {
  type        = string
}

resource "formal_user" "name" {
  type                   = var.user_type
  name                   = "terraform-test-local-formal_user-with-termination-protection"
}
