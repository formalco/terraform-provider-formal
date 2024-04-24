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

resource "formal_resource" "postgres1" {
  hostname                   = "terraform-test-local.formal-policy.with-termination-protection"
  name                       = "terraform-test-local-formal_policy-with-termination-protection"
  technology                 = "postgres"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_user" "human" {
  type = "human"
  name = "terraform-test-policy-human-user"
  first_name = "test2"
  last_name = "test2"
  email = "test@test-formal.com"
}

resource "formal_policy" "name" {
  description            = "terraform-test-policy"
  module                 = <<EOT
package formal.v2

import future.keywords.if

pre_request := {
  "action": "block",
  "type": "block_with_formal_message"
} if {
  input.datastore.id == "${formal_resource.postgres1.id}"
}
EOT
  name                   = "terraform-test-local-formal_policy-with-termination-protection"
  notification           = "none"
  owner       = formal_user.human.email
  status                 = "active"
  termination_protection = var.termination_protection
}
