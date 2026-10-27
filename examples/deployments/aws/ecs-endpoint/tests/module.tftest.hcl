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
      tolist(output.container_definitions[0].linuxParameters.capabilities.add) == tolist(["NET_ADMIN", "NET_BIND_SERVICE"]) &&
      length(output.container_definitions[0].secrets) == 1 &&
      output.container_definitions[0].secrets[0].name == "FORMAL_API_KEY" &&
      output.container_definitions[0].secrets[0].valueFrom == var.api_key_secret_arn &&
      tolist(output.container_definitions[0].entryPoint) == tolist(["/usr/local/bin/formal"])
    )
    error_message = "endpoint container is misconfigured"
  }

  assert {
    condition     = !contains(keys(output.container_definitions[0]), "logConfiguration")
    error_message = "logConfiguration must be omitted when not set"
  }

  assert {
    condition = (
      tolist(output.container_definitions[0].command) == tolist(["agent", "--headless"]) &&
      alltrue([for c in slice(output.container_definitions, 1, 3) : !anytrue([for e in c.environment : contains(["HTTPS_PROXY", "JAVA_TOOL_OPTIONS"], e.name)])])
    )
    error_message = "EC2 tasks intercept transparently and get no proxy settings"
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

run "fargate_uses_the_explicit_proxy" {
  command = plan

  variables {
    launch_type = "FARGATE"
    containers = [
      {
        name  = "app"
        image = "app:1"
        environment = [
          { name = "NO_PROXY", value = ".internal.example,localhost" },
          { name = "JAVA_TOOL_OPTIONS", value = "-Xmx512m" },
        ]
      },
      {
        name  = "init"
        image = "init:1"
      },
    ]
  }

  assert {
    condition = (
      length(output.container_definitions[0].linuxParameters.capabilities.add) == 0 &&
      tolist(output.container_definitions[0].linuxParameters.capabilities.drop) == tolist(["ALL"]) &&
      tolist(output.container_definitions[0].command) == tolist(["agent", "--headless", "--explicit-proxy=127.0.0.1:3128"])
    )
    error_message = "the Fargate endpoint must add no capabilities and run in explicit capture mode"
  }

  assert {
    condition = alltrue([
      for c in slice(output.container_definitions, 1, 3) :
      contains(c.environment, { name = "HTTPS_PROXY", value = "http://127.0.0.1:3128" }) &&
      contains(c.environment, { name = "https_proxy", value = "http://127.0.0.1:3128" }) &&
      contains(c.environment, { name = "HTTP_PROXY", value = "http://127.0.0.1:3128" }) &&
      contains(c.environment, { name = "http_proxy", value = "http://127.0.0.1:3128" }) &&
      contains(c.environment, { name = "SSL_CERT_FILE", value = "/formal-ca/ca-bundle.pem" }) &&
      contains(c.environment, { name = "GRPC_DEFAULT_SSL_ROOTS_FILE_PATH", value = "/formal-ca/ca-bundle.pem" }) &&
      contains(c.environment, { name = "CURL_CA_BUNDLE", value = "/formal-ca/ca-bundle.pem" }) &&
      contains(c.dependsOn, { containerName = "formal-endpoint", condition = "HEALTHY" })
    ])
    error_message = "every Fargate app container must use the Endpoint as its HTTP and HTTPS proxy and trust its CA"
  }

  assert {
    condition = (
      contains(output.container_definitions[1].environment, { name = "NO_PROXY", value = ".internal.example,localhost,127.0.0.1,::1,169.254.169.254,169.254.170.2" }) &&
      contains(output.container_definitions[2].environment, { name = "no_proxy", value = "localhost,127.0.0.1,::1,169.254.169.254,169.254.170.2" })
    )
    error_message = "NO_PROXY must keep the app's exclusions and add loopback and the ECS metadata endpoints"
  }

  assert {
    condition = (
      contains(output.container_definitions[1].environment, { name = "JAVA_TOOL_OPTIONS", value = "-Xmx512m -Dhttps.proxyHost=127.0.0.1 -Dhttps.proxyPort=3128 -Dhttp.proxyHost=127.0.0.1 -Dhttp.proxyPort=3128 -Dhttp.nonProxyHosts=*.internal.example|localhost|127.0.0.1|::1|169.254.169.254|169.254.170.2" }) &&
      contains(output.container_definitions[2].environment, { name = "JAVA_TOOL_OPTIONS", value = "-Dhttps.proxyHost=127.0.0.1 -Dhttps.proxyPort=3128 -Dhttp.proxyHost=127.0.0.1 -Dhttp.proxyPort=3128 -Dhttp.nonProxyHosts=localhost|127.0.0.1|::1|169.254.169.254|169.254.170.2" })
    )
    error_message = "JAVA_TOOL_OPTIONS must keep the app's options and add the Endpoint as the HTTP and HTTPS proxy"
  }
}

run "fargate_rejects_an_app_https_proxy" {
  command = plan

  variables {
    launch_type = "FARGATE"
    containers = [{
      name        = "app"
      image       = "app:1"
      environment = [{ name = "https_proxy", value = "http://corporate-proxy:3128" }]
    }]
  }

  expect_failures = [output.container_definitions]
}

run "fargate_rejects_an_app_http_proxy" {
  command = plan

  variables {
    launch_type = "FARGATE"
    containers = [{
      name        = "app"
      image       = "app:1"
      environment = [{ name = "http_proxy", value = "http://corporate-proxy:3128" }]
    }]
  }

  expect_failures = [output.container_definitions]
}

run "accepts_null_attributes" {
  command = plan

  variables {
    containers = [{ name = "app", image = "app:1", environment = null, mountPoints = null, dependsOn = null }]
  }

  assert {
    condition = (
      output.container_definitions[1].mountPoints == [{ sourceVolume = "formal-ca", containerPath = "/formal-ca", readOnly = true }] &&
      output.container_definitions[1].dependsOn == [{ containerName = "formal-endpoint", condition = "HEALTHY" }]
    )
    error_message = "null attributes must be treated as empty"
  }
}

run "uses_a_supplied_ca" {
  command = plan

  variables {
    ca_secret_arn = "arn:aws:secretsmanager:us-east-1:000000000000:secret:formal-ca-AbCdEf"
  }

  assert {
    condition = (
      output.container_definitions[0].entryPoint[0] == "/bin/sh" &&
      strcontains(output.container_definitions[0].entryPoint[2], "printenv FORMAL_CA_KEY > /var/lib/formal/ca-key/tls.key") &&
      strcontains(output.container_definitions[0].entryPoint[2], "exec /usr/local/bin/formal \"$@\"") &&
      tolist(output.container_definitions[0].command) == tolist(["agent", "--headless"]) &&
      contains(output.container_definitions[0].secrets, { name = "FORMAL_CA_CERT", valueFrom = "${var.ca_secret_arn}:tls.crt::" }) &&
      contains(output.container_definitions[0].secrets, { name = "FORMAL_CA_KEY", valueFrom = "${var.ca_secret_arn}:tls.key::" })
    )
    error_message = "the endpoint must write the supplied CA to its key directory before it starts"
  }

  assert {
    condition     = alltrue([for c in slice(output.container_definitions, 1, 3) : !anytrue([for s in try(c.secrets, []) : s.name == "FORMAL_CA_KEY"])])
    error_message = "app containers must never get the CA key"
  }
}

run "rejects_a_non_secrets_manager_ca" {
  command = plan

  variables {
    ca_secret_arn = "arn:aws:ssm:us-east-1:000000000000:parameter/formal-ca"
  }

  expect_failures = [var.ca_secret_arn]
}

run "signs_in_with_oidc" {
  command = plan

  variables {
    api_key_secret_arn  = null
    oidc_integration_id = "integrationoidc_01abc"
  }

  assert {
    condition = (
      output.container_definitions[0].entryPoint[0] == "/bin/sh" &&
      strcontains(output.container_definitions[0].entryPoint[2], "printf 'auth_mode = \"oidc\"") &&
      strcontains(output.container_definitions[0].entryPoint[2], "integration_id = \"integrationoidc_01abc\"") &&
      strcontains(output.container_definitions[0].entryPoint[2], "exec /usr/local/bin/formal \"$@\"") &&
      length(output.container_definitions[0].secrets) == 0
    )
    error_message = "with OIDC the endpoint must configure the aws token source and get no API key"
  }
}

run "combines_oidc_and_a_supplied_ca" {
  command = plan

  variables {
    api_key_secret_arn  = null
    oidc_integration_id = "integrationoidc_01abc"
    ca_secret_arn       = "arn:aws:secretsmanager:us-east-1:000000000000:secret:formal-ca-AbCdEf"
  }

  assert {
    condition = (
      strcontains(output.container_definitions[0].entryPoint[2], "printenv FORMAL_CA_KEY") &&
      strcontains(output.container_definitions[0].entryPoint[2], "auth_mode") &&
      length(output.container_definitions[0].secrets) == 2
    )
    error_message = "OIDC and a supplied CA must both be set up"
  }
}

run "requires_one_credential" {
  command = plan

  variables {
    api_key_secret_arn = null
  }

  expect_failures = [output.container_definitions]
}

run "rejects_two_credentials" {
  command = plan

  variables {
    oidc_integration_id = "integrationoidc_01abc"
  }

  expect_failures = [output.container_definitions]
}

run "rejects_a_malformed_oidc_integration_id" {
  command = plan

  variables {
    api_key_secret_arn  = null
    oidc_integration_id = "integrationoidc_01abc\"; touch /tmp/x; \""
  }

  expect_failures = [var.oidc_integration_id]
}

run "rejects_unknown_launch_type" {
  command = plan

  variables {
    launch_type = "EXTERNAL"
  }

  expect_failures = [var.launch_type]
}
