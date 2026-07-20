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
