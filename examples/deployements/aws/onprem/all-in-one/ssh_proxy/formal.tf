terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>4.0.0"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name       = var.name
  technology = "ssh"
  hostname   = var.ssh_sidecar_hostname
}

resource "formal_resource" "instance_1" {
  technology = "ssh"
  name       = "${var.name}-resource"
  hostname   = var.ssh_hostname
  port       = 22
}

resource "formal_native_user" "main_instance_1" {
  resource_id        = formal_resource.instance_1.id
  native_user_id     = var.iam_access_key_id
  native_user_secret = var.iam_secret_access_key
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

resource "formal_sidecar_resource_link" "link_1" {
  resource_id = formal_resource.instance_1.id
  sidecar_id  = formal_sidecar.main.id
  port        = 22
}