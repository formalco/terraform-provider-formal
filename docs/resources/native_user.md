---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_native_user Resource - terraform-provider-formal"
subcategory: ""
description: |-
  This resource creates a Native User.
---

# formal_native_user (Resource)

This resource creates a Native User.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `native_user_id` (String) The username of the Native User.
- `native_user_secret` (String, Sensitive) The password of the Native User.
- `resource_id` (String) The Sidecar ID for the resource this Native User is for.

### Optional

- `termination_protection` (Boolean) If set to true, this Native User cannot be deleted.
- `use_as_default` (Boolean) The password of the Native User.

### Read-Only

- `id` (String) The ID of the Native User.

