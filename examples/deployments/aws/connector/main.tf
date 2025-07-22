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
  cidr               = var.vpc_cidr
  private_subnets    = var.private_subnet_cidrs
  public_subnets     = var.public_subnet_cidrs
  availability_zones = var.availability_zones
  environment        = var.environment
}

# Generate Formal hostname for automatic TLS certificate management
locals {
  connector_hostname = "${var.name}.${var.formal_org_name}.connectors.joinformal.com"
}

module "demo_connector" {
  source                         = "./connector"
  formal_api_key                 = var.formal_api_key
  name                           = "${var.name}-demo-connector"
  connector_hostname             = local.connector_hostname
  connector_dns_record           = module.common.url
  environment                    = var.environment
  container_image                = var.connector_image
  vpc_id                         = module.common.vpc_id
  ecs_task_execution_role_arn    = module.common.ecs_task_execution_role_arn
  ecs_task_role_arn              = module.common.ecs_task_role_arn
  ecs_cluster_id                 = module.common.ecs_cluster_id
  ecs_cluster_name               = module.common.ecs_cluster_name
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
  container_cpu                  = var.container_cpu
  container_memory               = var.container_memory
}

resource "formal_resource" "echo" {
  name        = "${var.name}-echo"
  technology  = "http"
  hostname    = "echo.free.beeceptor.com"
  port        = 443
}

resource "formal_connector_listener" "echo" {
  name         = "${var.name}-echo"
  port         = 8443
}

resource "formal_connector_listener_rule" "echo" {
  connector_listener_id = formal_connector_listener.echo.id
  type                  = "resource"
  rule                  = formal_resource.echo.id
}

resource "formal_connector_listener_link" "echo" {
  connector_id          = module.demo_connector.connector_id
  connector_listener_id = formal_connector_listener.echo.id
}
