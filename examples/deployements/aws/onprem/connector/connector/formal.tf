terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "4.0.13"
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

resource "formal_connector_hostname" "main" {
  connector_id = formal_connector.main.id
  hostname     = var.connector_hostname
  managed_tls  = true
}

resource "formal_connector_listener" "postgres_listener" {
  name = var.connector_postgres_listener_name
  port = var.connector_postgres_listener_port
}

resource "formal_connector_listener_rule" "postgres_rule" {
  connector_listener_id = formal_connector_listener.postgres_listener.id
  type                  = "technology"
  rule                  = "postgres"
}

resource "formal_connector_listener_link" "postgres_link" {
  connector_id          = formal_connector.main.id
  connector_listener_id = formal_connector_listener.postgres_listener.id
}

resource "formal_connector_listener" "kubernetes_listener" {
  name = var.connector_postgres_listener_name
  port = var.connector_kubernetes_port
}

resource "formal_connector_listener_rule" "kubernetes_rule" {
  connector_listener_id = formal_connector_listener.kubernetes_listener.id
  type                  = "technology"
  rule                  = "kubernetes"
}

resource "formal_connector_listener_link" "kubernetes_link" {
  connector_id          = formal_connector.main.id
  connector_listener_id = formal_connector_listener.kubernetes_listener.id
}