---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_dataplane Resource - terraform-provider-formal"
subcategory: ""
description: |-
  Creating a Dataplane with Formal.
---

# formal_dataplane (Resource)

Creating a Dataplane with Formal.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `availability_zones` (Number) Number of availability zones.
- `cloud_account_id` (String) Cloud account ID for deploying the dataplane.
- `cloud_region` (String) The cloud region the dataplane should be deployed in.
- `customer_vpc_id` (String) The VPC ID that this dataplane should be deployed in.
- `name` (String) Friendly name for this dataplane.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) ID of this dataplane.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)


