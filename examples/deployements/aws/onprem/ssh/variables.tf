variable "region" {
  default = "eu-west-2"
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
  default = ["172.0.0.0/20", "172.0.32.0/20", "172.0.64.0/20"]
}

variable "public_subnets" {
  default = ["172.0.16.0/20", "172.0.48.0/20", "172.0.80.0/20"]
}

variable "availability_zones" {
  default = ["eu-west-2a", "eu-west-2b", "eu-west-2c"]
}

variable "dockerhub_username" {}
variable "dockerhub_password" {}

variable "datadog_api_key" {}

variable "health_check_port" {
  default = 8080
}
variable "main_port" {
  default = 2022
}
variable "hostname" {}

variable "container_image" {}

variable "container_cpu" {
  default = 2048
}

variable "container_memory" {
  default = 4096
}