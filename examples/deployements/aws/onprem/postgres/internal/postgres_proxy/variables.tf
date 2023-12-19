variable "name" {}

variable "environment" {}

variable "datadog_api_key" {}

variable "health_check_port" {}
variable "main_port" {}

variable "container_image" {}

variable "container_cpu" {}

variable "container_memory" {}

variable "vpc_id" {}
variable "docker_hub_secret_arn" {}
variable "formal_api_key" {}

variable "ecs_cluster_id" {}
variable "ecs_cluster_name" {}

variable "private_subnets" {}
variable "public_subnets" {}

variable "postgres_hostname" {}

variable "postgres_sidecar_hostname" {}

variable "postgres_username" {}
variable "postgres_password" {}