terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.0.7"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "4.15.1"
    }
  }
}

provider "formal" {
  client_id  = var.client_id
  secret_key = var.secret_key
}

provider "aws" {
  region     = "us-east-1"
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}


# Cloud Account Integration Demo (for Managed Cloud) 

# Note the specified aws_cloud_region is the region the CloudFormation stack will be deployed in, which must be deployed with an aws provider setup for eu-west-1, us-east-1, or us-east-2.
resource "formal_cloud_account" "integrated_aws_account" {
  cloud_account_name = "our aws account"
  cloud_provider     = "aws"
  aws_cloud_region   = "us-east-1"
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

# ==============================

# Dataplane
resource "formal_dataplane" "my_dataplane" {
  name               = "my_dataplane"
  cloud_region       = var.datastore_region
  cloud_account_id   = var.cloud_account_id
  availability_zones = 3
  vpc_peering        = true
}


# Sidecar, Managed Example
# For Onprem Sidecars, you can access the Control Plane TLS Certificate variable using: formal_sidecar.my_onprem_datastore.formal_control_plane_tls_certificate
resource "formal_sidecar" "my_sidecar" {
  name               = var.datastore_name
  cloud_provider     = "aws"
  cloud_region       = var.datastore_region
  cloud_account_id   = formal_cloud_account.integrated_aws_account.id
  deployment_type    = "managed"
  fail_open          = false
  network_type       = "internet-facing"
  dataplane_id       = var.dataplane_id
  global_kms_decrypt = false
  datastore_id       = formal_datastore.pg_datastore.id
  version            = ""
}

# Datastore
resource "formal_datastore" "pg_datastore" {
  name                        = var.datastore_name
  hostname                    = var.datastore_hostname
  technology                  = "postgres"
  port                        = var.datastore_port
  health_check_db_name        = "postgres"
  default_access_behavior     = "allow"
  db_discovery_job_wait_time  = "3h"
  db_discovery_native_role_id = "postgres"
}

# Native Role
resource "formal_native_role" "db_role" {
  datastore_id       = formal_datastore.pg_datastore.id
  native_role_id     = "postgres"
  native_role_secret = var.native_role_secret
  use_as_default     = false // per sidecar, exactly one native role must be marked as the default.
}


# Role
resource "formal_role" "dior_the_data_scientist" {
  type       = "human"
  email      = "dior@acme.com"
  first_name = "dior"
  last_name  = "scientist"
}


# Link Native Role to the above Role
resource "formal_native_role_link" "dior_uses_db_role" {
  datastore_id         = formal_native_role.db_role.datastore_id
  native_role_id       = formal_native_role.db_role.native_role_id
  formal_identity_id   = formal_role.dior_the_data_scientist.id
  formal_identity_type = "role"
}

# Key to be used for Field Encryption
resource "formal_key" "encrypt_email_field_key" {
  name             = "email field encrypting key"
  cloud_region     = "us-east-1"
  key_type         = "aws_kms"
  managed_by       = "managed_cloud"
  cloud_account_id = formal_cloud_account.integrated_aws_account.id
}


# Specify a Field Encryption 
resource "formal_field_encryption" "encrypt_email_field" {
  datastore_id = formal_sidecar.my_datastore.datastore_id
  path         = "main.public.customers.email"
  key_storage  = "control_plane_only"
  key_id       = formal_key.encrypt_email_field_key.id
  alg          = "aes_random"
}

# Default Field Encryption
resource "formal_default_field_encryption" "encrypt_email_field" {
  data_key_storage = "control_plane_and_with_data"
  kms_key_id       = formal_key.encrypt_email_field_key.id
  encryption_alg   = "aes_random"
}

# An "Allow Decrypt emails" Policy
resource "formal_policy" "decrypt_emails_policy" {
  name        = "authorize emails"
  description = "this policy, when linked to a role or group, allows them to decrypt emails."
  module      = <<-EOF
package formal.validator
tags := {}

decrypt { 
    type := "column_name_equal"
    input.path = "main.public.customers.email" 
}	
EOF
}


# Link above Policy to a Role
resource "formal_policy_link" "allow_decrypt_emails_for_user" {
  type      = "role"
  item_id   = formal_role.dior_the_data_scientist.id
  policy_id = formal_policy.decrypt_emails_policy.id
}



# A sample "Mask email usernames" Policy. Note this is different from a Field Encryption. This is applied to a specific datastore's 'email' field.
resource "formal_policy" "mask_email_policy" {
  name        = "mask emails"
  description = "this policy masks email usernames"
  module      = <<-EOF
package formal.validator
tags := {"email_address"}

mask[action] {
    type := "tag_detected"
    tags[input.tag]
    action := "email_mask_username"
EOF
}
