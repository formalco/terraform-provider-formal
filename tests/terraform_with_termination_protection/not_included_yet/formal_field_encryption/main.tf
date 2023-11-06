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
  hostname                   = "terraform-test-postgres2"
  name                       = "terraform-test-postgres2"
  technology                 = "postgres"
  db_discovery_job_wait_time = "1m"
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
  alg                    = "aes_deterministic"
  datastore_id           = formal_datastore.postgres1.id
  key_id                 = formal_encryption_key.name.id
  key_storage            = "control_plane_only"
  path                   = "postgres.public.users.id"
  termination_protection = var.termination_protection
}
