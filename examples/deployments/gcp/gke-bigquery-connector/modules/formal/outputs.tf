output "connector_id" {
  description = "ID of the created connector"
  value       = formal_connector.main.id
}

output "connector_api_key" {
  description = "API key of the connector"
  value       = formal_connector.main.api_key
}

output "bigquery_resource_id" {
  description = "ID of the BigQuery resource"
  value       = formal_resource.bigquery.id
}
