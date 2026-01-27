output "kubernetes_service_ip" {
  description = "Load balancer IP address of the connector (internal by default)"
  value       = data.kubernetes_service.formal_connector.status[0].load_balancer[0].ingress[0].ip
}

output "kubernetes_service_internal_hostname" {
  description = "Internal hostname of the connector"
  value       = "${data.kubernetes_service.formal_connector.metadata[0].name}.${var.namespace}.svc.cluster.local"
}

output "formal_connector_id" {
  description = "ID of the connector"
  value       = module.formal.connector_id
}

output "formal_resource_id" {
  description = "ID of the Cloud SQL resource in Formal"
  value       = module.formal.resource_id
}

output "google_service_account_email" {
  description = "Email of the service account created for the connector"
  value       = module.wif.service_account_email
}

output "connection_string" {
  description = "Example psql connection string (for use within the cluster)"
  value       = "psql \"host=${data.kubernetes_service.formal_connector.metadata[0].name}.${var.namespace}.svc.cluster.local port=${var.postgres_port} user=<formal-user> dbname=<database>@${module.formal.resource_name}\""
}
