---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "formal_integration_cloud Resource - terraform-provider-formal"
subcategory: ""
description: |-
  Registering a Cloud integration.
---

# formal_integration_cloud (Resource)

Registering a Cloud integration.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_region` (String) Region of the cloud provider.
- `name` (String) Name of the Integration.
- `type` (String) Type of the Integration. (Supported: aws)

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `aws_formal_iam_role` (String) The IAM role ID Formal will use to access your resources.
- `aws_formal_pingback_arn` (String) The SNS topic ARN CloudFormation can use to send events to Formal.
- `aws_formal_stack_name` (String) A generated name for your CloudFormation stack.
- `aws_template_body` (String) The template body of the CloudFormation stack.
- `id` (String) The ID of the Integration.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
