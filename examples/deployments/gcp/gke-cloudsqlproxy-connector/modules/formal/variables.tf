variable "name" {
  description = "Name of the connector"
  type        = string
}

variable "resource_name" {
  description = "Name of the Cloud SQL resource in Formal"
  type        = string
  default     = null
}

variable "postgres_port" {
  description = "Port for the Postgres listener"
  type        = number
  default     = 5432
}

variable "formal_api_key" {
  description = "Formal API key"
  type        = string
  sensitive   = true
}

variable "gcp_service_account_email" {
  description = "GCP service account email for Cloud SQL IAM authentication"
  type        = string
}
