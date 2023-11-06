terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = ">= 3.2.11"
    }
  }
}

provider "formal" {}

variable "termination_protection" {
  type        = bool
  description = "Whether termination protection is enabled for the resource."
}

resource "formal_datastore" "postgres1" {
  hostname                   = "terraform-test-local.formal-native-role-link.with-termination-protection"
  name                       = "terraform-test-local-formal_native_role_link-with-termination-protection"
  technology                 = "postgres"
  db_discovery_job_wait_time = "1m"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_native_role" "name" {
  datastore_id       = formal_datastore.postgres1.id
  native_role_id     = "postgres"
  native_role_secret = "postgres"
}

resource "formal_user" "name" {
  type = "machine"
  name = "terraform-test-local-formal_native_role_link-with-termination-protection"
}

resource "formal_native_role_link" "name" {
  datastore_id           = formal_datastore.postgres1.id
  formal_identity_id     = formal_user.name.id
  formal_identity_type   = "user"
  native_role_id         = formal_native_role.name.native_role_id
  termination_protection = var.termination_protection
}
