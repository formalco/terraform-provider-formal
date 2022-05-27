terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>1.0.9"
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

# Cloud Account Integration (for Managed Cloud)
provider "aws" {
  region     = "eu-west-1"
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

resource "formal_cloud_account" "integrated_aws_account" {
  cloud_account_name = "our aws account"
  cloud_provider     = "aws"
}

# NOTE: this stack must be deployed with an aws provider setup
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


# Datastore Sidecar
resource "formal_datastore" "my_datastore" {
  technology       = var.datastore_technology # postgres, redshift, snowflake
  name             = var.datastore_name
  hostname         = var.datastore_hostname
  port             = var.datastore_port
  deployment_type  = "managed"
  cloud_provider   = "aws"
  cloud_region     = var.datastore_region
  cloud_account_id = formal_cloud_account.integrated_aws_account.id
  customer_vpc_id  = var.customer_vpc_id
  fail_open        = false
  username         = var.datastore_username
  password         = var.datastore_password
}

# Role
resource "formal_role" "dior_the_data_scientist" {
  type       = "human"
  email      = "dior@acme.com"
  first_name = "dior"
  last_name  = "scientist"
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
  datastore_id = formal_datastore.my_datastore.datastore_id
  path         = "main.public.customers.email"
  key_storage  = "control_plane_only"
  key_id       = formal_key.encrypt_email_field_key.id
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
