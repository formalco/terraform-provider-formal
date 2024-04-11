terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>4.0.0"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name       = var.name
  technology = "http"
  hostname   = var.sidecar_hostname
}

resource "formal_resource" "main" {
  technology = "http"
  name       = var.name
  hostname   = var.datastore_hostname
  port       = var.main_port
}

resource "formal_sidecar_resource_link" "main" {
  resource_id = formal_resource.main.id
  sidecar_id  = formal_sidecar.main.id
  port        = 443
}