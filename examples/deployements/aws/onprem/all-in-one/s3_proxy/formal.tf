terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>4.0.0"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name       = var.name
  technology = "s3"
  hostname   = var.s3_sidecar_hostname
}

resource "formal_resource" "main" {
  technology = "s3"
  name       = "${var.name}-resource"
  hostname   = var.s3_hostname
  port       = var.main_port
}

resource "formal_sidecar_resource_link" "main" {
  resource_id = formal_resource.main.id
  sidecar_id  = formal_sidecar.main.id
  port        = 0
}

# Native Role
resource "formal_native_user" "main_s3" {
  resource_id        = formal_resource.main.id
  native_user_id     = var.iam_user_key_id
  native_user_secret = var.iam_user_secret_key
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

# resource "formal_key" "encryption_key" {
#   cloud_region = "eu-west-1"
#   key_type     = "aws_kms"
#   managed_by   = "customer_managed"
#   name         = "formal-s3-demo-key"
#   key_id       = aws_kms_key.field_encryption.id
# }