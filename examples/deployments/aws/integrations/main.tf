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

provider "formal" {
  api_key = var.formal_api_key
}

provider "aws" {
  region = var.region
}

resource "aws_s3_bucket" "demo" {
  bucket        = "${var.name}-demo-integration"
  force_destroy = true
}

resource "formal_integration_cloud" "demo" {
  name         = "${var.name}-demo-integration"
  cloud_region = var.region

  aws {
    template_version              = "1.2.0"
    enable_eks_autodiscovery      = true
    enable_rds_autodiscovery      = true
    enable_redshift_autodiscovery = true
    allow_s3_access               = true
    s3_bucket_arn                 = aws_s3_bucket.demo.arn
  }
}

resource "time_sleep" "wait-aws" {
  depends_on      = [formal_integration_cloud.demo]
  create_duration = "20s"
}

resource "aws_cloudformation_stack" "demo" {
  depends_on    = [formal_integration_cloud.demo, time_sleep.wait-aws]
  name          = formal_integration_cloud.demo.aws_formal_stack_name
  template_body = formal_integration_cloud.demo.aws_template_body
  parameters = {
    FormalIntegrationId         = formal_integration_cloud.demo.id
    FormalIAMRoleId             = formal_integration_cloud.demo.aws_formal_iam_role
    FormalSNSTopicARN           = formal_integration_cloud.demo.aws_formal_pingback_arn
    EnableEKSAutodiscovery      = formal_integration_cloud.demo.aws_enable_eks_autodiscovery
    EnableRDSAutodiscovery      = formal_integration_cloud.demo.aws_enable_rds_autodiscovery
    EnableRedshiftAutodiscovery = formal_integration_cloud.demo.aws_enable_redshift_autodiscovery
    AllowS3Access               = formal_integration_cloud.demo.aws_allow_s3_access
    S3BucketARN                 = formal_integration_cloud.demo.aws_s3_bucket_arn
  }
  capabilities = ["CAPABILITY_NAMED_IAM"]
}

resource "formal_integration_log" "demo" {
  name = "${var.name}-demo-integration"

  aws_s3 {
    s3_bucket_name       = aws_s3_bucket.demo.bucket
    cloud_integration_id = formal_integration_cloud.demo.id
  }
}
