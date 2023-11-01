terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 3.2.6"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key # you can also use env variable FORMAL_API_KEY
}

resource "formal_policy" "decrypt" {
  name         = "decrypt"
  description  = "decrpyt any columns that has the name encrypted-col."
  owners       = ["john@formal.com"]
  notification = "none"
  status       = "active"
  active       = true
  module       = <<-EOF
package formal.v2

import future.keywords.if
import future.keywords.in

post_request := { "action": "decrypt", "columns": columns } if {
    columns := [col | col := input.row[_]; col["name"] == "encrypted-col"]
}
EOF
}


resource "formal_policy" "mask_emails" {
  name         = "mask-email"
  description  = "Mask any column that has the email data labal email_address."
  owners       = ["john@company.com"]
  notification = "consumer"
  status       = "active"
  active       = true
  module       = <<-EOF
package formal.v2

import future.keywords.if

post_request := { "action": "mask", "type": "redact.partial", "sub_type": "email_mask_username", "columns": columns, "typesafe": "fallback_to_default" } if {
    columns := [col | col := input.columns[_]; col["data_label"] == "email_address";]
}
EOF
}

resource "formal_policy" "row_level_hashing" {
  name         = "test-row-level-hashing-eu"
  description  = "hash every row that has the eu column set to true."
  owners       = ["john@company.com"]
  notification = "all"
  status       = "active"
  active       = true
  module       = <<-EOF
package formal.v2

import future.keywords.if

post_request := { "action": "mask", "type": "hash.with_salt", "columns": input.columns } if {
    colValue := [col | col := input.row[_]; col["name"] == "eu"; col["value"] == true]
    count(colValue) > 0
}
EOF
}

resource "formal_policy" "block_db_with_formal_message" {
  name         = "block_db_with_formal_message"
  description  = "this policy block connection to sidecar based on the name of db and drop the connection with a formal message"
  owners       = ["john@company.com"]
  notification = "all"
  active       = true
  status       = "active"
  module       = <<-EOF
package formal.v2

import future.keywords.if
import future.keywords.in

default session := { "action": "block", "type": "block_with_formal_message" }

session := { "action": "allow", "reason": "the policy is blocking request" } if {
	input.db_name == "main"
	"USAnalyst" in input.user.groups
	input.datastore.technology == "postgres"
}
EOF
}


resource "formal_policy" "http_pre_request_name_hash" {
  name         = "http_pre_request_name_hash"
  description  = "this policy hash every names in body request of HTTP requests"
  owners       = ["john@company.com"]
  notification = "all"
  status       = "active"
  active       = true
  module       = <<-EOF
package formal.v2

import future.keywords.if

pre_request := { "action": "mask", "type": "hash.with_salt", "columns": columns } if {
    columns := [col | col := input.row[_]; col["data_label"] == "name"]
}
EOF
}