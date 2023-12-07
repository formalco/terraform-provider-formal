terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "environment" {
  type        = string
}

resource "formal_datastore" "postgres1" {
  hostname                   = "test-local.formal-datastore-with-termination-protection"
  name                       = "terraform-test-local-formal_datastore-with-termination-protection"
  technology                 = "ssh"
  db_discovery_job_wait_time = "6h"
  environment                = var.environment
  port                       = 5432
  timeouts {
    create = "1m"
  }
}
