terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.27.0"
    }
    formal = {
      source  = "formalco/formal"
      version = "~> 4.23"
    }
  }
}

provider "aws" {
  region = var.region
}

provider "formal" {
  api_key = var.formal_api_key
}

locals {
  ecr_registry    = "654654333078.dkr.ecr.${var.ecr_region}.amazonaws.com"
  connector_image = "${local.ecr_registry}/formalco-prod-connector:${var.connector_image_tag}"
  satellite_image = "${local.ecr_registry}/formalco-prod-policy-data-loader-satellite:${var.satellite_image_tag}"

  connector_ecr_repo = "arn:aws:ecr:${var.ecr_region}:654654333078:repository/formalco-prod-connector"
  satellite_ecr_repo = "arn:aws:ecr:${var.ecr_region}:654654333078:repository/formalco-prod-policy-data-loader-satellite"

  ecs_cluster_name = element(split("/", var.ecs_cluster_arn), length(split("/", var.ecs_cluster_arn)) - 1)
}

resource "formal_connector" "main" {
  name = var.name
}

resource "formal_connector_configuration" "main" {
  connector_id           = formal_connector.main.id
  log_level              = var.connector_log_level
  otel_endpoint_hostname = "localhost"
  otel_endpoint_port     = 4317
}

resource "formal_satellite" "policy_data_loader" {
  name           = "${var.name}-policy-data-loader"
  satellite_type = "policy_data_loader"
}

resource "formal_connector_satellite_link" "policy_data_loader" {
  connector_id   = formal_connector.main.id
  satellite_id   = formal_satellite.policy_data_loader.id
  satellite_type = "policy_data_loader"
}

# Minimal example loader; replace worker_code with your own logic.
resource "formal_policy_data_loader" "example" {
  name            = "${var.name}-example"
  description     = "Example policy data loader"
  key             = "example"
  status          = "active"
  worker_runtime  = "python3.11"
  worker_schedule = var.example_policy_data_loader_schedule

  worker_code = <<-PYTHON
    import json
    print(json.dumps({"loaded": True}))
  PYTHON
}
