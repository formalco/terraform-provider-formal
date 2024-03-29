---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_sidecar Resource - terraform-provider-formal"
subcategory: ""
description: |-
  Registering a Sidecar with Formal.
---

# formal_sidecar (Resource)

Registering a Sidecar with Formal.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `deployment_type` (String) How the Sidecar should be deployed: `managed`, or `onprem`.
- `name` (String) Friendly name for this Sidecar.
- `technology` (String) Technology of the Datastore: supported values are`snowflake`, `postgres`, `redshift`, `mysql`, `mariadb`, `s3`, `dynamodb`, `mongodb`, `documentdb`, `http` and `ssh`.

### Optional

- `dataplane_id` (String) If deployment_type is managed, this is the ID of the Dataplane
- `datastore_id` (String, Deprecated) The Datastore ID that the new Sidecar will be attached to.
- `fail_open` (Boolean) Configure DNS failover from the sidecar to the original datastore. In the unlikely case where the sidecar is unhealthy, having this value of `true` will forward traffic to the original database. Default `false`.
- `formal_hostname` (String) The hostname of the created sidecar.
- `global_kms_decrypt` (Boolean) Enable all Field Encryptions created by this sidecar to be decrypted by other sidecars.
- `network_type` (String) Configure the sidecar network type. Value can be `internet-facing`, `internal` or `internet-and-internal`.
- `termination_protection` (Boolean) If set to true, this Sidecar cannot be deleted.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `version` (String) Version of the Sidecar to deploy for `managed`.

### Read-Only

- `api_key` (String, Sensitive) Api key for the deployed Sidecar.
- `created_at` (Number) Creation time of the sidecar.
- `formal_control_plane_tls_certificate` (String, Sensitive) If deployment_type is onprem, this is the Control Plane TLS Certificate to add to the deployed Sidecar.
- `id` (String) The ID of this Sidecar.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)


