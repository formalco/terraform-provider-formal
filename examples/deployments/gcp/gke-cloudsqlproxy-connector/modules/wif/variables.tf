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
  description = "Kubernetes namespace where the connector will be deployed"
  type        = string
  default     = "default"
}

variable "cloud_sql_instance_connection" {
  description = "Cloud SQL instance connection name (format: project:region:instance)"
  type        = string
}
