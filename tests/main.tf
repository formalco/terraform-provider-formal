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
  db_discovery_job_wait_time = "6h"
  environment                = "DEV"
  port                       = 5432
  timeouts {
    create = "1m"
  }
}

resource "formal_group" "name" {
  description = "terraform-test-group"
  name        = "terraform-test-group"
}

resource "formal_group_link_role" "name" {
  group_id = formal_group.name.id
  user_id  = formal_user.name.id
}

resource "formal_integration_bi" "name" {
  name              = "terraform-test-integration-app"
  type              = "metabase"
  linked_db_user_id = "postgres"
  metabase_hostname = "https://metabase.com"
  metabase_password = "metabasepassword"
  metabase_username = "metabaseusername"
}

resource "formal_integration_data_catalog" "name" {
  type = "datahub"
  api_key = "api_key_datahub_placeholder"
  generalized_metadata_service_url = "https://datahub.com"
  sync_direction = "bidirectional"
  synced_entities = ["tags"]
}

resource "formal_integration_log" "splunk" {
  name           = "terraform-test-integration-log-splunk"
  type           = "splunk"
  splunk_access_token = "aaaaa"
  splunk_port = 443
  splunk_host     = "https://splunk.com"
}

resource "formal_integration_log" "s3" {
  name                  = "terraform-test-integration-log-s3"
  type                  = "s3"
  aws_access_key_id     = "aaaaa"
  aws_access_key_secret = "aaaaa"
  aws_region            = "us-west-1"
  aws_s3_bucket_name    = "terraform-test-integration-log-s3"
}

resource "formal_native_role" "name" {
  resource_id       = formal_resource.postgres1.id
  native_role_id     = "postgres1"
  native_role_secret = "postgres1"
}

resource "formal_user" "name" {
  type = "machine"
  name = "terraform-test-user"
}

resource "formal_native_role_link" "name" {
  resource_id         = formal_resource.postgres1.id
  formal_identity_id   = formal_user.name.id
  formal_identity_type = "user"
  native_role_id       = formal_native_role.name.native_role_id
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
  owners       = ["farid@joinformal.com"]
  status       = "draft"
}

resource "formal_satellite" "name" {
  name = "terraform-test-satellite"
}

resource "formal_sidecar" "name" {
  deployment_type    = "onprem"
  global_kms_decrypt = false
  name               = "terraform-test-sidecar"
  technology         = "postgres"
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
  active = true
  name = "name"
  description = "description"
  included_paths = ["main.path"]
  excluded_paths = ["main.path2"]
  owners = [{
      object_type = "firstObjectType"
      object_id = "firstObjectId"
  }]
}

resource "formal_tracker" "name" {
  resource_id = formal_resource.postgres1.id
  path = "dummy.path"
  allow_clear_text_value = true
}

resource "formal_data_discovery" "name" {
  resource_id = formal_resource.postgres1.id
  native_user_id = formal_native_role.name.native_role_id
  schedule = "12h"
  database = "main"
  deletion_policy = "mark_for_deletion"
}