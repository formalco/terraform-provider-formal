provider "helm" {
  kubernetes {
    host                   = var.eks_cluster_endpoint
    cluster_ca_certificate = base64decode(var.eks_cluster_certificate_authority_data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      # This requires the awscli to be installed locally where Terraform is executed
      args = ["eks", "get-token", "--cluster-name", var.eks_cluster_name]
    }
  }
}

resource "helm_release" "example" {
  name      = "formal-snowflake-helm-chart"
  namespace = "formal-external"
  chart     = var.chart_oci
  version   = "0.0.1"
  values = [
    "${file("./helm/values.yaml")}"
  ]
}