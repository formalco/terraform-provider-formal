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

resource "formal_datastore" "postgres1" {
  hostname                   = "terraform.test-local.formal-datastore-with-termination-protection"
  name                       = "terraform-test-local-formal_datastore-with-termination-protection"
  technology                 = "postgres"
  db_discovery_job_wait_time = "1m"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
  termination_protection = var.termination_protection
}
