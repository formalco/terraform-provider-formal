variable "name" {}

variable "environment" {}

variable "vpc_id" {}
variable "main_port" {}

variable "datadog_api_key" {}
variable "formal_api_key" {}

variable "ecs_cluster_name" {}
variable "ecs_cluster_id" {}

variable "health_check_port" {}
variable "ssh_hostname" {}

variable "iam_access_key_id" {}
variable "iam_secret_access_key" {}

variable "private_subnets" {}
variable "public_subnets" {}

variable "docker_hub_secret_arn" {}

variable "container_image" {}

variable "container_cpu" {}

variable "container_memory" {}

variable "ssh_sidecar_hostname" {}