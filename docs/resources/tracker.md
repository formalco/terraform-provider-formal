---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_tracker Resource - terraform-provider-formal"
subcategory: ""
description: |-
  Creating a Tracker in Formal.
---

# formal_tracker (Resource)

Creating a Tracker in Formal.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `path` (String) Path associated with this tracker.
- `resource_id` (String) Tracker linked to the following resource id.

### Optional

- `allow_clear_text_value` (Boolean) If set to true, this Tracker allow clear text value.
- `termination_protection` (Boolean) If set to true, this Tracker cannot be deleted.

### Read-Only

- `created_at` (String) When the policy was created.
- `id` (String) ID of this Tracker.
- `updated_at` (String) Last update time.
