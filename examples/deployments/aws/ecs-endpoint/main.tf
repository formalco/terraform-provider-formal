locals {
  endpoint_name = "formal-endpoint"
  volume_name   = "formal-ca"
  ca_path       = "/formal-ca/ca-bundle.pem"

  ca_env = {
    for name in [
      "SSL_CERT_FILE",
      "NIX_SSL_CERT_FILE",
      "NODE_EXTRA_CA_CERTS",
      "REQUESTS_CA_BUNDLE",
      "PIP_CERT",
      "AWS_CA_BUNDLE",
      "GIT_SSL_CAINFO",
      "CARGO_HTTP_CAINFO",
    ] : name => local.ca_path
  }

  endpoint = merge(
    {
      name              = local.endpoint_name
      image             = var.image
      essential         = true
      user              = "0"
      cpu               = var.cpu
      memoryReservation = var.memory_reservation
      linuxParameters = {
        capabilities = {
          drop = ["ALL"]
          add  = ["NET_ADMIN"]
        }
      }
      secrets = [{
        name      = "FORMAL_API_KEY"
        valueFrom = var.api_key_secret_arn
      }]
      mountPoints = [{
        sourceVolume  = local.volume_name
        containerPath = "/var/lib/formal/ca"
        readOnly      = false
      }]
      # App containers wait on HEALTHY, so they never send traffic before
      # interception is armed and the CA they are told to trust exists.
      healthCheck = {
        command     = ["CMD-SHELL", "test -s /var/lib/formal/ca/ca-bundle.pem && formal transparent-proxy status | grep -q 'is active'"]
        interval    = 5
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    },
    var.log_configuration == null ? {} : { logConfiguration = var.log_configuration },
  )

  apps = [
    for c in var.containers : merge(c, {
      environment = [
        for name, value in merge(
          { for e in try(c.environment, []) : e.name => e.value },
          local.ca_env,
        ) : { name = name, value = value }
      ]
      mountPoints = concat(try(c.mountPoints, []), [{
        sourceVolume  = local.volume_name
        containerPath = "/formal-ca"
        readOnly      = true
      }])
      dependsOn = concat(try(c.dependsOn, []), [{
        containerName = local.endpoint_name
        condition     = "HEALTHY"
      }])
    })
  ]
}
