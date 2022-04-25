terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {
  client_id  = var.client_id
  secret_key = var.secret_key
}


# Datastore
resource "formal_datastore" "a_sensitive_datastore_managed_sidecar" {
  technology       = var.datastore_technology # postgres, redshift, snowflake
  name             = var.datastore_name
  hostname         = var.datastore_hostname
  port             = var.datastore_port
  deployment_type  = "managed"
  cloud_provider   = "aws"
  cloud_region     = var.datastore_region
  cloud_account_id = var.cloud_account_id
  customer_vpc_id  = var.customer_vpc_id
  fail_open        = false
  username         = var.datastore_username
  password         = var.datastore_password
}



# Role
resource "formal_role" "brian_managed_consumer" {
  type       = "human"
  email      = "brian+managedconsumer@joinformal.com"
  first_name = "brian"
  last_name  = "is in the kitchen"
}


# Key
resource "formal_key" "encrypt_email_field_key" {
  name             = "key to be used for encrypting email field"
  cloud_region     = "eu-west-1"
  key_type         = "aws_kms"
  managed_by       = "managed_cloud"
  cloud_account_id = var.cloud_account_id
}


# Field encryption 
resource "formal_field_encryption" "encrypt_email_field" {
  datastore_id = formal_datastore.a_sensitive_datastore_managed_sidecar.datastore_id
  path         = "postgres.public.customers.email"
  key_storage  = "control_plane_only"
  key_id       = formal_key.encrypt_email_field_key.id
}



# Decrypt emails Policy for role/group
resource "formal_policy" "decrypt_emails_policy" {
  name        = "authorize emails"
  description = "this policy, when linked to a role or group, allows them to decrypt emails."
  module      = <<-EOF
package formal.validator
tags := {}

decrypt { 
    type := "column_name_equal"
    input.path = "postgres.public.customers.email" 
}	
EOF
}


# Policy-Link Role
resource "formal_policy_link" "allow_decrypt_emails_for_brian" {
  type      = "role"
  item_id   = formal_role.brian_managed_consumer.id
  policy_id = formal_policy.decrypt_emails_policy.id
}


















resource "formal_key" "email_field_encryption_key" {
  cloud_region = "eu-west-3"
  key_type     = "aws_kms"
  managed_by   = "saas_managed"
  name         = "key to encrypt email field"
}
