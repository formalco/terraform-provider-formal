output "aks_cluster_name" {
  description = "Name of the created AKS cluster"
  value       = azurerm_kubernetes_cluster.this.name
}

output "aks_oidc_issuer_url" {
  description = "OIDC issuer enabled for Azure Workload Identity"
  value       = azurerm_kubernetes_cluster.this.oidc_issuer_url
}

output "connector_managed_identity_client_id" {
  description = "Client ID used by the Connector's Azure managed identity native user"
  value       = azurerm_user_assigned_identity.connector.client_id
}

output "formal_connector_id" {
  description = "ID of the Connector created in Formal"
  value       = formal_connector.kubernetes.id
}

output "formal_kubernetes_resource_id" {
  description = "ID of the AKS resource created in Formal"
  value       = formal_resource.kubernetes.id
}
