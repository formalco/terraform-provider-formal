terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
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
    formal = {
      source  = "formalco/formal"
      version = ">= 4.6.2"
    }
    time = {
      source  = "hashicorp/time"
      version = ">= 0.9.0"
    }
  }
}

provider "aws" {
  region = var.region
}

provider "formal" {
  api_key = var.formal_api_key
}

data "aws_eks_cluster" "cluster" {
  name = var.cluster_name
}

data "aws_eks_cluster_auth" "cluster" {
  name = var.cluster_name
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.cluster.token
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.cluster.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.cluster.token
  }
}

# Configure the target Kubernetes resource in Formal Control Plane (the deployment Kubernetes cluster itself)
resource "formal_resource" "kubernetes_resource" {
  technology = "kubernetes"
  name       = data.aws_eks_cluster.cluster.name
  hostname   = data.aws_eks_cluster.cluster.endpoint
  port       = 443
}

resource "formal_native_user" "kubernetes_native_user" {
  resource_id        = formal_resource.kubernetes_resource.id
  native_user_id     = "iam"
  native_user_secret = "iam"
  use_as_default     = true
}

# Configure the Formal connector in Formal Control Plane
resource "formal_connector" "kubernetes_connector" {
  name = "kubernetes-connector"
}

resource "formal_connector_listener" "kubernetes_listener" {
  name = "kubernetes-listener"
  port = 443
}

resource "formal_connector_listener_rule" "kubernetes_rule" {
  connector_listener_id = formal_connector_listener.kubernetes_listener.id
  type                  = "technology"
  rule                  = "kubernetes"
}

resource "formal_connector_listener_link" "kubernetes_link" {
  connector_id          = formal_connector.kubernetes_connector.id
  connector_listener_id = formal_connector_listener.kubernetes_listener.id
}

# Create the IAM role for the connector and bind it to the created service account
data "aws_caller_identity" "current" {}

module "connector_irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.14"

  role_name_prefix = "formal-connector-"

  oidc_providers = {
    main = {
      provider_arn               = replace(data.aws_eks_cluster.cluster.identity[0].oidc[0].issuer, "https://", "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/")
      namespace_service_accounts = ["${var.namespace}:formal-connector"]
    }
  }
}

resource "kubernetes_service_account" "connector" {
  metadata {
    name      = "formal-connector"
    namespace = var.namespace
    annotations = {
      "eks.amazonaws.com/role-arn" = module.connector_irsa.iam_role_arn
    }
  }
}

# Allow the Connector to pull the target resource kubeconfig from AWS EKS API
resource "aws_iam_policy" "eks_describe_policy" {
  name        = "formal-connector-eks-describe-policy"
  description = "Policy for Formal connector to describe EKS cluster and get caller identity"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "eks:DescribeCluster",
          "sts:GetCallerIdentity"
        ],
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_describe_policy_attachment" {
  role       = module.connector_irsa.iam_role_name
  policy_arn = aws_iam_policy.eks_describe_policy.arn
}

# Give Kubernetes permissions to the pod IAM role in the target Kubernetes cluster
resource "aws_eks_access_entry" "connector" {
  cluster_name  = var.cluster_name
  principal_arn = module.connector_irsa.iam_role_arn
  type          = "STANDARD"
}

resource "aws_eks_access_policy_association" "connector" {
  cluster_name  = var.cluster_name
  principal_arn = module.connector_irsa.iam_role_arn
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
  access_scope {
    type = "cluster"
  }
}

# Deploy the Connector in the EKS cluster
resource "helm_release" "formal_connector" {
  name       = "formal-connector"
  repository = "https://formalco.github.io/helm-charts"
  chart      = "connector"
  namespace  = var.namespace

  values = concat(
    fileexists(var.helm_values) ? [file(var.helm_values)] : [],
    [yamlencode({
      formalAPIKey = formal_connector.kubernetes_connector.api_key
      connectorId  = formal_connector.kubernetes_connector.id
      ports = {
        kubernetes  = 443
        healthCheck = 8080
      }
      serviceAccount = {
        name = kubernetes_service_account.connector.metadata[0].name
      }
    })]
  )
}

# Set the Connector hostname in Formal Control Plane according to the DNS record of the EKS service
data "kubernetes_service" "formal_connector" {
  metadata {
    name      = "formal-connector"
    namespace = var.namespace
  }

  depends_on = [helm_release.formal_connector]
}

resource "formal_connector_hostname" "kubernetes_connector_hostname" {
  connector_id = formal_connector.kubernetes_connector.id
  hostname     = "kubernetes-connector.${var.formal_org_name}.connectors.joinformal.com"
  dns_record   = data.kubernetes_service.formal_connector.status[0].load_balancer[0].ingress[0].hostname

  depends_on = [data.kubernetes_service.formal_connector]
}
