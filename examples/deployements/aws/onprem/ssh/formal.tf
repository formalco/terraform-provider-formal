terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.4.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~>3.0"
    }
  }
}

provider "aws" {
  region = var.region
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "onprem"
  technology         = "ssh"
  fail_open          = false
  global_kms_decrypt = false
  formal_hostname    = var.hostname
}

resource "formal_datastore" "instance_1" {
  technology = "ssh"
  name       = var.name
  hostname   = aws_instance.main.public_dns
  port       = 22
}

resource "formal_native_role" "main_instance_1" {
  datastore_id       = formal_datastore.instance_1.id
  native_role_id     = aws_iam_access_key.example_access_key.id
  native_role_secret = aws_iam_access_key.example_access_key.secret
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

resource "formal_sidecar_datastore_link" "link_1" {
  datastore_id = formal_datastore.instance_1.id
  sidecar_id   = formal_sidecar.main.id
  port         = 22
}
