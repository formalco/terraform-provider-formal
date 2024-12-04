terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.47.0"
    }
    formal = {
      source  = "formalco/formal"
      version = "4.1.0"
    }
  }
  required_version = ">= 0.14.9"
}

provider "formal" {}

provider "aws" {
  region = "eu-west-1"
}

resource "formal_integration_cloud" "demo" {
  name         = "tf-demo-integration"
  type         = "aws"
  cloud_region = "eu-west-1"
}

resource "aws_cloudformation_stack" "demo" {
  name          = formal_integration_cloud.demo.aws_formal_stack_name
  template_body = formal_integration_cloud.demo.aws_template_body
  parameters = {
    FormalIAMRoleId     = formal_integration_cloud.demo.aws_formal_iam_role
    FormalSNSTopicARN   = formal_integration_cloud.demo.aws_formal_pingback_arn
    FormalIntegrationId = formal_integration_cloud.demo.id
  }
  capabilities = ["CAPABILITY_NAMED_IAM"]
}

resource "aws_s3_bucket" "demo" {
  bucket        = "tf-demo-integration"
  force_destroy = true
}

resource "formal_integration_log" "demo" {
  name = "tf-demo-integration"

  aws_s3 {
    s3_bucket_name       = aws_s3_bucket.demo.bucket
    cloud_integration_id = formal_integration_cloud.demo.id
  }
}
