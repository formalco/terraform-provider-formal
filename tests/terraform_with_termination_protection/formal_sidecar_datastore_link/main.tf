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

resource "formal_resource" "postgres1" {
  hostname                   = "terraform-test-local.formal-sidecar-datastore-link.with-termination-protection.com"
  name                       = "terraform-test-local-formal-sidecar-datastore-link-with-termination-protection"
  technology                 = "http"
  environment                = "DEV"
  port                       = 443
}

resource "formal_sidecar" "name" {
  name                   = "terraform-test-sidecar-resource-link-with-termination-protection"
  technology             = "http"
  hostname               = "test.com"
  termination_protection = false
}

resource "formal_sidecar_resource_link" "name" {
  resource_id           = formal_resource.postgres1.id
  port                   = 443
  sidecar_id             = formal_sidecar.name.id
  termination_protection = var.termination_protection
}
