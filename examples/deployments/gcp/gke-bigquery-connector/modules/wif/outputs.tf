output "service_account_email" {
  description = "Email of the created Google service account"
  value       = google_service_account.connector_sa.email
}

output "service_account_id" {
  description = "Client ID of the created Google service account (needed for domain-wide delegation setup)"
  value       = google_service_account.connector_sa.unique_id
}

output "workload_identity_pool_id" {
  description = "ID of the Workload Identity Pool"
  value       = google_iam_workload_identity_pool.connector_pool.id
}

output "workload_identity_provider_id" {
  description = "ID of the Workload Identity Provider"
  value       = google_iam_workload_identity_pool_provider.gke_provider.id
}
