terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 4.12.8"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_connector" "main" {
  name = var.name
}

# BigQuery Resource
resource "formal_resource" "bigquery" {
  technology = "bigquery"
  name       = "bigquery"
  hostname   = "bigquery.googleapis.com"
  port       = 443
}

# BigQuery Listener
resource "formal_connector_listener" "bigquery_listener" {
  name = "bigquery-listener"
  port = var.bigquery_port
}

resource "formal_connector_listener_rule" "bigquery_rule" {
  connector_listener_id = formal_connector_listener.bigquery_listener.id
  type                  = "resource"
  rule                  = formal_resource.bigquery.id
}

# Listener Link
resource "formal_connector_listener_link" "bigquery_link" {
  connector_id          = formal_connector.main.id
  connector_listener_id = formal_connector_listener.bigquery_listener.id
}
