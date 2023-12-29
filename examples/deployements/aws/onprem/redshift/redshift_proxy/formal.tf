terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.2.3"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "onprem"
  technology         = "redshift"
  global_kms_decrypt = false
  formal_hostname    = var.redshift_sidecar_hostname
}

resource "formal_datastore" "ds" {
  technology = "redshift"
  name       = var.name
  hostname   = var.redshift_hostname
  port       = var.main_port
}

resource "formal_sidecar_datastore_link" "main" {
  datastore_id = formal_datastore.ds.id
  sidecar_id   = formal_sidecar.main.id
  port         = 5439
}

# Native Role
resource "formal_native_role" "main_redshift" {
  datastore_id       = formal_datastore.ds.id
  native_role_id     = var.redshift_username
  native_role_secret = var.redshift_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}
