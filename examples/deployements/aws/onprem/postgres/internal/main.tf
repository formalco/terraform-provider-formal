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
  postgres_username              = var.postgres_username
  postgres_password              = var.postgres_password
}

