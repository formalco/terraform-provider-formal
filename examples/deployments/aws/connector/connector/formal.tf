terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 4.12.8"
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
  connector_id           = formal_connector.main.id
  log_level              = "debug"
  otel_endpoint_hostname = "localhost"
  otel_endpoint_port     = 4317
}



