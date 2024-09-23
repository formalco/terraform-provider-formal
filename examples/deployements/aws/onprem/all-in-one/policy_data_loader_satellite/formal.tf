terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "4.0.17"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_satellite" "main" {
  name = "demo-policy-data-loader"
}

resource "formal_policy_data_loader" "zendesk_loader" {
  name            = "Load Zendesk tickets and related users"
  description     = "Use Zendesk API to fetch active tickets and their related users: submitters, requesters, assignees."
  key             = "zendesk_tickets"
  status          = "active"
  worker_schedule = "*/30 * * * * *"
  worker_runtime  = "python3.11"
  worker_code     = file("${path.module}/zendesk_loader.py")
}
