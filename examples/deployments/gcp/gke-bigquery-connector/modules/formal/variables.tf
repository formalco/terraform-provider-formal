variable "name" {
  description = "Name of the connector"
  type        = string
}

variable "bigquery_port" {
  description = "Port for the BigQuery listener"
  type        = number
  default     = 7777
}

variable "formal_api_key" {
  description = "Formal API key"
  type        = string
  sensitive   = true
}
