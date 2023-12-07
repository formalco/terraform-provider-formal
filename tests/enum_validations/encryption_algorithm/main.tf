terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "encryption_algorithm" {
  type        = string
}

resource "formal_datastore" "postgres1" {
  hostname                   = "terraform-test-postgres2"
  name                       = "terraform-test-postgres2"
  technology                 = "postgres"
  db_discovery_job_wait_time = "6h"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_encryption_key" "name" {
  cloud_region = "us-west-1"
  key_id       = "terraform-test-local-formal_field_encryption-with-termination-protection"
  key_name     = "terraform-test-local-formal_field_encryption-with-termination-protection"
}

resource "formal_field_encryption" "name" {
  alg                    = var.encryption_algorithm
  datastore_id           = formal_datastore.postgres1.id
  key_id                 = formal_encryption_key.name.id
  key_storage            = "control_plane_only"
  path                   = "postgres.public.users.id"
}
