# Required variables
variable "region" {
  description = "AWS region for deployment"
  type        = string
}

variable "availability_zones" {
  description = "List of availability zones (e.g., [\"us-west-2a\", \"us-west-2b\"])"
  type        = list(string)
}

variable "formal_api_key" {
  description = "Your Formal API key (provided by Formal)"
  type        = string
  sensitive   = true
}

variable "formal_org_name" {
  description = "Your Formal organization name (provided by Formal)"
  type        = string
}

# Optional variables
variable "name" {
  description = "Name prefix for all resources"
  type        = string
  default     = "demo"
}

variable "environment" {
  description = "Environment name for resource tagging"
  type        = string
  default     = "demo-formal-env"
}


# Networking
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "172.0.0.0/16"
}

variable "private_subnet_cidrs" {
  description = "List of CIDR blocks for private subnets"
  type        = list(string)
  default     = ["172.0.1.0/24", "172.0.2.0/24"]
}

variable "public_subnet_cidrs" {
  description = "List of CIDR blocks for public subnets"
  type        = list(string)
  default     = ["172.0.101.0/24", "172.0.102.0/24"]
}

# Container
variable "container_cpu" {
  description = "CPU units for the connector container"
  type        = number
  default     = 1024
}

variable "container_memory" {
  description = "Memory in MB for the connector container"
  type        = number
  default     = 2048
}

variable "connector_image" {
  description = "Container image for the Formal connector"
  type        = string
  default     = "654654333078.dkr.ecr.eu-west-1.amazonaws.com/formalco-prod-connector:latest"
}

