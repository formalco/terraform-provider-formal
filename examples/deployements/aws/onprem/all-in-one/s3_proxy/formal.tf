terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.2.3"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "onprem"
  technology         = "s3"
  global_kms_decrypt = false
  formal_hostname    = var.s3_sidecar_hostname
}

resource "formal_datastore" "main" {
  technology = "s3"
  name       = "${var.name}-datastore"
  hostname   = var.s3_hostname
  port       = var.main_port
}

resource "formal_sidecar_datastore_link" "main" {
  datastore_id = formal_datastore.main.id
  sidecar_id   = formal_sidecar.main.id
  port         = 0
}

# Native Role
resource "formal_native_role" "main_s3" {
  datastore_id       = formal_datastore.main.id
  native_role_id     = var.iam_user_key_id
  native_role_secret = var.iam_user_secret_key
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

resource "formal_key" "encryption_key" {
  cloud_region = "eu-west-1"
  key_type     = "aws_kms"
  managed_by   = "customer_managed"
  name         = "formal-s3-demo-key"
  key_id       = aws_kms_key.field_encryption.id
}