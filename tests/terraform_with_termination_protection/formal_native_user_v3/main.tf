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
  hostname    = "terraform-test-local.formal-native-user-v3.with-termination-protection.com"
  name        = "terraform-test-local-formal-native-user-v3-with-termination-protection"
  technology  = "postgres"
  environment = "DEV"
  port        = 5432

  native_users_v3_enabled = true

  timeouts {
    create = "1m"
  }
}

resource "formal_native_user_v3" "name" {
  resource_id            = formal_resource.postgres1.id
  label                  = "postgres"
  termination_protection = var.termination_protection

  basic {
    username = "postgres"
    password {
      literal = "postgres"
    }
  }
}

resource "formal_resource_native_user_selection" "name" {
  resource_id = formal_resource.postgres1.id
  cel         = jsonencode(formal_native_user_v3.name.id)
}
