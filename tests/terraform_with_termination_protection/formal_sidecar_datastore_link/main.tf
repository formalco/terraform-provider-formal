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
  hostname                   = "terraform-test-local.formal-sidecar-datastore-link.with-termination-protection"
  name                       = "terraform-test-local-formal-sidecar-datastore-link-with-termination-protection"
  technology                 = "postgres"
  db_discovery_job_wait_time = "6h"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_sidecar" "name" {
  name                   = "terraformtestlocalsidecardatastorelinkwithterminationprotection"
  technology             = "postgres"
  hostname               = "test.com"
  termination_protection = false
}

resource "formal_sidecar_resource_link" "name" {
  resource_id           = formal_resource.postgres1.id
  port                   = 5432
  sidecar_id             = formal_sidecar.name.id
  termination_protection = var.termination_protection
}
