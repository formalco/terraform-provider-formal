terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "integration_datahub_sync_direction" {
  type        = string
}

resource "formal_integration_datahub" "name" {
  active                           = true
  api_key                          = "api_key_datahub_placeholder"
  generalized_metadata_service_url = "https://datahub.com"
  sync_direction                   = var.integration_datahub_sync_direction
  synced_entities                  = ["tags"]
}
