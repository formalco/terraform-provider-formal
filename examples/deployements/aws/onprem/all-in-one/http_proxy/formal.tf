terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.4.0"
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
  technology         = "http"
  global_kms_decrypt = false
  formal_hostname    = var.sidecar_hostname
}

# resource "formal_datastore" "main" {
#   technology = "http"
#   name       = var.name
#   hostname   = "zzzzz.fly.dev"
#   port       = var.main_port
# }

resource "formal_datastore" "main" {
  technology = "http"
  name       = "${var.name}-datastore"
  hostname   = "api.stripe.com"
  port       = var.main_port
}


# resource "formal_sidecar_datastore_link" "main" {
#   datastore_id = formal_datastore.main.id
#   sidecar_id   = formal_sidecar.main.id
#   port         = 443
# }

# resource "formal_sidecar_datastore_link" "stripe" {
#   datastore_id = formal_datastore.stripe.id
#   sidecar_id   = formal_sidecar.main.id
#   port         = 444
# }