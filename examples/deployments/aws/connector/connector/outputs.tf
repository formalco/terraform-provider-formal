output "connector_id" {
  description = "The ID of the Formal connector"
  value       = formal_connector.main.id
}

output "nlb_dns_name" {
  description = "The DNS name of the Network Load Balancer"
  value       = aws_lb.main.dns_name
}