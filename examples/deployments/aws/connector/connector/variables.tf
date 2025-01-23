variable "name" {}

variable "environment" {}

variable "datadog_api_key" {}

variable "health_check_port" {}

variable "container_image" {}
variable "container_cpu" {}
variable "container_memory" {}

variable "vpc_id" {}
variable "formal_api_key" {}

variable "ecs_cluster_id" {}
variable "ecs_cluster_name" {}

variable "private_subnets" {}
variable "public_subnets" {}

variable "data_classifier_satellite_url" {}
variable "data_classifier_satellite_port" {}

variable "ecs_task_execution_role_arn" {}
variable "ecs_task_role_arn" {}

variable "connector_hostname" {}
variable "connector_dns_record" {}
variable "connector_postgres_listener_name" {}
variable "connector_postgres_listener_port" {}

variable "connector_mysql_port" {}

variable "connector_kubernetes_listener_name" {}
variable "connector_kubernetes_listener_port" {
  default = 443
}