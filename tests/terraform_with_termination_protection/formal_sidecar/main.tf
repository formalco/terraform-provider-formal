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
  name                   = "test-formal-sidecar-with-termination-protection"
  technology             = "http"
  hostname               = "echo.com"
  termination_protection = var.termination_protection
}
