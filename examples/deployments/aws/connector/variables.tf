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


variable "private_subnets" {}
variable "public_subnets" {}
variable "availability_zones" {}



variable "container_cpu" {
  default = 1024
}
variable "container_memory" {
  default = 2048
}

variable "demo_connector_hostname" {}
variable "demo_connector_dns_record" {}



variable "demo_connector_container_image" {}

