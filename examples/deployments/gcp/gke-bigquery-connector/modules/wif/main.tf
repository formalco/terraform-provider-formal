# Get cluster info
data "google_container_cluster" "cluster" {
  name     = var.cluster_name
  location = var.region
  project  = var.project_id
}

# Create a Google Service Account
resource "google_service_account" "connector_sa" {
  account_id   = "formal-connector-sa"
  display_name = "Service Account for Formal Connector"
  project      = var.project_id
}

# Create Workload Identity Pool
resource "random_id" "connector_pool_id" {
  byte_length = 4
}

resource "google_iam_workload_identity_pool" "connector_pool" {
  workload_identity_pool_id = "formal-pool-${random_id.connector_pool_id.hex}"
  display_name              = "Formal Connector Pool"
  description               = "Identity pool for Formal Connector"
  project                   = var.project_id
}

# Create Workload Identity Provider
resource "google_iam_workload_identity_pool_provider" "gke_provider" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.connector_pool.workload_identity_pool_id
  workload_identity_pool_provider_id = "formal-gke-provider"
  display_name                       = "GKE Provider"
  project                            = var.project_id

  attribute_mapping = {
    "google.subject"      = "assertion.sub"
    "attribute.namespace" = "assertion.namespace"
    "attribute.pod"       = "assertion.pod"
  }

  oidc {
    issuer_uri = "https://container.googleapis.com/v1/${data.google_container_cluster.cluster.id}"
  }
}

# Allow the GKE service account to impersonate the Google service account
resource "google_service_account_iam_binding" "workload_identity_binding" {
  service_account_id = google_service_account.connector_sa.name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    "serviceAccount:${var.project_id}.svc.id.goog[${var.namespace}/formal-connector]"
  ]
}

# Grant service account token creator role for user impersonation
resource "google_project_iam_member" "token_creator_role" {
  project = var.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${google_service_account.connector_sa.email}"
}
