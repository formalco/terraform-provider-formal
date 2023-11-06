terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = ">= 3.2.11"
    }
  }
}

provider "formal" {}

variable "termination_protection" {
  type        = bool
  description = "Whether termination protection is enabled for the resource."
}

resource "formal_integration_log" "name" {
  name                   = "terraform-test-local-formal_integration_log-with-termination-protection"
  type                   = "splunk"
  splunk_api_key         = "aaaaa"
  splunk_url             = "https://splunk.com"
  termination_protection = var.termination_protection
}
