output "container_definitions" {
  description = "Application containers plus the Formal Endpoint sidecar. Pass to jsonencode() for aws_ecs_task_definition.container_definitions."
  value       = concat([local.endpoint], local.apps)
}

output "volume_name" {
  description = "Task volume that shares the Formal CA. Add a volume block with this name to the task definition."
  value       = local.volume_name
}
