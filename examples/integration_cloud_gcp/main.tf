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

provider "formal" {
  api_key = var.formal_api_key
}

provider "google" {
  project = var.gcp_project_id
}

# 1. Register the GCP Cloud Integration with Formal. Formal returns the AWS IAM
#    role ARN it will use to federate into your GCP workload identity pool.
resource "formal_integration_cloud" "gcp" {
  name = var.name

  gcp {
    project_id = var.gcp_project_id
  }
}

# 2. Provision the GCP-side resources (service account, workload identity pool
#    provider) using the Formal-maintained Google module. It consumes the
#    integration fields reported back.
module "formal_gcp" {
  source = "github.com/formalco/terraform-formal-gcp"

  integration_id  = formal_integration_cloud.gcp.id
  formal_role_arn = formal_integration_cloud.gcp.aws_formal_role_arn
  project_id      = var.gcp_project_id
}

# 3. Report the created service account and workload identity pool provider back
#    to Formal to activate the integration. This is a dedicated resource to avoid
#    a dependency cycle with formal_integration_cloud.gcp (which feeds the module).
resource "formal_integration_cloud_gcp_activation" "gcp" {
  integration_id                  = formal_integration_cloud.gcp.id
  service_account_email           = module.formal_gcp.service_account_email
  workload_identity_pool_provider = module.formal_gcp.workload_identity_pool_provider
}
