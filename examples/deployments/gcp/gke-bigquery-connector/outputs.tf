output "kubernetes_service_ip" {
  description = "Load balancer IP address of the connector (internal by default)"
  value       = data.kubernetes_service.formal_connector.status[0].load_balancer[0].ingress[0].ip
}

output "kubernetes_service_internal_hostname" {
  description = "Internal hostname of the connector"
  value       = "${data.kubernetes_service.formal_connector.metadata[0].name}.${var.namespace}.svc.cluster.local"
}

output "formal_connector_id" {
  description = "ID of the connector (needed for the Helm chart)"
  value       = module.formal.connector_id
}

output "google_service_account_email" {
  description = "Email of the service account created for the connector (needed for the Helm chart)"
  value       = module.wif.service_account_email
}

output "google_service_account_id" {
  description = "Client ID of the service account (needed for domain-wide delegation setup)"
  value       = module.wif.service_account_id
}

output "next_steps" {
  description = "Next steps to complete the setup"
  value       = "Configure domain-wide delegation in Google Workspace Admin Console using the google_service_account_id output. Instructions are available in the README."
}
