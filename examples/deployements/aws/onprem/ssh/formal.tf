terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>4.0.0"
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
  technology         = "ssh"
  hostname    = var.hostname
}

resource "formal_resource" "instance_1" {
  technology = "ssh"
  name       = var.name
  hostname   = aws_instance.main.public_dns
  port       = 22
}

resource "formal_native_user" "main_instance_1" {
  resource_id       = formal_resource.instance_1.id
  native_user_id     = aws_iam_access_key.example_access_key.id
  native_user_secret = aws_iam_access_key.example_access_key.secret
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

resource "formal_sidecar_resource_link" "link_1" {
  resource_id = formal_resource.instance_1.id
  sidecar_id   = formal_sidecar.main.id
  port         = 22
}
