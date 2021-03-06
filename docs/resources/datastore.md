---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_datastore Resource - terraform-provider-formal"
subcategory: ""
description: |-
  Registering a Datastore with Formal.
---

# formal_datastore (Resource)

Registering a Datastore with Formal.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_provider` (String) Cloud Provider that the sidecar sholud deploy in. Supported values at the moment are `aws`.
- `deployment_type` (String) How the sidecar for this datastore should be deployed: `saas`, `managed`, or `onprem`.
- `fail_open` (Boolean) Configure DNS failover from the sidecar to the original datastore. In the unlikely case where the sidecar is healthy, having this value of `true` will forward traffic to the original database. Default `false`.
- `hostname` (String) Hostname of the datastore.
- `name` (String) Friendly name for this datastore.
- `password` (String, Sensitive) Password for the original datastore that the sidecar should use. Please be sure to set this secret via Terraform environment variables.
- `technology` (String) Technology of the datastore: supported values are `snowflake`, `postgres`, and `redshift`.
- `username` (String, Sensitive) Username for the original datastore that the sidecar should use. Please be sure to set this secret via Terraform environment variables.

### Optional

- `cloud_account_id` (String) Required for managed cloud - the Formal ID for the connected Cloud Account. You can find this after creating the connection in the Formal Console.
- `cloud_region` (String) The cloud region the sidecar should be deployed in. For SaaS deployment models, supported values are `eu-west-1`, `eu-west-3`, `us-east-1`, and `us-west-2`
- `customer_vpc_id` (String) Required for managed cloud -- the VPC ID of the datastore.
- `dataplane_id` (String) If deployment_type is managed, this is the ID of the Dataplane
- `global_kms_decrypt` (Boolean) Enable all Field Encryptions created by this sidecar to be decrypted by other sidecars.
- `id` (String) The ID of this resource.
- `port` (Number) The port your datastore is listening on. Required if your `technology` is `postgres` or `redshift`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_at` (Number) Creation time of the sidecar.
- `datastore_id` (String) Formal ID for the datastore.
- `formal_control_plane_tls_certificate` (String, Sensitive) If deployment_type is onprem, this is the Control Plane TLS Certificate to add to the deployed Sidecar.
- `formal_hostname` (String) The hostname of the created sidcar.
- `net_stack_id` (String) Net Stack ID
- `org_id` (String) The Formal ID for your organisation.
- `stack_name` (String) Name of the CloudFormation stack if deployed as managed.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)


