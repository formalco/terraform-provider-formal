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
  resource_id        = formal_resource.main.id
  native_user_id     = var.postgres_username
  native_user_secret = var.postgres_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}
