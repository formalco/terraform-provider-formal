terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.0.23"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~>3.0"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

provider "aws" {
  region     = var.region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

# Cloud Account Integration Demo (for Managed Cloud) 
# Note the specified aws_cloud_region is the region the CloudFormation stack will be deployed in, which must be deployed with an aws provider setup for eu-west-1, us-east-1, or us-east-2.
resource "formal_cloud_account" "integrated_aws_account" {
  cloud_account_name = var.name
  cloud_provider     = "aws"
  aws_cloud_region   = var.region
}

# Declare the CloudFormation stack
resource "aws_cloudformation_stack" "integrate_with_formal" {
  name = formal_cloud_account.integrated_aws_account.aws_formal_stack_name
  parameters = {
    FormalID          = formal_cloud_account.integrated_aws_account.aws_formal_id
    FormalIamRole     = formal_cloud_account.integrated_aws_account.aws_formal_iam_role
    FormalHandshakeID = formal_cloud_account.integrated_aws_account.aws_formal_handshake_id
    FormalPingbackArn = formal_cloud_account.integrated_aws_account.aws_formal_pingback_arn
  }
  template_body = formal_cloud_account.integrated_aws_account.aws_formal_template_body
  capabilities  = ["CAPABILITY_NAMED_IAM"]
}

resource "formal_dataplane" "main" {
  name               = var.name
  cloud_region       = var.region
  cloud_account_id   = formal_cloud_account.integrated_aws_account.id
  availability_zones = 2

  depends_on = [
    formal_cloud_account.integrated_aws_account,
    aws_cloudformation_stack.integrate_with_formal
  ]
}

resource "formal_datastore" "main" {
  technology              = "snowflake"
  name                    = var.name
  hostname                = var.snowflake_hostname
  port                    = var.snowflake_port
  default_access_behavior = "allow"
}

resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "managed"
  cloud_provider     = "aws"
  cloud_region       = var.region
  cloud_account_id   = formal_cloud_account.integrated_aws_account.id
  fail_open          = false
  dataplane_id       = formal_dataplane.main.id
  global_kms_decrypt = true
  network_type       = "internet-facing" //internal, internet-and-internal
  datastore_id       = formal_datastore.main.id
}

# Native Role
resource "formal_native_role" "main_postgres" {
  datastore_id       = formal_datastore.main.id
  native_role_id     = var.snowflake_username
  native_role_secret = var.snowflake_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}
