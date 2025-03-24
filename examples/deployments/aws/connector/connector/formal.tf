terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "4.0.15"
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

## Add postgres listener for every postgres resources
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

## Add postgres listener for every kubernetes resource
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

## Add MySQL resource
resource "formal_resource" "main" {
  technology = "mysql"
  name       = "${var.name}-resource"
  hostname   = var.mysql_hostname
  port       = var.main_port
}

# Native Role
resource "formal_native_user" "main_mysql" {
  resource_id        = formal_resource.main.id
  native_user_id     = "test"
  native_user_secret = "test"
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

resource "formal_connector_listener" "mysql_listener" {
  name = "mysql-listener"
  port = var.connector_mysql_port
}

resource "formal_connector_listener_rule" "mysql_rule" {
  connector_listener_id = formal_connector_listener.mysql_listener.id
  type                  = "resource"
  rule                  = fornal_resource.main.id
}

resource "formal_connector_listener_link" "mysql_link" {
  connector_id          = formal_connector.main.id
  connector_listener_id = formal_connector_listener.mysql_listener.id
}
