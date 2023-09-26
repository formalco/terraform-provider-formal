terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.53.0"
    }
    formal = {
      source  = "formalco/formal"
      version = "~>3.2.3"
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
  dockerhub_username = var.dockerhub_username
  dockerhub_password = var.dockerhub_password
}

module "data_classifier_satellite" {
  source                = "./data_classifier_satellite"
  name                  = "${var.name}-data-classifier-satelitte"
  environment           = var.environment
  formal_api_key        = var.formal_api_key
  main_port             = var.data_classifier_satellite_port
  health_check_port     = var.health_check_port
  datadog_api_key       = var.datadog_api_key
  container_image       = var.data_classifier_satellite_container_image
  container_cpu         = var.container_cpu
  container_memory      = var.container_memory
  vpc_id                = module.common.vpc_id
  docker_hub_secret_arn = module.common.docker_hub_secret_arn
  ecs_cluster_id        = module.common.ecs_cluster_id
  ecs_cluster_name      = module.common.ecs_cluster_name
  private_subnets       = module.common.private_subnets
  public_subnets        = module.common.public_subnets
}

module "snowflake_proxy" {
  source                         = "./snowflake_proxy"
  name                           = "${var.name}-snowflake-proxy"
  environment                    = var.environment
  formal_api_key                 = var.formal_api_key
  main_port                      = var.snowflake_port
  snowflake_sidecar_hostname     = var.snowflake_sidecar_hostname
  snowflake_hostname             = var.snowflake_hostname
  health_check_port              = var.health_check_port
  datadog_api_key                = var.datadog_api_key
  container_image                = var.snowflake_container_image
  container_cpu                  = var.container_cpu
  container_memory               = var.container_memory
  vpc_id                         = module.common.vpc_id
  docker_hub_secret_arn          = module.common.docker_hub_secret_arn
  ecs_cluster_id                 = module.common.ecs_cluster_id
  ecs_cluster_name               = module.common.ecs_cluster_name
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
  data_classifier_satellite_url  = module.data_classifier_satellite.url
  data_classifier_satellite_port = var.data_classifier_satellite_port
  snowflake_username             = var.snowflake_username
  snowflake_password             = var.snowflake_password
}

module "postgres_proxy" {
  source                         = "./postgres_proxy"
  name                           = "${var.name}-postgres-proxy"
  environment                    = var.environment
  formal_api_key                 = var.formal_api_key
  main_port                      = var.postgres_port
  postgres_sidecar_hostname      = var.postgres_sidecar_hostname
  postgres_hostname              = module.postgres_proxy.rds_hostname
  health_check_port              = var.health_check_port
  datadog_api_key                = var.datadog_api_key
  container_image                = var.postgres_container_image
  container_cpu                  = var.container_cpu
  container_memory               = var.container_memory
  vpc_id                         = module.common.vpc_id
  docker_hub_secret_arn          = module.common.docker_hub_secret_arn
  ecs_cluster_id                 = module.common.ecs_cluster_id
  ecs_cluster_name               = module.common.ecs_cluster_name
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
  data_classifier_satellite_url  = module.data_classifier_satellite.url
  data_classifier_satellite_port = var.data_classifier_satellite_port
  postgres_username              = var.postgres_username
  postgres_password              = var.postgres_password
}

module "http_proxy" {
  source                         = "./http_proxy"
  name                           = "${var.name}-http-proxy"
  environment                    = var.environment
  formal_api_key                 = var.formal_api_key
  main_port                      = var.http_port
  health_check_port              = var.health_check_port
  datadog_api_key                = var.datadog_api_key
  container_image                = var.http_container_image
  container_cpu                  = var.container_cpu
  container_memory               = var.container_memory
  vpc_id                         = module.common.vpc_id
  docker_hub_secret_arn          = module.common.docker_hub_secret_arn
  ecs_cluster_id                 = module.common.ecs_cluster_id
  ecs_cluster_name               = module.common.ecs_cluster_name
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
  data_classifier_satellite_url  = module.data_classifier_satellite.url
  data_classifier_satellite_port = var.data_classifier_satellite_port
  sidecar_hostname               = var.http_sidecar_hostname
  datastore_hostname             = var.http_hostname
}

module "redshift_proxy" {
  source                         = "./redshift_proxy"
  name                           = "${var.name}-redshift-proxy"
  environment                    = var.environment
  formal_api_key                 = var.formal_api_key
  main_port                      = var.redshift_port
  health_check_port              = var.health_check_port
  datadog_api_key                = var.datadog_api_key
  container_image                = var.postgres_container_image
  container_cpu                  = var.container_cpu
  container_memory               = var.container_memory
  vpc_id                         = module.common.vpc_id
  docker_hub_secret_arn          = module.common.docker_hub_secret_arn
  ecs_cluster_id                 = module.common.ecs_cluster_id
  ecs_cluster_name               = module.common.ecs_cluster_name
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
  data_classifier_satellite_url  = module.data_classifier_satellite.url
  data_classifier_satellite_port = var.data_classifier_satellite_port
  redshift_username              = var.redshift_username
  redshift_password              = var.redshift_password
}

module "s3_proxy" {
  source                         = "./s3_proxy"
  name                           = "${var.name}-s3-proxy"
  environment                    = var.environment
  formal_api_key                 = var.formal_api_key
  main_port                      = var.postgres_port
  health_check_port              = var.health_check_port
  datadog_api_key                = var.datadog_api_key
  container_image                = var.postgres_container_image
  container_cpu                  = var.container_cpu
  container_memory               = var.container_memory
  vpc_id                         = module.common.vpc_id
  docker_hub_secret_arn          = module.common.docker_hub_secret_arn
  ecs_cluster_id                 = module.common.ecs_cluster_id
  ecs_cluster_name               = module.common.ecs_cluster_name
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
}

module "ssh_proxy" {
  source                         = "./ssh_proxy"
  name                           = "${var.name}-ssh-proxy"
  environment                    = var.environment
  formal_api_key                 = var.formal_api_key
  health_check_port              = var.health_check_port
  datadog_api_key                = var.datadog_api_key
  container_image                = var.ssh_container_image
  container_cpu                  = var.container_cpu
  container_memory               = var.container_memory
  vpc_id                         = module.common.vpc_id
  docker_hub_secret_arn          = module.common.docker_hub_secret_arn
  ecs_cluster_id                 = module.common.ecs_cluster_id
  ecs_cluster_name               = module.common.ecs_cluster_name
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
}