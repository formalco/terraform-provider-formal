terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0, < 3.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.0, < 3.0"
    }
    formal = {
      source  = "formalco/formal"
      version = "~> 4.12.8"
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

# IAM (standard): the Connector uses its ambient AWS credentials from Pod Identity.
# Do not set native_user_secret to an IAM role ARN; that would AssumeRole a second time.
resource "formal_native_user" "kubernetes_native_user" {
  resource_id        = formal_resource.kubernetes_resource.id
  native_user_id     = "aws-ambient"
  native_user_secret = "iam"
  use_as_default     = true
}

# Configure the Formal connector in Formal Control Plane
resource "formal_connector" "kubernetes_connector" {
  name = "kubernetes-connector"
}

resource "formal_connector_listener" "kubernetes_listener" {
  name         = "kubernetes-listener"
  port         = 443
  connector_id = formal_connector.kubernetes_connector.id
}

resource "formal_connector_listener_rule" "kubernetes_rule" {
  connector_listener_id = formal_connector_listener.kubernetes_listener.id
  type                  = "technology"
  rule                  = "kubernetes"
}

# EKS Pod Identity Agent is required for Pod Identity associations.
# Skip this resource (or import the existing add-on) if the agent is already installed.
resource "aws_eks_addon" "pod_identity_agent" {
  cluster_name = var.cluster_name
  addon_name   = "eks-pod-identity-agent"
}

# Pod Identity IAM role. The Connector uses this identity directly (IAM standard).
# Scope the trust policy to this cluster, namespace, and service account so another
# association in the account cannot reuse a cluster-admin role.
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "connector_assume_role" {
  statement {
    sid    = "AllowEksAuthToAssumeRoleForPodIdentity"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["pods.eks.amazonaws.com"]
    }

    actions = [
      "sts:AssumeRole",
      "sts:TagSession",
    ]

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values   = [data.aws_eks_cluster.cluster.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:RequestTag/kubernetes-namespace"
      values   = [var.namespace]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:RequestTag/kubernetes-service-account"
      values   = [kubernetes_service_account.connector.metadata[0].name]
    }
  }
}

resource "aws_iam_role" "connector" {
  name_prefix        = "formal-connector-"
  assume_role_policy = data.aws_iam_policy_document.connector_assume_role.json
}

resource "kubernetes_service_account" "connector" {
  metadata {
    name      = "formal-connector"
    namespace = var.namespace
  }
}

# Permissions the Connector needs to fetch the EKS cluster kubeconfig.
# https://docs.formal.ai/docs/guides/core-concepts/resources/kubernetes#aws-iam-permissions
resource "aws_iam_role_policy" "eks_describe" {
  name = "formal-connector-eks-describe"
  role = aws_iam_role.connector.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "eks:DescribeCluster",
          "sts:GetCallerIdentity"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_eks_pod_identity_association" "connector" {
  cluster_name    = var.cluster_name
  namespace       = var.namespace
  service_account = kubernetes_service_account.connector.metadata[0].name
  role_arn        = aws_iam_role.connector.arn

  depends_on = [aws_eks_addon.pod_identity_agent]
}

# Give Kubernetes permissions to the pod IAM role in the target Kubernetes cluster
resource "aws_eks_access_entry" "connector" {
  cluster_name  = var.cluster_name
  principal_arn = aws_iam_role.connector.arn
  type          = "STANDARD"
}

resource "aws_eks_access_policy_association" "connector" {
  cluster_name  = var.cluster_name
  principal_arn = aws_iam_role.connector.arn
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
        create = false
        name   = kubernetes_service_account.connector.metadata[0].name
      }
    })]
  )

  depends_on = [
    aws_eks_pod_identity_association.connector,
    aws_eks_access_policy_association.connector,
  ]
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
