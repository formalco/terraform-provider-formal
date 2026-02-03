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
  kubernetes = {
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

  project_id                    = var.project_id
  cluster_name                  = var.cluster_name
  region                        = var.region
  namespace                     = var.namespace
  cloud_sql_instance_connection = var.cloud_sql_instance_connection
}

module "formal" {
  source = "./modules/formal"

  name                      = var.connector_name
  formal_api_key            = var.formal_api_key
  postgres_port             = var.postgres_port
  gcp_service_account_email = module.wif.service_account_email
}

# ECR credentials job for pulling Formal Connector image from ECR
resource "helm_release" "ecr_cred" {
  name       = "formal-ecr-cred"
  repository = "https://formalco.github.io/helm-charts"
  chart      = "ecr-cred"
  version    = "0.3.0"
  namespace  = var.namespace

  values = [yamlencode({
    ecrAccessKeyId     = var.ecr_access_key_id
    ecrSecretAccessKey = var.ecr_secret_access_key
  })]
}

resource "helm_release" "formal_connector" {
  name       = "formal-connector"
  repository = "https://formalco.github.io/helm-charts"
  chart      = "connector"
  version    = "0.11.0"
  namespace  = var.namespace

  values = [yamlencode({
    formalAPIKey        = module.formal.connector_api_key
    pullWithCredentials = true

    ports = [
      {
        name = "postgres"
        port = var.postgres_port
      }
    ]

    serviceAccount = {
      create = true
      name   = "formal-connector"
      annotations = {
        "iam.gke.io/gcp-service-account" = module.wif.service_account_email
      }
    }

    service = {
      type = "LoadBalancer"
      annotations = {
        "cloud.google.com/load-balancer-type" = "Internal"
      }
    }

    # Cloud SQL Proxy sidecar for secure connectivity to Cloud SQL
    # Uses port 5433 internally to avoid conflict with connector's port 5432
    # Note: Do not use --auto-iam-authn - the Formal Connector handles IAM auth via iam_gcp
    sidecars = [
      {
        name  = "cloud-sql-proxy"
        image = "gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.14.2"
        args = [
          var.cloud_sql_instance_connection,
          "--port=5433"
        ]
        securityContext = {
          runAsNonRoot = true
        }
        resources = {
          requests = {
            cpu    = "100m"
            memory = "128Mi"
          }
          limits = {
            cpu    = "500m"
            memory = "256Mi"
          }
        }
      }
    ]
  })]

  depends_on = [module.wif, helm_release.ecr_cred]
}

data "kubernetes_service" "formal_connector" {
  metadata {
    name      = "formal-connector"
    namespace = var.namespace
  }

  depends_on = [helm_release.formal_connector]
}
