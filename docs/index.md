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
      version = "~> 4.0.1"
    }
  }
}

# Configure the Formal Provider
provider "formal" {
  api_key  = var.formal_api_key
}

# Create a User
resource "formal_user" "dior_the_data_scientist" {
  type       = "human"
  email      = "dior@acme.com"
  first_name = "dior"
  last_name  = "scientist"
}

# Create a Resource
resource "formal_resource" "postgres_resource" {
  hostname    = "postgres-hostname"
  name        = "postgres-staging"
  technology  = "postgres"
  environment = "DEV"
  port        = 5432
}
```


## Authentication and Configuration

Configuration for the Formal Provider is derived from the API tokens you can generate via the [Formal Console](console.joinformal.app).

### Provider Configuration

!> **Warning:** Hard-coded credentials are not recommended in any Terraform
configuration and risks secret leakage should this file ever be committed to a
public version control system.

Credentials can be provided by adding an `api_key`.

Usage:

```terraform
provider "formal" {
  api_key  = var.formal_api_key
  retrieve_sensitive_values = true
}
```

Credentials can be provided by using the `FORMAL_API_KEY` environment variables.

For example:

Usage:

```terraform
provider "formal" {}
```

```bash
export FORMAL_API_KEY="some_api_key"
```

#### Retrieving Sensitive Values

You can configure the Formal Provider to disable retrieving sensitive values from the Formal API. This is useful for resources such as `formal_control_plane_tls_certificate` and `machine_role_access_token` where the sensitive values are returned by default. To enable this feature, set the `retrieve_sensitive_values` parameter to `false`.

### Deploying with a Managed Cloud model

Registering resources such as Keys and Datastores under the Managed Cloud model require the `cloud_account_id` parameter, which is the Formal ID of your cloud integration. You can find this information in the "Integrations" side panel in the [Formal Console](https://app.joinformal.com).


## Examples

See examples for each resource in the `examples/` folder of the [Formal Terraform Github Repository](https://github.com/formalco/terraform-provider-formal/tree/main/examples).
