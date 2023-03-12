terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 3.0.9"
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

resource "formal_policy" "decrypt" {
  name        = "decrypt"
  description = "this policy, when linked to a role or group, allows them to decrypt the demo table."

  module = <<-EOF
package formal.validator

decrypt { 
    input.path = "main.public.demo_field_encryption.name" 
}	
EOF
}


resource "formal_policy" "mask_emails" {
  name        = "mask-email"
  description = "this policy, when linked to a role or group, masks the emails' username."
  module      = <<-EOF
package formal.validator

mask[[action, typesafe]] {
  input.tag = "email_address"  
  action := "email_mask_username"
  typesafe := ""
}
EOF
}

resource "formal_policy" "mask_emails_typesafe_fallback_to_default" {
  name        = "mask-email"
  description = "this policy, masks the emails' username is type safe and fallback to default."
  module      = <<-EOF
package formal.validator

mask[[action, typesafe]] {
  input.tag = "email_address"  
  action := "email_mask_username"
  typesafe := "fallback_to_default"
}
EOF
}

resource "formal_policy" "mask_emails_typesafe_fallback_to_null" {
  name        = "mask-email"
  description = "this policy, masks the emails' username is type safe and fallback to null."
  module      = <<-EOF
package formal.validator

mask[[action, typesafe]] {
  input.tag = "email_address"  
  action := "email_mask_username"
  typesafe := "fallback_to_null"
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

mask[action] {
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

resource "formal_policy" "block_for_ip_address" {
  name        = "block_db_with_formal_message"
  description = "this policy block connection to sidecar if the ip address is 127.0.0.1 and drop the connection with a fake error."
  module      = <<-EOF
package formal.validator

block[action] {
  input.ip_address = "127.0.0.1"
  action := "block_with_fake_error"
}
EOF
}

resource "formal_policy" "allow_for_ip_address" {
  name        = "allow_for_ip_address"
  description = "this policy allow connection to sidecar if the ip address is 127.0.0.1."
  module      = <<-EOF
package formal.validator

allow {
  input.ip_address = "127.0.0.1"
}
EOF
}

resource "formal_policy" "block_on_sunday" {
  name        = "block_on_weekends"
  description = "this policy block connection to sidecar on sunday."
  module      = <<-EOF
package formal.validator

block[action] {
  time.weekday(time.now_ns()) = "Sunday"
  action := "block_with_formal_message"
}
EOF
}

resource "formal_policy" "block_on_weekends" {
  name        = "block_on_weekends"
  description = "this policy block connection to sidecar on weekends."
  module      = <<-EOF
package formal.validator

block[action] {
	current_day := time.weekday(time.now_ns())
  weekend := {"Sunday", "Saturday"}
    
  weekend[current_day]

  action := "block_with_formal_message"
}
EOF
}

resource "formal_policy_link" "block" {
  type      = "group"
  item_id   = var.group_id
  policy_id = formal_policy.block_with_fake_error.id
}
