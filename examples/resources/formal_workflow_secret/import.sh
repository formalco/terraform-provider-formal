# Workflow secrets are imported by workflow ID and secret ID. The value is never
# returned, so set value or value_wo after importing.
terraform import formal_workflow_secret.slack_webhook_url '<workflow_id>/<secret_id>'
