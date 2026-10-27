variables {
  api_key_secret_arn = "arn:aws:secretsmanager:us-east-1:000000000000:secret:formal"
  containers = [
    {
      name  = "app"
      image = "app:1"
      environment = [
        { name = "MODE", value = "prod" },
        { name = "SSL_CERT_FILE", value = "/etc/ssl/app.pem" },
      ]
      mountPoints = [{ sourceVolume = "data", containerPath = "/data" }]
      dependsOn   = [{ containerName = "init", condition = "SUCCESS" }]
    },
    {
      name  = "init"
      image = "init:1"
    },
  ]
}

run "adds_endpoint_sidecar" {
  command = plan

  assert {
    condition     = length(output.container_definitions) == 3
    error_message = "expected the endpoint plus both app containers"
  }

  assert {
    condition = (
      output.container_definitions[0].name == "formal-endpoint" &&
      output.container_definitions[0].essential &&
      output.container_definitions[0].user == "0" &&
      output.container_definitions[0].linuxParameters.capabilities.add == ["NET_ADMIN"] &&
      output.container_definitions[0].secrets[0].valueFrom == var.api_key_secret_arn
    )
    error_message = "endpoint container is misconfigured"
  }

  assert {
    condition     = !contains(keys(output.container_definitions[0]), "logConfiguration")
    error_message = "logConfiguration must be omitted when not set"
  }
}

run "wires_app_containers" {
  command = plan

  assert {
    condition = alltrue([
      for c in slice(output.container_definitions, 1, 3) :
      contains(c.dependsOn, { containerName = "formal-endpoint", condition = "HEALTHY" }) &&
      contains(c.mountPoints, { sourceVolume = "formal-ca", containerPath = "/formal-ca", readOnly = true })
    ])
    error_message = "every app container must mount the CA and wait for the endpoint"
  }

  assert {
    condition = (
      contains(output.container_definitions[1].environment, { name = "MODE", value = "prod" }) &&
      contains(output.container_definitions[1].environment, { name = "SSL_CERT_FILE", value = "/formal-ca/ca-bundle.pem" }) &&
      length([for e in output.container_definitions[1].environment : e if e.name == "SSL_CERT_FILE"]) == 1
    )
    error_message = "app env must be kept, with the Formal CA overriding trust variables"
  }

  assert {
    condition = (
      contains(output.container_definitions[1].dependsOn, { containerName = "init", condition = "SUCCESS" }) &&
      contains(output.container_definitions[1].mountPoints, { sourceVolume = "data", containerPath = "/data" })
    )
    error_message = "existing dependsOn and mountPoints must be kept"
  }
}

run "sets_log_configuration" {
  command = plan

  variables {
    log_configuration = { logDriver = "awslogs", options = { "awslogs-group" = "/ecs/formal" } }
  }

  assert {
    condition     = output.container_definitions[0].logConfiguration.logDriver == "awslogs"
    error_message = "logConfiguration must be passed through"
  }
}
