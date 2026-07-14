terraform {
  required_providers {
    formal = {
      source = "formalco/formal"
    }
    google = {
      source = "hashicorp/google"
    }
  }
}

variable "formal_api_key" {
  type        = string
  description = "The Formal API key used to authenticate the Formal provider."
  sensitive   = true
}

variable "gcp_project_id" {
  type        = string
  description = "The GCP project ID this integration grants Formal access to."
}

variable "log_bucket_name" {
  type        = string
  description = "Name of an existing GCS bucket Formal delivers logs to. Formal does not create the bucket; it must already exist in the project."
}

provider "formal" {
  api_key = var.formal_api_key
}

provider "google" {
  project = var.gcp_project_id
}

# 1. Register the GCP Cloud Integration. allow_gcs_access with an empty gcs_buckets
#    list grants log delivery to any bucket in the project; set gcs_buckets to
#    restrict it. Formal returns the AWS role ARN it federates with, plus the IAM
#    roles / buckets to grant based on the capabilities enabled here.
resource "formal_integration_cloud" "gcp" {
  name = "gcp-integration"

  gcp {
    project_id                              = var.gcp_project_id
    enable_compute_instances_autodiscovery  = true
    enable_gke_clusters_autodiscovery       = true
    enable_cloudsql_instances_autodiscovery = true
    allow_gcs_access                        = true
  }
}

# 2. Provision the GCP-side resources (service account, workload identity pool
#    provider, IAM grants) via the Formal-maintained Google module, driven by the
#    roles and buckets Formal derived from the enabled capabilities.
module "formal_gcp" {
  source = "github.com/formalco/terraform-formal-gcp"

  integration_id  = formal_integration_cloud.gcp.id
  formal_role_arn = formal_integration_cloud.gcp.aws_formal_role_arn
  project_id      = var.gcp_project_id
  roles           = formal_integration_cloud.gcp.gcp_roles
  gcs_buckets     = formal_integration_cloud.gcp.gcp_gcs_buckets
}

# 3. Report the created service account and workload identity pool provider back to
#    Formal to activate the integration. A dedicated resource avoids a dependency
#    cycle with formal_integration_cloud.gcp (which feeds the module).
resource "formal_integration_cloud_gcp_activation" "gcp" {
  integration_id                  = formal_integration_cloud.gcp.id
  service_account_email           = module.formal_gcp.service_account_email
  workload_identity_pool_provider = module.formal_gcp.workload_identity_pool_provider
}

# 4. Deliver Formal logs to a GCS bucket in the connected project. Waits for
#    activation so the service account and its IAM grants are in place first.
resource "formal_integration_log" "gcs" {
  name = "gcp-integration-logs"

  gcs {
    cloud_integration_id = formal_integration_cloud.gcp.id
    gcs_bucket_name      = var.log_bucket_name
  }

  depends_on = [formal_integration_cloud_gcp_activation.gcp]
}
