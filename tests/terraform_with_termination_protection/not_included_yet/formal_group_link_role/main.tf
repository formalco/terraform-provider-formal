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

resource "formal_group" "name" {
  description = "terraform-test-local-formal_group_link_role-with-termination-protection"
  name        = "terraform-test-local-formal_group-with-termination-protection"
}

resource "formal_user" "name" {
  type = "machine"
  name = "terraform-test-local-formal_group_link_role-with-termination-protection"
}

resource "formal_group_link_user" "name" {
  group_id               = formal_group.name.id
  user_id                = formal_user.name.id
  termination_protection = var.termination_protection
}
