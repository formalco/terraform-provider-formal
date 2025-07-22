terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "4.10.1"
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
  connector_id      = formal_connector.main.id
  log_level         = "debug"
  health_check_port = 8080
}

resource "formal_connector_hostname" "main" {
  connector_id = formal_connector.main.id
  hostname     = var.connector_hostname
  dns_record   = var.connector_dns_record
}



