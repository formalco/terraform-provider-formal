output "connector_id" {
  description = "ID of the created connector"
  value       = formal_connector.main.id
}

output "connector_api_key" {
  description = "API key of the connector"
  value       = formal_connector.main.api_key
}

output "resource_id" {
  description = "ID of the Cloud SQL resource"
  value       = formal_resource.cloudsql.id
}

output "resource_name" {
  description = "Name of the Cloud SQL resource"
  value       = formal_resource.cloudsql.name
}
