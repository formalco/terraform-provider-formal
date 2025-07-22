terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.47.0"
    }
    formal = {
      source  = "formalco/formal"
      version = "4.0.15"
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

module "demo_connector" {
  source                         = "./connector"
  formal_api_key                 = var.formal_api_key
  name                           = "${var.name}-demo-connector"
  connector_hostname             = var.demo_connector_hostname
  health_check_port              = var.health_check_port
  environment                    = var.environment
  container_image                = var.demo_connector_container_image
  vpc_id                         = module.common.vpc_id
  ecs_task_execution_role_arn    = module.common.ecs_task_execution_role_arn
  ecs_task_role_arn              = module.common.ecs_task_role_arn
  ecs_cluster_id                 = module.common.ecs_cluster_id
  ecs_cluster_name               = module.common.ecs_cluster_name
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
  container_cpu                  = var.container_cpu
  container_memory               = var.container_memory
  data_classifier_satellite_url  = module.common.url
  data_classifier_satellite_port = var.data_classifier_satellite_port
}

