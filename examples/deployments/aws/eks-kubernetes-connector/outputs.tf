output "kubernetes_service_external_ip" {
  description = "External IP address of the connector"
  value       = data.kubernetes_service.formal_connector.status[0].load_balancer[0].ingress[0].ip
}

output "kubernetes_service_internal_hostname" {
  description = "Internal hostname of the connector"
  value       = "${data.kubernetes_service.formal_connector.metadata[0].name}.${var.namespace}.svc.cluster.local"
}

output "formal_connector_id" {
  description = "ID of the connector"
  value       = formal_connector.kubernetes_connector.id
}

output "formal_connector_hostname" {
  description = "Formal hostname for the connector"
  value       = formal_connector_hostname.kubernetes_connector_hostname.hostname
}
