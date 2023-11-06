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

resource "formal_encryption_key" "name" {
  cloud_region = "us-west-1"
  key_id       = "terraform-test-local-formal_key-with-termination-protection"
  key_name     = "terraform-test-local-formal_key-with-termination-protection"
}

resource "formal_key" "name" {
  cloud_region           = "eu-west-1"
  key_type               = "aws_kms"
  managed_by             = "customer_managed"
  name                   = "terraform-test-local-formal_key-with-termination-protection"
  key_id                 = formal_encryption_key.name.id
  termination_protection = var.termination_protection
}
