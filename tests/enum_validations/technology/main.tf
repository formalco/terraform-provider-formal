terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "technology" {
  type        = string
}

resource "formal_datastore" "postgres1" {
  hostname                   = "test-local.formal-datastore-with-termination-protection"
  name                       = "terraform-test-local-formal_datastore-with-termination-protection"
  technology                 = var.technology
  db_discovery_job_wait_time = "6h"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}
