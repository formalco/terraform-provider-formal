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

variable "cidr" {
  default = "172.0.0.0/16"
}

variable "data_classifier_satellite_port" {
  default = 50055
}

variable "private_subnets" {}
variable "public_subnets" {}
variable "availability_zones" {}

variable "datadog_api_key" {}

variable "dockerhub_username" {}
variable "dockerhub_password" {}

variable "container_cpu" {
  default = 1024
}
variable "container_memory" {
  default = 2048
}

variable "demo_connector_hostname" {}
variable "demo_connector_postgres_listener_name" {}
variable "demo_connector_postgres_listener_port" {
  default = 5432
}
variable "connector_mysql_port" {
  default = 3306
}
variable "connector_kubernetes_listener_name" {}
variable "connector_kubernetes_listener_port" {
  default = 443
}

variable "connector_clickhouse_port" {
  default = 8443
}

variable "connector_s3_browser_port" {
  default = 9200
}

variable "demo_connector_container_image" {}

