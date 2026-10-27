output "container_definitions" {
  description = "Application containers plus the Formal Endpoint sidecar. Pass to jsonencode() for aws_ecs_task_definition.container_definitions."
  value       = concat([local.endpoint], local.apps)

  precondition {
    condition     = (var.oidc_integration_id == null) != (var.api_key_secret_arn == null)
    error_message = "Set exactly one of oidc_integration_id and api_key_secret_arn."
  }

  precondition {
    condition     = !local.explicit || !anytrue([for env in local.app_env : length(setintersection(keys(env), ["HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"])) > 0])
    error_message = "With launch_type FARGATE the Formal Endpoint is the proxy, so app containers must not set HTTPS_PROXY or HTTP_PROXY."
  }
}

output "volume_name" {
  description = "Task volume that shares the Formal CA. Add a volume block with this name to the task definition."
  value       = local.volume_name
}
