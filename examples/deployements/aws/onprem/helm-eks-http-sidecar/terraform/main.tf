terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.7.0"
    }
    formal = {
      source  = "formalco/formal"
      version = "~>3.2.3"
    }
    helm = {
      source = "hashicorp/helm"
      version = "~>2.11.0"
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

module "eks" {
  source                         = "./eks"
  name                           = "${var.name}-http-proxy"
  environment                    = var.environment
  vpc_id                         = module.common.vpc_id
  private_subnets                = module.common.private_subnets
  public_subnets                 = module.common.public_subnets
}