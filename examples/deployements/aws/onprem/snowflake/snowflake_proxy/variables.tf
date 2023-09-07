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

variable "data_classifier_satellite_url" {}
variable "data_classifier_satellite_port" {}

variable "snowflake_hostname" {}

variable "snowflake_sidecar_hostname" {}

variable "snowflake_username" {}
variable "snowflake_password" {}