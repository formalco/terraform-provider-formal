variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "region" {
  description = "The GCP region"
  type        = string
}

variable "cluster_name" {
  description = "Name of the GKE cluster"
  type        = string
}

variable "namespace" {
  description = "Kubernetes namespace to deploy the connector"
  type        = string
  default     = "default"
}

variable "formal_api_key" {
  description = "Formal API key for the connector"
  type        = string
  sensitive   = true
}

variable "ecr_access_key_id" {
  description = "ECR access key ID"
  type        = string
}

variable "ecr_secret_access_key" {
  description = "ECR secret access key"
  type        = string
  sensitive   = true
}

variable "helm_values" {
  description = "Path to the Helm values file. If not found, will use the default values.yaml from the chart"
  type        = string
  default     = "helm/values.yaml"
}
