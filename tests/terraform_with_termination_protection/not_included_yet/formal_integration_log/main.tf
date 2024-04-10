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

resource "formal_integration_log" "name" {
  name                   = "terraform-test-local-formal_integration_log-with-termination-protection"
  type                   = "splunk"
  splunk_access_token = "aaaaa"
  splunk_port = 443
  splunk_host     = "https://splunk.com"
  termination_protection = var.termination_protection
}
