terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

resource "formal_resource" "postgres1" {
  hostname                   = "terraform-test-postgres1"
  name                       = "terraform-test-postgres1"
  technology                 = "postgres"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_integration_bi" "name" {
  name              = "terraform-test-integration-app"
  metabase {
    hostname = "metabase.com"
    password = "metabasepassword"
    username = "metabaseusername"
  }
}

resource "formal_integration_log" "datadog" {
  name           = "terraform-test-integration-log-datadog"
  datadog {
    site = "test.com"
    api_key = "test"
    account_id = "test"
  }
}

resource "formal_integration_log" "splunk" {
  name           = "terraform-test-integration-log-splunk"
  splunk {
    access_token = "aaaaa"
    port = 443
    host     = "splunk.com"
  }
}

resource "formal_integration_log" "s3" {
  name                  = "terraform-test-integration-log-s3"
  aws_s3 {
    access_key_id     = "aaaaa"
    access_key_secret = "aaaaa"
    region            = "us-west-1"
    s3_bucket_name    = "terraform-test-integration-log-s3"
  }
}

resource "formal_integration_mfa" "duo" {
  name = "test"
  duo {
    api_hostname     = "key.com"
    secret_key = "key"
    integration_key = "key"
  }
}

resource "formal_native_user" "name" {
  resource_id       = formal_resource.postgres1.id
  native_user_id     = "postgres1"
  native_user_secret = "postgres1"
}

resource "formal_user" "name" {
  type = "machine"
  name = "terraform-test-user"
}

resource "formal_user" "human" {
  type = "human"
  name = "terraform-test-human-user"
  first_name = "test2"
  last_name = "test2"
  email = "test@test-formal.com"
  admin = true
}

resource "formal_native_user_link" "name" {
  formal_identity_id   = formal_user.name.id
  formal_identity_type = "user"
  native_user_id       = formal_native_user.name.id
}

resource "formal_policy" "name" {
  description  = "terraform-test-policy"
  module       = <<EOT
package formal.v2

import future.keywords.if

pre_request := {
  "action": "block",
  "type": "block_with_formal_message"
} if {
  input.datastore.id == "${formal_resource.postgres1.id}"
}
EOT
  name         = "terraform-test-policy"
  notification = "none"
  owners       = [formal_user.human.email]
  status       = "draft"
}

resource "formal_satellite" "name" {
  name = "terraform-test-satellite"
  termination_protection = false
}

resource "formal_sidecar" "name" {
  name               = "terraform-test-sidecar"
  hostname               = "test.com"
  technology         = "postgres"
  termination_protection = false
}

resource "formal_sidecar_resource_link" "name" {
  resource_id = formal_resource.postgres1.id
  port         = 5432
  sidecar_id   = formal_sidecar.name.id
}

resource "formal_resource_health_check" "name" {
  resource_id = formal_resource.postgres1.id
  database_name = "test-1"
}

resource "formal_data_domain" "name" {
  name = "name"
  description = "description"
  included_paths = ["main.path"]
  excluded_paths = ["main.path2"]
    dynamic "owners" {
    for_each = [
      { object_type = "firstObjectType", object_id = "firstObjectId" }
    ]
    content {
      object_type = owners.value.object_type
      object_id = owners.value.object_id
    }
  }
}

resource "formal_tracker" "name" {
  resource_id = formal_resource.postgres1.id
  path = "dummy.path"
  allow_clear_text_value = true
}

resource "formal_data_discovery" "name" {
  resource_id = formal_resource.postgres1.id
  native_user_id = formal_native_user.name.id
  schedule = "12h"
  deletion_policy = "mark_for_deletion"
}

resource "formal_group" "name" {
  description = "terraform-test-group"
  name        = "terraform-test-group"
}

resource "formal_group_user_link" "name" {
  group_id = formal_group.name.id
  user_id  = formal_user.name.id
}

resource "formal_policy_external_data_loader" "name" {
  name = "test-external-data-loader-2"
  host = "formal.zendesk.com"
  port = 443
  auth_type = "basic"
  basic_auth_username = "basic"
  basic_auth_password = "basic"
}
