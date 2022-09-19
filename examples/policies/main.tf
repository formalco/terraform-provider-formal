terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>1.0.27"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "formal" {
  client_id  = var.formal_client_id
  secret_key = var.formal_secret_key
}

resource "formal_policy" "mask_emails" {
  name        = "mask-email"
  description = "this policy, when linked to a role or group, masks the emails' username."
  module      = <<-EOF
package formal.validator
tags := {"email_address"}

mask[[action, typesafe]] {
  type := "tag_detected"
  tags[input.tag]        
  action := "email_mask_username"
  typesafe := ""
}
EOF
}

resource "formal_policy_link" "masked_email" {
  type      = "group"
  item_id   = var.group_id
  policy_id = formal_policy.mask_emails.id
}

resource "formal_policy" "row_level_hashing" {
  name        = "test-row-level-hashing-eu"
  description = "this policy, when linked to a role or group, hash the first name of any row that has an eu value."
  module      = <<-EOF
package formal.validator
tags := {}

mask[action] {
  type := "row_equal"
  input.row_value = true
  input.column_name = "eu"
  action := {"first_name": "hash"}
}
EOF
}

resource "formal_policy_link" "row_level_hashing" {
  type      = "group"
  item_id   = var.group_id
  policy_id = formal_policy.row_level_hashing.id
}

/*
 * Blocking Policies
 */

// Block connections with Formal Message 
resource "formal_policy" "block_db_with_formal_message" {
  name        = "block_db_with_formal_message"
  description = "this policy block connection to sidecar based on the name of db and throw an error message about Formal."
  module      = <<-EOF
package formal.validator

block[action] {
  input.db_name = "main"
  action := "block_with_formal_message"
}
EOF
}

// Block connections to the datastore silently
resource "formal_policy" "block_silently" {
  name        = "block_db_with_formal_message"
  description = "this policy block connection to sidecar based on the name of db and drop the connection silently."
  module      = <<-EOF
package formal.validator

block[action] {
  input.db_name = "main"
  action := "block_silently"
}
EOF
}

// Block connections to the datastore with a fake error
resource "formal_policy" "block_with_fake_error" {
  name        = "block_db_with_formal_message"
  description = "this policy block connection to sidecar based on the name of db and drop the connection with a fake error."
  module      = <<-EOF
package formal.validator

block[action] {
  input.db_name = "main"
  action := "block_with_fake_error"
}
EOF
}

resource "formal_policy_link" "block" {
  type      = "group"
  item_id   = var.group_id
  policy_id = formal_policy.block_with_fake_error.id
}
