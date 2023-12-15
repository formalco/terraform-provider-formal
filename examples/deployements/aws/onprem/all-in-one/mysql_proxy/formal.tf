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
  technology         = "mysql"
  global_kms_decrypt = false
  network_type       = "internet-facing"
  formal_hostname    = var.mysql_sidecar_hostname
}

resource "formal_datastore" "main" {
  technology = "mysql"
  name       = "${var.name}-datastore"
  hostname   = var.mysql_hostname
  port       = var.main_port
}

resource "formal_sidecar_datastore_link" "main" {
  datastore_id = formal_datastore.main.id
  sidecar_id   = formal_sidecar.main.id
  port         = 3306
}

# Native Role
resource "formal_native_role" "main_mysql" {
  datastore_id       = formal_datastore.main.id
  native_role_id     = var.mysql_username
  native_role_secret = var.mysql_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}
