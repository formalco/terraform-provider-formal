terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>4.0.10"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name       = var.name
  technology = "kubernetes"
  hostname   = var.kubernetes_sidecar_hostname
}

resource "formal_resource" "main" {
  technology = "kubernetes"
  name       = "${var.name}-datastore"
  hostname   = var.kubernetes_hostname
  port       = var.main_port
}

resource "formal_sidecar_resource_link" "main" {
  resource_id = formal_resource.main.id
  sidecar_id  = formal_sidecar.main.id
  port        = 443
}

# Native Role
resource "formal_native_user" "main_kubernetes" {
  resource_id        = formal_resource.main.id
  native_user_id     = var.kubernetes_username
  native_user_secret = var.kubernetes_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}
