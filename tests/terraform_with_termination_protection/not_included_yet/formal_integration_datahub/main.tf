terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "termination_protection" {
  type        = bool
  description = "Whether termination protection is enabled for the resource."
}

# Not priority
# resource "formal_integration_datahub" "name" {
# active                           = true
# api_key                          = "api_key_datahub_placeholder"
# generalized_metadata_service_url = "https://datahub.com"
# sync_direction                   = "bidirectional"
# synced_entities                  = ["tags"]
# termination_protection           = var.termination_protection
# }
