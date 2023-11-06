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

resource "formal_datastore" "postgres1" {
  hostname                   = "terraform-test-local.formal-integration-log-link.with-termination-protection"
  name                       = "terraform-test-local-formal_integration_log_link-with-termination-protection"
  technology                 = "postgres"
  db_discovery_job_wait_time = "1m"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_integration_log" "name" {
  name           = "terraform-test-local-formal_integration_log_link-with-termination-protection"
  type           = "splunk"
  splunk_api_key = "aaaaa"
  splunk_url     = "https://splunk.com"
}

resource "formal_integration_log_link" "name" {
  integration_id         = formal_integration_log.name.id
  datastore_id           = formal_datastore.postgres1.id
  termination_protection = var.termination_protection
}
