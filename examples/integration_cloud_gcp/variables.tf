variable "formal_api_key" {
  type        = string
  description = "The Formal API key used to authenticate the Formal provider."
  sensitive   = true
}

variable "name" {
  type        = string
  description = "Name of the Formal Cloud Integration."
  default     = "gcp-integration"
}

variable "gcp_project_id" {
  type        = string
  description = "The GCP project ID this integration grants Formal access to."
}
