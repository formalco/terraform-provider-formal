variable "region" {}

variable "formal_api_key" {
  type      = string
  sensitive = true
}

variable "health_check_port" {
  default = 8080
}

variable "name" {
  default = "demo-env"
}
variable "environment" {
  default = "demo-formal-env"
}

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

variable "availability_zones" {
  description = "List of availability zones (must match your region, e.g., [\"us-west-2a\", \"us-west-2b\"])"
  type        = list(string)
}



variable "container_cpu" {
  default = 1024
}
variable "container_memory" {
  default = 2048
}

variable "demo_connector_hostname" {}
variable "demo_connector_dns_record" {}



variable "connector_image" {
  description = "Container image for the Formal connector"
  type        = string
  default     = "654654333078.dkr.ecr.eu-west-1.amazonaws.com/formalco-prod-connector:latest"
}

