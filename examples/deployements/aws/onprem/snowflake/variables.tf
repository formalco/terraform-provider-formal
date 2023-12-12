variable "region" {
  default = "eu-west-3"
}

variable "formal_api_key" {
  type      = string
  sensitive = true
}

variable "name" {}
variable "environment" {}

variable "cidr" {
  default = "172.0.0.0/16"
}
variable "private_subnets" {
  default = ["172.0.0.0/20"]
}
variable "public_subnets" {
  default = ["172.0.16.0/20"]
}
variable "availability_zones" {
  default = ["eu-west-3a"]
}

variable "datadog_api_key" {}

variable "dockerhub_username" {}
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
variable "snowflake_port" {
  default = 443
}
variable "data_classifier_satellite_port" {
  default = 50055
}

variable "snowflake_container_image" {}
variable "data_classifier_satellite_container_image" {}


variable "snowflake_sidecar_hostname" {}
variable "snowflake_hostname" {}

variable "snowflake_username" {}
variable "snowflake_password" {}