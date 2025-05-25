variable "region" {
  description = "AWS region"
  type        = string
}

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "namespace" {
  description = "Kubernetes namespace to deploy the connector"
  type        = string
  default     = "formal"
}

variable "formal_api_key" {
  description = "Formal API key"
  type        = string
  sensitive   = true
}

variable "formal_org_name" {
  description = "Name of your Formal organization"
  type        = string
}

variable "helm_values" {
  description = "Path to additional Helm values file"
  type        = string
  default     = "values.yaml"
}
