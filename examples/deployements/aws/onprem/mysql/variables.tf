variable "region" {
  default = "eu-west-3"
}

variable "formal_api_key" {
  type      = string
  sensitive = true
}

variable "name" {
  default = "mysql"
}
variable "environment" {
  default = "demo-formal"
}

variable "cidr" {
  default = "172.0.0.0/16"
}
variable "private_subnets" {
  default = ["172.0.0.0/20", "172.0.32.0/20", "172.0.64.0/20"]
}
variable "public_subnets" {
  default = ["172.0.16.0/20", "172.0.48.0/20", "172.0.80.0/20"]
}
variable "availability_zones" {
  default = ["eu-west-3a", "eu-west-3b", "eu-west-3c"]
}

variable "datadog_api_key" {}

variable "dockerhub_username" {
  default = "lorisformal"
}
variable "dockerhub_password" {}

variable "container_cpu" {
  default = 2048
}
variable "container_memory" {
  default = 4096
}

variable "health_check_port" {
  default = 8080
}
variable "mysql_port" {
  default = 3306
}
variable "data_classifier_satellite_port" {
  default = 50055
}

variable "data_classifier_satellite_container_image" {
  default = "formalco/docker-prod-data-classifier-satellite"
}


variable "mysql_sidecar_hostname" {
  default = "mysql-formal-demo.proxy.formalcloud.net"
}

variable "mysql_container_image" {
  default = "formalco/dockerm-prod-mysql-sidecar"
}

variable "mysql_username" {
  default = "formal"
}
variable "mysql_password" {
  default = "FormalDemo1234"
}
