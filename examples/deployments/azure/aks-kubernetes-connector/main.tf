terraform {
  required_version = ">= 1.11"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.81"
    }
    formal = {
      source  = "formalco/formal"
      version = "~> 4.24.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0, < 3.0"
    }
  }
}

provider "azurerm" {
  features {}

  resource_provider_registrations = "none"
}

provider "formal" {
  api_key = var.formal_api_key
}

data "azurerm_client_config" "current" {}

data "azurerm_resource_group" "deployment" {
  name = var.resource_group_name
}

data "azurerm_user_assigned_identity" "gar_pull" {
  name                = var.gar_pull_identity_name
  resource_group_name = var.resource_group_name
}

resource "azurerm_kubernetes_cluster" "this" {
  name                = var.cluster_name
  location            = data.azurerm_resource_group.deployment.location
  resource_group_name = data.azurerm_resource_group.deployment.name
  dns_prefix          = var.cluster_name
  sku_tier            = "Free"

  oidc_issuer_enabled       = true
  workload_identity_enabled = true

  default_node_pool {
    name       = "system"
    node_count = 1
    vm_size    = var.node_vm_size

    upgrade_settings {
      drain_timeout_in_minutes      = 0
      max_surge                     = "10%"
      node_soak_duration_in_minutes = 0
    }
  }

  azure_active_directory_role_based_access_control {
    azure_rbac_enabled = false
    tenant_id          = data.azurerm_client_config.current.tenant_id
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin    = "azure"
    load_balancer_sku = "standard"
  }
}

provider "kubernetes" {
  host                   = azurerm_kubernetes_cluster.this.kube_admin_config[0].host
  client_certificate     = base64decode(azurerm_kubernetes_cluster.this.kube_admin_config[0].client_certificate)
  client_key             = base64decode(azurerm_kubernetes_cluster.this.kube_admin_config[0].client_key)
  cluster_ca_certificate = base64decode(azurerm_kubernetes_cluster.this.kube_admin_config[0].cluster_ca_certificate)
}

provider "helm" {
  kubernetes = {
    host                   = azurerm_kubernetes_cluster.this.kube_admin_config[0].host
    client_certificate     = base64decode(azurerm_kubernetes_cluster.this.kube_admin_config[0].client_certificate)
    client_key             = base64decode(azurerm_kubernetes_cluster.this.kube_admin_config[0].client_key)
    cluster_ca_certificate = base64decode(azurerm_kubernetes_cluster.this.kube_admin_config[0].cluster_ca_certificate)
  }
}

resource "azurerm_user_assigned_identity" "connector" {
  name                = "${var.cluster_name}-formal-connector"
  location            = data.azurerm_resource_group.deployment.location
  resource_group_name = data.azurerm_resource_group.deployment.name
}

resource "azurerm_federated_identity_credential" "connector" {
  name                      = "formal-connector"
  user_assigned_identity_id = azurerm_user_assigned_identity.connector.id
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = azurerm_kubernetes_cluster.this.oidc_issuer_url
  subject                   = "system:serviceaccount:${var.namespace}:formal-connector"
}

resource "kubernetes_cluster_role_binding" "connector_cluster_admin" {
  metadata {
    name = "${var.connector_name}-cluster-admin"
  }

  subject {
    kind      = "User"
    name      = azurerm_user_assigned_identity.connector.principal_id
    api_group = "rbac.authorization.k8s.io"
  }

  role_ref {
    kind      = "ClusterRole"
    name      = "cluster-admin"
    api_group = "rbac.authorization.k8s.io"
  }
}

resource "azurerm_federated_identity_credential" "gar_pull" {
  name                      = "formal-gar-${var.cluster_name}"
  user_assigned_identity_id = data.azurerm_user_assigned_identity.gar_pull.id
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = azurerm_kubernetes_cluster.this.oidc_issuer_url
  subject                   = "system:serviceaccount:${var.namespace}:formal-gar-secret-manager"
}

resource "formal_resource" "kubernetes" {
  technology = "kubernetes"
  name       = azurerm_kubernetes_cluster.this.name
  hostname   = azurerm_kubernetes_cluster.this.fqdn
  port       = 443
}

resource "formal_resource_tls_configuration" "kubernetes" {
  resource_id       = formal_resource.kubernetes.id
  tls_config        = "verify-full"
  tls_min_version   = "TLSv1.2"
  tls_ca_truststore = base64decode(azurerm_kubernetes_cluster.this.kube_config[0].cluster_ca_certificate)
}

resource "formal_native_user" "kubernetes" {
  resource_id        = formal_resource.kubernetes.id
  native_user_id     = azurerm_user_assigned_identity.connector.client_id
  native_user_secret = "iam_azure"
  use_as_default     = true
}

resource "formal_connector" "kubernetes" {
  name = var.connector_name
}

resource "formal_connector_listener" "kubernetes" {
  name         = "${var.connector_name}-kubernetes"
  port         = 443
  connector_id = formal_connector.kubernetes.id
}

resource "formal_connector_listener_rule" "kubernetes" {
  connector_listener_id = formal_connector_listener.kubernetes.id
  type                  = "technology"
  rule                  = "kubernetes"
}

resource "helm_release" "azure_gar_cred" {
  name       = "formal-azure-gar-cred"
  repository = "https://formalco.github.io/helm-charts"
  chart      = "azure-gar-cred"
  version    = "0.1.0"
  namespace  = var.namespace

  create_namespace = true
  wait_for_jobs    = true

  values = [yamlencode({
    azure = {
      clientId = data.azurerm_user_assigned_identity.gar_pull.client_id
      tenantId = data.azurerm_client_config.current.tenant_id
    }
    gcp = {
      providerId          = var.gar_provider_id
      serviceAccountEmail = var.gar_service_account_email
    }
  })]

  depends_on = [azurerm_federated_identity_credential.gar_pull]
}

resource "helm_release" "formal_connector" {
  name       = "formal-connector"
  repository = "https://formalco.github.io/helm-charts"
  chart      = "connector"
  version    = "0.16.0"
  namespace  = var.namespace

  values = [yamlencode({
    formalAPIKey        = formal_connector.kubernetes.api_key
    pullWithCredentials = true
    replicaCount        = 1

    image = {
      repository = "us-docker.pkg.dev/formal-public-assets/formalco-prod-connector/formalco-prod-connector"
    }

    ports = [
      {
        name = "kubernetes"
        port = 443
      }
    ]

    resources = {
      requests = {
        cpu    = "100m"
        memory = "256Mi"
      }
      limits = {
        cpu    = "500m"
        memory = "1Gi"
      }
    }

    podLabels = {
      "azure.workload.identity/use" = "true"
    }

    serviceAccount = {
      create = true
      name   = "formal-connector"
      annotations = {
        "azure.workload.identity/client-id" = azurerm_user_assigned_identity.connector.client_id
        "azure.workload.identity/tenant-id" = data.azurerm_client_config.current.tenant_id
      }
    }

    service = {
      type = "LoadBalancer"
      annotations = {
        "service.beta.kubernetes.io/azure-load-balancer-internal" = "true"
      }
    }
  })]

  depends_on = [
    azurerm_federated_identity_credential.connector,
    kubernetes_cluster_role_binding.connector_cluster_admin,
    helm_release.azure_gar_cred,
  ]
}
