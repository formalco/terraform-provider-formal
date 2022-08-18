---
layout: ""
page_title: "Provider: Formal"
subcategory: ""
description: |-
  Use the Formal Terraform Provider to interact with the many resources supported by Formal. 

  
---

# Formal Terraform Provider

Use the Formal Terraform Provider to interact with the
many resources supported by Formal.

Use the navigation to the left to read about the available resources.

## Example Usage

```terraform
terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 1.0.35"
    }
  }
}

# Configure the Formal Provider
provider "formal" {
  client_id  = var.client_id
  secret_key = var.secret_key
}

# Create a Role
resource "formal_role" "dior_the_data_scientist" {
  type       = "human"
  email      = "dior@acme.com"
  first_name = "dior"
  last_name  = "scientist"
}

# Create a Field encryption 
resource "formal_field_encryption" "encrypt_email_field" {
  datastore_id = formal_datastore.my_snowflake_datastore.datastore_id
  path         = "main.public.customers.email"
  key_storage  = "control_plane_only"
  key_id       = formal_key.my_email_encryption_key.id
}
```


## Authentication and Configuration

Configuration for the Formal Provider is derived from the API tokens you can generate via the [Formal Console](console.joinformal.app).

### Provider Configuration

!> **Warning:** Hard-coded credentials are not recommended in any Terraform
configuration and risks secret leakage should this file ever be committed to a
public version control system.

Credentials can be provided by adding an a `client_id` and `secret_key`.

Usage:

```terraform
provider "formal" {
  client_id = "my-client-id"
  secret_key = "my-secret-key"
}
```

### Deploying with a Managed Cloud model

Registering resources such as Keys and Datastores under the Managed Cloud model require the `cloud_account_id` parameter, which is the Formal ID of your cloud integration. You can find this information in the "Integrations" side panel in the [Formal Console](console.joinformal.app).


## Examples

See examples for each resource in the `examples/` folder.
