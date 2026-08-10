output "connector_id" {
  description = "ID of the registered Connector."
  value       = formal_connector.main.id
}

output "connector_api_key" {
  description = "API key for the Connector container."
  value       = formal_connector.main.api_key
  sensitive   = true
}

output "connector_nlb_dns_name" {
  description = "DNS name of the Connector's network load balancer."
  value       = aws_lb.connector.dns_name
}

output "connector_service_name" {
  description = "Name of the Connector ECS service."
  value       = aws_ecs_service.connector.name
}

output "connector_log_group_name" {
  description = "CloudWatch log group for the Connector container."
  value       = aws_cloudwatch_log_group.connector.name
}

output "satellite_id" {
  description = "ID of the registered policy-data-loader Satellite."
  value       = formal_satellite.policy_data_loader.id
}

output "satellite_api_key" {
  description = "API key for the Satellite container."
  value       = formal_satellite.policy_data_loader.api_key
  sensitive   = true
}

output "satellite_service_name" {
  description = "Name of the Satellite ECS service."
  value       = aws_ecs_service.satellite.name
}

output "satellite_hostname" {
  description = "In-VPC DNS name the Connector uses to reach the Satellite on :50056."
  value       = formal_satellite_hostname.satellite.hostname
}

output "satellite_log_group_name" {
  description = "CloudWatch log group for the Satellite container."
  value       = aws_cloudwatch_log_group.satellite.name
}

output "policy_data_loader_id" {
  description = "ID of the example policy data loader."
  value       = formal_policy_data_loader.example.id
}
