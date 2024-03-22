terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.53.0"
    }
    formal = {
      source  = "formalco/formal"
      version = "~>3.4.0"
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

module "mysql_proxy" {
  source                         = "./mysql"
  name                           = "${var.name}-mysql-proxy"
  environment                    = var.environment
  formal_api_key                 = var.formal_api_key
  main_port                      = var.mysql_port
  mysql_sidecar_hostname         = var.mysql_sidecar_hostname
  mysql_hostname                 = module.mysql_proxy.rds_hostname
  health_check_port              = var.health_check_port
  datadog_api_key                = var.datadog_api_key
  container_image                = var.mysql_container_image
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
  mysql_username                 = var.mysql_username
  mysql_password                 = var.mysql_password
}

