# Get cluster info
data "google_container_cluster" "cluster" {
  name     = var.cluster_name
  location = var.region
  project  = var.project_id
}

# Create a Google Service Account for the connector
resource "google_service_account" "connector_sa" {
  account_id   = "formal-cloudsql-connector"
  display_name = "Service Account for Formal Connector with Cloud SQL"
  project      = var.project_id
}

# Grant Cloud SQL Client role to the service account
resource "google_project_iam_member" "cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.connector_sa.email}"
}

# Grant Cloud SQL Instance User role for IAM authentication
resource "google_project_iam_member" "cloudsql_instance_user" {
  project = var.project_id
  role    = "roles/cloudsql.instanceUser"
  member  = "serviceAccount:${google_service_account.connector_sa.email}"
}

# Allow the GKE service account to impersonate the Google service account via Workload Identity
resource "google_service_account_iam_binding" "workload_identity_binding" {
  service_account_id = google_service_account.connector_sa.name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    "serviceAccount:${var.project_id}.svc.id.goog[${var.namespace}/formal-connector]"
  ]
}

# Create IAM database user in Cloud SQL for the service account
# The name must be the service account email without .gserviceaccount.com suffix
resource "google_sql_user" "iam_user" {
  name     = trimsuffix(google_service_account.connector_sa.email, ".gserviceaccount.com")
  instance = element(split(":", var.cloud_sql_instance_connection), 2)
  project  = var.project_id
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}
