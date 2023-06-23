terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.0.18"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~>3.0"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_datastore" "instance_1" {
  technology = "ssh"
  name       = var.name
  hostname   = var.hostname
  port       = var.port
}

resource "formal_native_role" "main_snowflake" {
  datastore_id       = formal_datastore.instance_1.id
  native_role_id     = var.username
  native_role_secret = var.password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

resource "formal_datastore" "instance_2" {
  technology = "ssh"
  name       = var.name
  hostname   = var.hostname
  port       = var.port
}

resource "formal_native_role" "main_snowflake" {
  datastore_id       = formal_datastore.instance_2.id
  native_role_id     = var.username
  native_role_secret = var.password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}


resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "onprem"
  technology         = "ssh"
  fail_open          = false
  global_kms_decrypt = false
  network_type       = "internet-facing"
}

resource "formal_sidecar_datastore_link" "link_1" {
  datastore_id = formal_datastore.instance_1.id
  sidecar_id   = formal_sidecar.main.id
  port         = 2022
}

resource "formal_sidecar_datastore_link" "link_2" {
  datastore_id = formal_datastore.instance_2.id
  sidecar_id   = formal_sidecar.main.id
  port         = 2023
}