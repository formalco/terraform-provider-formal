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
  hostname                   = "terraform-test-local.formal-native-user.with-termination-protection.com"
  name                       = "terraform-test-local-formal_native_user-with-termination-protection"
  technology                 = "postgres"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_native_user" "name" {
  resource_id           = formal_resource.postgres1.id
  native_user_id         = "postgres"
  native_user_secret     = "postgres"
  use_as_default         = true
  termination_protection = var.termination_protection
}
