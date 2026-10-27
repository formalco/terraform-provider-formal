locals {
  endpoint_name = "formal-endpoint"
  volume_name   = "formal-ca"
  ca_path       = "/formal-ca/ca-bundle.pem"

  # Fargate grants no CAP_NET_ADMIN, so the Endpoint cannot redirect the task's
  # traffic. It serves clients that use HTTPS_PROXY and HTTP_PROXY instead.
  explicit   = var.launch_type == "FARGATE"
  proxy_host = "127.0.0.1"
  proxy_port = 3128
  proxy_addr = "${local.proxy_host}:${local.proxy_port}"
  proxy_url  = "http://${local.proxy_addr}"
  # The task metadata and credential endpoints must stay direct.
  no_proxy = ["localhost", "127.0.0.1", "::1", "169.254.169.254", "169.254.170.2"]
  # What `formal transparent-proxy status` prints once the Endpoint is intercepting.
  ready_status = local.explicit ? "Explicit proxy is listening" : "is active"

  # ECS has no secret volumes, so a supplied CA arrives as variables and a
  # shell writes it where the Endpoint loads its CA from.
  ca_key_dir = "/var/lib/formal/ca-key"
  ca_setup = var.ca_secret_arn == null ? [] : [
    "mkdir -p ${local.ca_key_dir}",
    "printenv FORMAL_CA_CERT > ${local.ca_key_dir}/tls.crt",
    "printenv FORMAL_CA_KEY > ${local.ca_key_dir}/tls.key",
    "unset FORMAL_CA_CERT FORMAL_CA_KEY",
  ]
  # OIDC settings only come from config.toml. auth_mode is a top-level key, so
  # it must precede the image's own tables.
  config_path = "/root/.formal/config.toml"
  oidc_setup = var.oidc_integration_id == null ? [] : [
    "mkdir -p /root/.formal",
    "{ printf 'auth_mode = \"oidc\"\\n\\n'; cat ${local.config_path} 2>/dev/null; printf '\\n[oidc]\\nprovider = \"aws\"\\nintegration_id = \"${var.oidc_integration_id}\"\\n'; } > ${local.config_path}.new",
    "mv ${local.config_path}.new ${local.config_path}",
  ]
  setup = concat(local.ca_setup, local.oidc_setup)
  entry_point = length(local.setup) == 0 ? ["/usr/local/bin/formal"] : ["/bin/sh", "-c", join(" && ", concat(
    ["umask 077"],
    local.setup,
    ["exec /usr/local/bin/formal \"$@\""],
  )), "formal"]

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
      "GRPC_DEFAULT_SSL_ROOTS_FILE_PATH",
      "CURL_CA_BUNDLE",
    ] : name => local.ca_path
  }

  endpoint = merge(
    {
      name              = local.endpoint_name
      image             = var.image
      essential         = true
      user              = "0"
      entryPoint        = local.entry_point
      command           = concat(["agent", "--headless"], local.explicit ? ["--explicit-proxy=${local.proxy_addr}"] : [])
      cpu               = var.cpu
      memoryReservation = var.memory_reservation
      linuxParameters = {
        capabilities = {
          drop = ["ALL"]
          # NET_BIND_SERVICE: transparent DNS replies come from port 53, and
          # ECS Managed Instances does not lower ip_unprivileged_port_start.
          add = local.explicit ? [] : ["NET_ADMIN", "NET_BIND_SERVICE"]
        }
      }
      secrets = concat(var.api_key_secret_arn == null ? [] : [{
        name      = "FORMAL_API_KEY"
        valueFrom = var.api_key_secret_arn
        }], var.ca_secret_arn == null ? [] : [
        { name = "FORMAL_CA_CERT", valueFrom = "${var.ca_secret_arn}:tls.crt::" },
        { name = "FORMAL_CA_KEY", valueFrom = "${var.ca_secret_arn}:tls.key::" },
      ])
      mountPoints = [{
        sourceVolume  = local.volume_name
        containerPath = "/var/lib/formal/ca"
        readOnly      = false
      }]
      # App containers wait on HEALTHY, so they never send traffic before
      # interception is armed and the CA they are told to trust exists.
      healthCheck = {
        command     = ["CMD-SHELL", "test -s /var/lib/formal/ca/ca-bundle.pem && formal transparent-proxy status | grep -q '${local.ready_status}'"]
        interval    = 5
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    },
    var.log_configuration == null ? {} : { logConfiguration = var.log_configuration },
  )

  # Missing and null attributes both mean none.
  app_env = [for c in var.containers : {
    for e in(try(c.environment, null) == null ? [] : c.environment) : e.name => e.value
  }]
  # Keep the app's own exclusions.
  app_no_proxy = [for env in local.app_env : join(",", distinct(concat(
    compact(split(",", lookup(env, "NO_PROXY", lookup(env, "no_proxy", "")))),
    local.no_proxy,
  )))]
  # Java ignores the proxy variables and reads only system properties, which every JVM
  # also takes from JAVA_TOOL_OPTIONS. nonProxyHosts writes .example.com as
  # *.example.com.
  app_java_tool_options = [for i, env in local.app_env : trimspace(join(" ", [
    lookup(env, "JAVA_TOOL_OPTIONS", ""),
    "-Dhttps.proxyHost=${local.proxy_host}",
    "-Dhttps.proxyPort=${local.proxy_port}",
    "-Dhttp.proxyHost=${local.proxy_host}",
    "-Dhttp.proxyPort=${local.proxy_port}",
    "-Dhttp.nonProxyHosts=${join("|", [for h in split(",", local.app_no_proxy[i]) : startswith(h, ".") ? "*${h}" : h])}",
  ]))]

  apps = [
    for i, c in var.containers : merge(c, {
      environment = [
        for name, value in merge(
          local.app_env[i],
          local.ca_env,
          local.explicit ? {
            HTTPS_PROXY        = local.proxy_url
            https_proxy        = local.proxy_url
            HTTP_PROXY         = local.proxy_url
            http_proxy         = local.proxy_url
            NO_PROXY           = local.app_no_proxy[i]
            no_proxy           = local.app_no_proxy[i]
            NODE_USE_ENV_PROXY = "1"
            JAVA_TOOL_OPTIONS  = local.app_java_tool_options[i]
          } : {},
        ) : { name = name, value = value }
      ]
      mountPoints = concat(try(c.mountPoints, null) == null ? [] : c.mountPoints, [{
        sourceVolume  = local.volume_name
        containerPath = "/formal-ca"
        readOnly      = true
      }])
      dependsOn = concat(try(c.dependsOn, null) == null ? [] : c.dependsOn, [{
        containerName = local.endpoint_name
        condition     = "HEALTHY"
      }])
    })
  ]
}
