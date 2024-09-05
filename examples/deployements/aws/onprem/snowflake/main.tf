terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.53.0"
    }
    formal = {
      source  = "formalco/formal"
      version = "~>4.0.0"
    }
  }

  required_version = ">= 0.14.9"
}

provider "aws" {
  region = var.region
}

provider "formal" {
  api_key = var.formal_api_key
}

module "common" {
  source             = "./common"
  name               = var.name
  cidr               = var.cidr
  private_subnets    = var.private_subnets
  public_subnets     = var.public_subnets
  availability_zones = var.availability_zones
  environment        = var.environment
}

module "data_classifier_satellite" {
  source            = "./data_classifier_satellite"
  name              = "${var.name}-data-classifier-satelitte"
  environment       = var.environment
  formal_api_key    = var.formal_api_key
  main_port         = var.data_classifier_satellite_port
  health_check_port = var.health_check_port
  datadog_api_key   = var.datadog_api_key
  container_image   = var.data_classifier_satellite_container_image
  container_cpu     = var.container_cpu
  container_memory  = var.container_memory
  vpc_id            = module.common.vpc_id
  ecs_cluster_id    = module.common.ecs_cluster_id
  ecs_cluster_name  = module.common.ecs_cluster_name
  private_subnets   = module.common.private_subnets
  public_subnets    = module.common.public_subnets
}

module "snowflake_proxy" {
  source                     = "./snowflake_proxy"
  name                       = "${var.name}-snowflake-proxy"
  environment                = var.environment
  formal_api_key             = var.formal_api_key
  main_port                  = var.snowflake_port
  snowflake_sidecar_hostname = var.snowflake_sidecar_hostname
  snowflake_hostname         = var.snowflake_hostname
  health_check_port          = var.health_check_port
  container_image            = var.snowflake_container_image
  container_cpu              = var.container_cpu
  container_memory           = var.container_memory
  vpc_id                     = module.common.vpc_id
  ecs_cluster_id             = module.common.ecs_cluster_id
  ecs_cluster_name           = module.common.ecs_cluster_name
  private_subnets            = module.common.private_subnets
  public_subnets             = module.common.public_subnets
  snowflake_username         = var.snowflake_username
  snowflake_password         = var.snowflake_password
  log_configuration = {
    logDriver = "awsfirelens"
    options = {
      "Name"       = "datadog",
      "Host"       = "http-intake.logs.datadoghq.eu",
      "TLS"        = "on",
      "dd_source"  = var.name,
      "provider"   = "ecs",
      "dd_service" = var.name,
      "apikey"     = var.datadog_api_key
    }
  }
  sidecar_container_definitions = [
    {
      name              = "log_router"
      image             = "public.ecr.aws/aws-observability/aws-for-fluent-bit:stable"
      memoryReservation = 50
      firelensConfiguration = {
        "type" = "fluentbit"
        "options" = {
          "enable-ecs-log-metadata" = "true"
        }
      }
    },
    {
      name              = "datadog-agent"
      image             = "public.ecr.aws/datadog/agent:latest"
      memoryReservation = 128
      portMappings = [
        {
          containerPort = 8126
          hostPort      = 8126
          protocol      = "tcp"
        }
      ]
      environment = [
        { name = "ECS_FARGATE", value = "true" },
        { name = "DD_APM_ENABLED", value = "true" },
        { name = "DD_LOGS_ENABLED", value = "true" },
        { name = "DD_LOGS_CONFIG_CONTAINER_COLLECT_ALL", value = "true" },
        { name = "DD_APM_NON_LOCAL_TRAFFIC", value = "true" },
        { name = "DD_API_KEY", value = var.datadog_api_key },
        { name = "DD_SITE", value = "datadoghq.eu" }
      ]
      healthCheck = {
        command  = ["CMD-SHELL", "agent health"]
        interval = 30
        timeout  = 5
        retries  = 3
      }
    }
  ]
  sidecar_container_dependencies = [
    { containerName = "log_router", condition = "START" },
    { containerName = "datadog-agent", condition = "HEALTHY" }
  ]
  ecs_enviroment_variables = [
    {
      name  = "DATA_CLASSIFIER_SATELLITE_URI"
      value = "${module.data_classifier_satellite.url}:${var.data_classifier_satellite_port}"
    },
    {
      name  = "SERVER_CONNECT_TLS"
      value = "true"
    },
    {
      name  = "CLIENT_LISTEN_TLS"
      value = "true"
    },
    {
      name  = "DD_VERSION"
      value = "1.0.0"
    },
    {
      name  = "DD_ENV"
      value = "prod"
    },
    {
      name  = "DD_SERVICE"
      value = var.name
    },
    {
      name  = "MANAGED_TLS_CERTS"
      value = "false"
    },
    {
      name  = "PII_SAMPLING_RATE"
      value = "8"
    }
  ]
}