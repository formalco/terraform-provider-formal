terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.7.0"
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

module "eks" {
  source               = "./eks"
  name                 = "${var.name}-bigquery-proxy"
  environment          = var.environment
  vpc_id               = module.common.vpc_id
  private_subnets      = module.common.private_subnets
  public_subnets       = module.common.public_subnets
  formal_bigquery_api_key = module.formal.formal_bigquery_sidecar_api_key
}

module "formal" {
  source                 = "./formal"
  name                   = "${var.name}-bigquery-proxy"
  environment            = var.environment
  formal_api_key         = var.formal_api_key
  main_port              = var.bigquery_port
  availability_zones     = var.availability_zones
  bigquery_sidecar_hostname = var.bigquery_sidecar_hostname
  bigquery_hostname         = "bigquery.googleapis.com"
  vpc_id                 = module.common.vpc_id
  public_subnets         = module.common.public_subnets
  bigquery_username         = var.bigquery_username
  bigquery_password         = var.bigquery_password
}

module "helm" {
  source                                 = "./helm"
  eks_cluster_name                       = module.eks.aws_eks_cluster_name
  eks_cluster_endpoint                   = module.eks.aws_eks_cluster_endpoint
  eks_cluster_certificate_authority_data = module.eks.aws_eks_cluster_ca_cert
  chart_oci                              = var.chart_oci
}