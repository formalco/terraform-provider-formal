terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 4.12.3"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_connector" "main" {
  name = var.name
}

resource "formal_connector_configuration" "main" {
  connector_id = formal_connector.main.id
  log_level    = "debug"
}

resource "formal_connector_hostname" "main" {
  connector_id = formal_connector.main.id
  hostname     = var.connector_hostname # e.g. "postgres.<org-name>.connectors.joinformal.com"
  dns_record   = var.dns_record         # CNAME record value to point to
}



