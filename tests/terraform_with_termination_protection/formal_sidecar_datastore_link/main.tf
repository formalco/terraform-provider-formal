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
  hostname                   = "terraform-test-local-formal_sidecar_datastore_link-with-termination-protection"
  name                       = "terraform-test-local-formal_sidecar_datastore_link-with-termination-protection"
  technology                 = "postgres"
  db_discovery_job_wait_time = "1m"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_sidecar" "name" {
  name               = "terraform-test-local-formal_sidecar_datastore_link-with-termination-protection"
  deployment_type    = "onprem"
  global_kms_decrypt = false
  technology         = "postgres"
  network_type       = "internal"
}

resource "formal_sidecar_datastore_link" "name" {
  datastore_id           = formal_datastore.postgres1.id
  port                   = 5432
  sidecar_id             = formal_sidecar.name.id
  termination_protection = var.termination_protection
}
