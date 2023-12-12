terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "integration_log_type" {
  type        = string
}

resource "formal_integration_log" "name" {
  name                   = "terraform-test-local-formal_integration_log-with-termination-protection"
  type                   = var.integration_log_type
  splunk_api_key         = "aaaaa"
  splunk_url             = "https://splunk.com"
}
