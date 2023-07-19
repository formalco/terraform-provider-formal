---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_integration_incident Resource - terraform-provider-formal"
subcategory: ""
description: |-
  Registering a Integration Incident app.
---

# formal_integration_incident (Resource)

Registering a Integration Incident app.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `api_key` (String) API Key of the Incident app.
- `name` (String) Friendly name for the Incident app.
- `type` (String) Type of the Incident app: pagerduty or custom

### Optional

- `logo` (String) Logo of the Incident app.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of the App.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)

