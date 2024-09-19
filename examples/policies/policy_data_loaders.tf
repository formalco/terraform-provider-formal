resource "formal_policy_data_loader" "policy_data_loader" {
  name            = "Load Zendesk tickets and related users"
  description     = "Use Zendesk API to fetch active tickets and their related users: submitters, requesters, assignees."
  key             = "zendesk_tickets"
  status          = "active"
  worker_schedule = "*/30 * * * * *"
  worker_runtime  = "python3.11"
  worker_code     = file("${path.module}/zendesk_loader.py")
}
