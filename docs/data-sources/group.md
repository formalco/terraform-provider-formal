---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_group Data Source - terraform-provider-formal"
subcategory: ""
description: |-
  Data source for looking up a Group by name.
---

# formal_group (Data Source)

Data source for looking up a Group by name.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the Group.

### Read-Only

- `description` (String) Description for this Group.
- `id` (String) The Formal ID for this Group.
- `termination_protection` (Boolean) If set to true, this Group cannot be deleted.
