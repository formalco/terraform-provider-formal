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

locals {
  helm_values_file = fileexists(var.helm_values) ? var.helm_values : "helm/values.yaml"
}

resource "helm_release" "formal_connector" {
  name      = "formal-connector"
  chart     = "./helm"
  namespace = var.namespace

  values = [
    file(local.helm_values_file),
    yamlencode({
      googleServiceAccount = module.wif.service_account_email
      formalAPIKey         = module.formal.connector_api_key
      connectorId          = module.formal.connector_id
      secrets = {
        ecrAccessKeyId     = var.ecr_access_key_id
        ecrSecretAccessKey = var.ecr_secret_access_key
      }
      ports = {
        bigquery    = 7777
        healthCheck = 8080
      }
    })
  ]

  depends_on = [module.wif]
}

data "kubernetes_service" "formal_connector" {
  metadata {
    name      = "formal-connector"
    namespace = var.namespace
  }

  depends_on = [helm_release.formal_connector]
}
