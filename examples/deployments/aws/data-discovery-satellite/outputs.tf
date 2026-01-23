output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.main.name
}

output "ecs_service_id" {
  description = "ID of the ECS service"
  value       = aws_ecs_service.main.id
}

output "task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = aws_ecs_task_definition.main.arn
}

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.main.name
}

output "task_execution_role_arn" {
  description = "ARN of the ECS task execution role"
  value       = local.execution_role_arn
}

output "task_role_arn" {
  description = "ARN of the ECS task role"
  value       = local.task_role_arn
}
