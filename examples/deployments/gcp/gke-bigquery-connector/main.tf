terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

data "google_client_config" "default" {}

provider "kubernetes" {
  host                   = "https://${data.google_container_cluster.cluster.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate)
}

provider "helm" {
  kubernetes {
    host                   = "https://${data.google_container_cluster.cluster.endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate)
  }
}

data "google_container_cluster" "cluster" {
  name     = var.cluster_name
  location = var.region
  project  = var.project_id
}

module "wif" {
  source = "./modules/wif"

  project_id   = var.project_id
  cluster_name = var.cluster_name
  region       = var.region
  namespace    = var.namespace
}

module "formal" {
  source = "./modules/formal"

  name           = "bigquery-connector"
  formal_api_key = var.formal_api_key
  bigquery_port  = 7777
}

resource "helm_release" "formal_connector" {
  name       = "formal-connector"
  repository = "https://formalco.github.io/helm-charts"
  chart      = "connector"
  version    = "0.14.0"
  namespace  = var.namespace

  values = [yamlencode({
    formalAPIKey = module.formal.connector_api_key

    # Use Formal's public GCP Artifact Registry image on GKE (no ECR credentials needed)
    image = {
      repository = "us-docker.pkg.dev/formal-public-assets/formalco-prod-connector/formalco-prod-connector"
    }

    ports = [
      {
        name = "bigquery"
        port = 7777
      }
    ]

    serviceAccount = {
      create = true
      name   = "formal-connector"
      annotations = {
        # Workload Identity: allow the Connector pod to impersonate the GCP SA (BigQuery access)
        "iam.gke.io/gcp-service-account" = module.wif.service_account_email
      }
    }

    service = {
      type = "LoadBalancer"
      annotations = {
        # Internal Load Balancer (VPC-only). Remove this annotation for an external LB.
        "cloud.google.com/load-balancer-type" = "Internal"
        # By default, Google Cloud Load Balancer's forwarding rule has
        # global access disabled: client VMs, Cloud VPN tunnels, or Cloud
        # Interconnect attachments (VLANs) must be located in the same
        # region as the internal passthrough Network Load Balancer. To
        # support clients in all regions, uncomment the following line:
        # "networking.gke.io/internal-load-balancer-allow-global-access" = "true"
      }
    }
  })]

  depends_on = [module.wif]
}

data "kubernetes_service" "formal_connector" {
  metadata {
    name      = "formal-connector"
    namespace = var.namespace
  }

  depends_on = [helm_release.formal_connector]
}
