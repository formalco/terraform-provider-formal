---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_sidecar_datastore_link Resource - terraform-provider-formal"
subcategory: ""
description: |-
  Linking a Datastore to a Sidecar in Formal.
---

# formal_sidecar_datastore_link (Resource)

Linking a Datastore to a Sidecar in Formal.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `datastore_id` (String) Datastore ID to be linked.
- `port` (Number) Port.
- `sidecar_id` (String) Sidecar ID that should be linked.

### Optional

- `termination_protection` (Boolean) If set to true, this Sidecar Datastore Link cannot be deleted.

### Read-Only

- `id` (String) Resource ID


