terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.2.3"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "onprem"
  technology         = "ssh"
  network_type       = "internet-facing"
  fail_open          = false
  global_kms_decrypt = false
}

resource "formal_datastore" "instance_1" {
  technology = "ssh"
  name       = "${var.name}-datastore"
  hostname   = var.ssh_hostname
  port       = 22
}

resource "formal_native_role" "main_instance_1" {
  datastore_id       = formal_datastore.instance_1.id
  native_role_id     = var.iam_access_key_id
  native_role_secret = var.iam_secret_access_key
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

resource "formal_sidecar_datastore_link" "link_1" {
  datastore_id = formal_datastore.instance_1.id
  sidecar_id   = formal_sidecar.main.id
  port         = 22
}