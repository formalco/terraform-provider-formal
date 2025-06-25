terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>4.0.0"
    }
  }

  required_version = ">= 0.14.9"
}

provider "formal" {
  api_key = var.formal_api_key
}

resource "formal_sidecar" "main" {
  name       = var.name
  technology = "postgres"
  hostname   = var.postgres_sidecar_hostname
}

resource "formal_resource" "main" {
  technology = "postgres"
  name       = var.name
  hostname   = var.postgres_hostname
  port       = var.main_port
}

# Extra hostname, used to access the same resource
# for example:
# postgres_extra_name=reader_endpoint
# postgres_extra_hostname=xxx.cluster-ro-xxx.xxx.rds.amazonaws.com
# you can then access the resource `psql -h <var.postgres_sidecar_hostname> -p 5432 -d <dbname>@<var.name>@<var.postgres_extra_name>`
resource "formal_resource_hostname" "name" {
  resource_id = formal_resource.main.id
  hostname    = var.postgres_extra_hostname
  name        = var.postgres_extra_name
}

resource "formal_sidecar_resource_link" "main" {
  resource_id = formal_resource.main.id
  sidecar_id  = formal_sidecar.main.id
  port        = 5432
}

# Native Role
resource "formal_native_user" "main_postgres" {
  resource_id     = formal_resource.main.id
  type            = "password"
  username        = var.postgres_username
  username_is_env = false
  password        = var.postgres_password
  password_is_env = false
  use_as_default  = true // per sidecar, exactly one native role must be marked as the default.
}

resource "formal_native_user" "main_postgres_extra" {
  resource_id     = formal_resource.main.id
  type            = "password"
  username        = var.postgres_extra_username
  username_is_env = false
  password        = var.postgres_extra_password
  password_is_env = false
}

resource "formal_native_user_link" "main" {
  formal_identity_id   = formal_resource.main.id
  formal_identity_type = "resource_hostname"
  native_user_id       = formal_native_user.main_postgres_extra.id
}
