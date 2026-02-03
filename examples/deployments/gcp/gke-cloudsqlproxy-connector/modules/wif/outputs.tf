output "service_account_email" {
  description = "Email of the created Google service account"
  value       = google_service_account.connector_sa.email
}

output "service_account_id" {
  description = "Unique ID of the created Google service account"
  value       = google_service_account.connector_sa.unique_id
}
