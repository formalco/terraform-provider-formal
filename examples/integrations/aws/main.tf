terraform {
  required_providers {
    formal = {
      source = "formalco/formal"
    }
    aws = {
      source = "hashicorp/aws"
    }
  }
}

variable "formal_api_key" {
  type        = string
  description = "The Formal API key used to authenticate the Formal provider."
  sensitive   = true
}

variable "aws_region" {
  type        = string
  description = "AWS region for the Cloud Integration."
  default     = "us-east-1"
}

provider "formal" {
  api_key = var.formal_api_key
}

provider "aws" {
  region = var.aws_region
}

# 1. Bucket Formal delivers logs to. Formal only writes objects; the bucket must
#    exist, so we create it here.
resource "aws_s3_bucket" "formal_logs" {
  bucket = "formal-connector-logs"
}

# 2. Register the AWS Cloud Integration. allow_s3_access + s3_bucket_arn scope the
#    log delivery permission; the enable_*_autodiscovery flags control resource
#    discovery. Formal returns the CloudFormation template and parameters to apply.
resource "formal_integration_cloud" "aws" {
  name         = "aws-integration"
  cloud_region = var.aws_region

  aws {
    template_version = "latest"
    allow_s3_access  = true
    s3_bucket_arn    = "${aws_s3_bucket.formal_logs.arn}/*"
  }
}

# 3. Provision the Formal IAM role in your account via the CloudFormation stack
#    Formal generated, driven entirely by the integration's computed attributes.
resource "aws_cloudformation_stack" "formal" {
  name          = formal_integration_cloud.aws.aws_formal_stack_name
  template_body = formal_integration_cloud.aws.aws_template_body
  capabilities  = ["CAPABILITY_NAMED_IAM"]

  parameters = {
    FormalIntegrationId         = formal_integration_cloud.aws.id
    FormalIAMRoleId             = formal_integration_cloud.aws.aws_formal_iam_role
    FormalSNSTopicARN           = formal_integration_cloud.aws.aws_formal_pingback_arn
    EnableRDSAutodiscovery      = formal_integration_cloud.aws.aws_enable_rds_autodiscovery
    EnableRedshiftAutodiscovery = formal_integration_cloud.aws.aws_enable_redshift_autodiscovery
    EnableEKSAutodiscovery      = formal_integration_cloud.aws.aws_enable_eks_autodiscovery
    EnableEC2Autodiscovery      = formal_integration_cloud.aws.aws_enable_ec2_autodiscovery
    EnableECSAutodiscovery      = formal_integration_cloud.aws.aws_enable_ecs_autodiscovery
    EnableS3Autodiscovery       = formal_integration_cloud.aws.aws_enable_s3_autodiscovery
    AllowS3Access               = formal_integration_cloud.aws.aws_allow_s3_access
    S3BucketARN                 = formal_integration_cloud.aws.aws_s3_bucket_arn
  }
}

# 4. Deliver Formal logs to the bucket. Waits for the stack so the IAM role
#    exists before logs start flowing.
resource "formal_integration_log" "s3" {
  name = "aws-integration-logs"

  aws_s3 {
    cloud_integration_id = formal_integration_cloud.aws.id
    s3_bucket_name       = aws_s3_bucket.formal_logs.bucket
  }

  depends_on = [aws_cloudformation_stack.formal]
}
