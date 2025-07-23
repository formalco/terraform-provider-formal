variable "name" {}

variable "environment" {}



variable "container_image" {}
variable "container_cpu" {}
variable "container_memory" {}

variable "vpc_id" {}
variable "formal_api_key" {}

variable "ecs_cluster_id" {}
variable "ecs_cluster_name" {}

variable "private_subnets" {}
variable "public_subnets" {}


variable "ecs_task_execution_role_arn" {}
variable "ecs_task_role_arn" {}

variable "connector_hostname" {}

variable "connector_ports" {
  description = "List of ports to open for the connector"
  type        = list(number)
}

