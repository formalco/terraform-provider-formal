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

# Cloud SQL Resource
# The connector connects to Cloud SQL via the Cloud SQL Proxy sidecar on localhost:5433
# (port 5433 avoids conflict with the connector's listener on port 5432)
resource "formal_resource" "cloudsql" {
  technology = "postgres"
  name       = coalesce(var.resource_name, "${var.name}-postgres")
  hostname   = "localhost"
  port       = 5433
}

# Disable TLS for the resource since Cloud SQL Proxy handles encryption
resource "formal_resource_tls_configuration" "cloudsql" {
  resource_id = formal_resource.cloudsql.id
  tls_config  = "disable"
}

# Native user for Cloud SQL IAM authentication
# The IAM database user is the service account email with .iam instead of .iam.gserviceaccount.com
# The connector generates the IAM token using iam_gcp auth type
resource "formal_native_user" "cloudsql_iam" {
  resource_id        = formal_resource.cloudsql.id
  native_user_id     = replace(var.gcp_service_account_email, ".gserviceaccount.com", "")
  native_user_secret = "iam_gcp"
  use_as_default     = true
}

# Postgres Listener
resource "formal_connector_listener" "postgres_listener" {
  name         = "${var.name}-postgres-listener"
  port         = var.postgres_port
  connector_id = formal_connector.main.id
}

resource "formal_connector_listener_rule" "postgres_rule" {
  connector_listener_id = formal_connector_listener.postgres_listener.id
  type                  = "technology"
  rule                  = "postgres"
}

# Listener Link
resource "formal_connector_listener_link" "postgres_link" {
  connector_id          = formal_connector.main.id
  connector_listener_id = formal_connector_listener.postgres_listener.id
}
