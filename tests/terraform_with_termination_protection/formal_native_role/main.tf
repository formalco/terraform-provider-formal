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
  hostname                   = "terraform-test-local.formal-native-role.with-termination-protection"
  name                       = "terraform-test-local-formal_native_role-with-termination-protection"
  technology                 = "postgres"
  db_discovery_job_wait_time = "6h"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_native_role" "name" {
  resource_id           = formal_resource.postgres1.id
  native_role_id         = "postgres"
  native_role_secret     = "postgres"
  use_as_default         = true
  termination_protection = var.termination_protection
}
