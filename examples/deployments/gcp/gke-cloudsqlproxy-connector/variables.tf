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

variable "connector_name" {
  description = "Name of the Formal connector"
  type        = string
  default     = "cloudsql-connector"
}

variable "cloud_sql_instance_connection" {
  description = "Cloud SQL instance connection name (format: project:region:instance)"
  type        = string
}

variable "postgres_port" {
  description = "Port for the Postgres listener"
  type        = number
  default     = 5432
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

