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

resource "formal_encryption_key" "name" {
  cloud_region           = "us-west-1"
  key_id                 = "terraform-test-local-encryption_key-with-termination-protection"
  key_name               = "terraform-test-local-encryption_key-with-termination-protection"
  termination_protection = var.termination_protection
}
