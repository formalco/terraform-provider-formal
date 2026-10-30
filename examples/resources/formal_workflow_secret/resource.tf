resource "formal_workflow" "notify" {
  name = "notify"
  code = file("${path.module}/workflows/notify.yaml")
}

# The workflow reads it as secrets.slack_webhook_url. Increment
# value_wo_version to rotate it; the value is not stored in state.
resource "formal_workflow_secret" "slack_webhook_url" {
  workflow_id      = formal_workflow.notify.id
  name             = "slack_webhook_url"
  value_wo         = var.slack_webhook_url
  value_wo_version = 1
}

variable "slack_webhook_url" {
  type      = string
  sensitive = true
}
