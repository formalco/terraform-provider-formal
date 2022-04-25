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
resource "formal_datastore" "my_datastore" {
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
# resource "formal_role" "dior_the_data_scientist" {
#   type       = "human"
#   email      = "dior@acme.com"
#   first_name = "dior"
#   last_name  = "scientist"
# }


# Key
# resource "formal_key" "encrypt_email_field_key" {
#   name             = "email field encrypting key"
#   cloud_region     = "eu-west-1"
#   key_type         = "aws_kms"
#   managed_by       = "managed_cloud"
#   cloud_account_id = var.cloud_account_id
# }


# Field encryption 
# resource "formal_field_encryption" "encrypt_email_field" {
#   datastore_id = formal_datastore.my_datastore.datastore_id
#   path         = "main.public.customers.email"
#   key_storage  = "control_plane_only"
#   key_id       = formal_key.encrypt_email_field_key.id
# }



# Decrypt emails Policy
# resource "formal_policy" "decrypt_emails_policy" {
#   name        = "authorize emails"
#   description = "this policy, when linked to a role or group, allows them to decrypt emails."
#   module      = <<-EOF
# package formal.validator
# tags := {}

# decrypt { 
#     type := "column_name_equal"
#     input.path = "postgres.public.customers.email" 
# }	
# EOF
# }


# Link Policy to Role
# resource "formal_policy_link" "allow_decrypt_emails_for_user" {
#   type      = "role"
#   item_id   = formal_role.dior_the_data_scientist.id
#   policy_id = formal_policy.decrypt_emails_policy.id
# }
