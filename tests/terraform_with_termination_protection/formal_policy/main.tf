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

resource "formal_datastore" "postgres1" {
  hostname                   = "terraform-test-local.formal-policy.with-termination-protection"
  name                       = "terraform-test-local-formal_policy-with-termination-protection"
  technology                 = "postgres"
  db_discovery_job_wait_time = "1m"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_policy" "name" {
  active                 = true
  description            = "terraform-test-policy"
  module                 = <<EOT
package formal.v2

import future.keywords.if

pre_request := {
  "action": "block",
  "type": "block_with_formal_message"
} if {
  input.datastore.id == "${formal_datastore.postgres1.id}"
}
EOT
  name                   = "terraform-test-local-formal_policy-with-termination-protection"
  notification           = "none"
  owners                 = ["farid@joinformal.com"]
  status                 = "active"
  termination_protection = var.termination_protection
}
