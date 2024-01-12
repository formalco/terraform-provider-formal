terraform {
  required_providers {
    formal = {
      version = "~> 1.0.0"
      source  = "joinformal.com/local/formal"
    }
  }
}

provider "formal" {}

# Deprecated
# resource "formal_cloud_account" "name" {
# }

# Deprecated
# resource "formal_dataplane" "name" {
# }

# Deprecated
# resource "formal_dataplane_routes" "name" {
# }

resource "formal_datastore" "postgres1" {
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

# Deprecated
# resource "formal_default_field_encryption" "name" {
#   data_key_storage = "control_plane_only"
#   encryption_alg   = "aes_deterministic"
#   kms_key_id       = formal_encryption_key.name.id
# }

# resource "formal_encryption_key" "name" {
#   cloud_region = "us-west-1"
#   key_id       = "terraform-test-encryption-key-id"
#   key_name     = "terraform-test-encryption-key-local"
# }

# resource "formal_field_encryption" "name" {
#   alg          = "aes_deterministic"
#   datastore_id = formal_datastore.postgres1.id
#   key_id       = formal_encryption_key.name.id
#   key_storage  = "control_plane_only"
#   path         = "postgres.public.users.id"
# }

resource "formal_group" "name" {
  description = "terraform-test-group"
  name        = "terraform-test-group"
}

resource "formal_group_link_role" "name" {
  group_id = formal_group.name.id
  role_id  = formal_user.name.id
}

resource "formal_integration_app" "name" {
  name              = "terraform-test-integration-app"
  type              = "metabase"
  linked_db_user_id = "postgres"
  metabase_hostname = "https://metabase.com"
  metabase_password = "metabasepassword"
  metabase_username = "metabaseusername"
}

# resource "formal_integration_datahub" "name" {
#   active = true
#   api_key = "api_key_datahub_placeholder"
#   generalized_metadata_service_url = "https://datahub.com"
#   sync_direction = "bidirectional"
#   synced_entities = ["tags"]
# }

resource "formal_integration_external_api" "name" {
  auth_type = "basic"
  name      = "terraform-test-integration-external-api"
  type      = "custom"
  url       = "https://zendesk.com"
}

resource "formal_integration_log" "splunk" {
  name           = "terraform-test-integration-log-splunk"
  type           = "splunk"
  splunk_api_key = "aaaaa"
  splunk_url     = "https://splunk.com"
}

resource "formal_integration_log" "s3" {
  name                  = "terraform-test-integration-log-s3"
  type                  = "s3"
  aws_access_key_id     = "aaaaa"
  aws_access_key_secret = "aaaaa"
  aws_region            = "us-west-1"
  aws_s3_bucket_name    = "terraform-test-integration-log-s3"
}

# resource "formal_key" "name" {
#   cloud_region = "eu-west-1"
#   key_type     = "aws_kms"
#   managed_by   = "customer_managed"
#   name         = "terraform-test-key-aws-kms"
#   key_id       = formal_encryption_key.name.id
# }

resource "formal_native_role" "name" {
  datastore_id       = formal_datastore.postgres1.id
  native_role_id     = "postgres1"
  native_role_secret = "postgres1"
}

resource "formal_user" "name" {
  type = "machine"
  name = "terraform-test-user"
}

resource "formal_native_role_link" "name" {
  datastore_id         = formal_datastore.postgres1.id
  formal_identity_id   = formal_user.name.id
  formal_identity_type = "user"
  native_role_id       = formal_native_role.name.native_role_id
}

resource "formal_policy" "name" {
  active       = false
  description  = "terraform-test-policy"
  module       = <<EOT
package formal.v2

import future.keywords.if

pre_request := {
  "action": "block",
  "type": "block_with_formal_message"
} if {
  input.datastore.id == "${formal_datastore.postgres1.id}"
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

resource "formal_sidecar_datastore_link" "name" {
  datastore_id = formal_datastore.postgres1.id
  port         = 5432
  sidecar_id   = formal_sidecar.name.id
}
