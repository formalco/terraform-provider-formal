variable "resource_group_name" {
  description = "Existing Azure resource group where the AKS cluster and managed identity are created"
  type        = string
}

variable "cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
}

variable "node_vm_size" {
  description = "VM size for the single system node"
  type        = string
  default     = "Standard_D2s_v5"
}

variable "namespace" {
  description = "Kubernetes namespace where the credential helper and Connector are deployed"
  type        = string
  default     = "formal"
}

variable "connector_name" {
  description = "Name of the Connector in Formal"
  type        = string
  default     = "aks-kubernetes-connector"
}

variable "formal_api_key" {
  description = "Formal API key"
  type        = string
  sensitive   = true
}

variable "gar_pull_identity_name" {
  description = "Name of an existing user-assigned managed identity that Formal has authorized to pull GAR images"
  type        = string
}

variable "gar_provider_id" {
  description = "Formal-provided GCP workload identity provider ID for the GAR pull identity"
  type        = string
}

variable "gar_service_account_email" {
  description = "Formal-provided GCP service account email for GAR image pulls"
  type        = string
}
