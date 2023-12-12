terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

variable "key_type" {
  type        = string
}

resource "formal_encryption_key" "name" {
  cloud_region = "us-west-1"
  key_id       = "terraform-test-local-formal_key-with-termination-protection"
  key_name     = "terraform-test-local-formal_key-with-termination-protection"
}

resource "formal_key" "name" {
  cloud_region           = "eu-west-1"
  key_type               = var.key_type
  managed_by             = "customer_managed"
  name                   = "terraform-test-local-formal_key-with-termination-protection"
  key_id                 = formal_encryption_key.name.id
}
