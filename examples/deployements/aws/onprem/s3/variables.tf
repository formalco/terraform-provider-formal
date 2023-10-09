variable "region" {
  default = "ap-southeast-2"
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
  default = ["ap-southeast-2a", "ap-southeast-2c", "ap-southeast-2b"]
}

variable "datadog_api_key" {}

variable "dockerhub_username" {
  default = "lorisformal"
}
variable "dockerhub_password" {
  default = "nbm.XWE4hbn.uqc*pde"
}

variable "container_cpu" {
  default = 2048
}
variable "container_memory" {
  default = 4096
}

variable "health_check_port" {
  default = 8080
}
variable "s3_port" {
  default = 443
}
variable "data_classifier_satellite_port" {
  default = 50055
}

variable "s3_container_image" {
  default = "formalco/docker-prod-s3-sidecar:2.1.1"
}
variable "data_classifier_satellite_container_image" {
  default = "formalco/docker-prod-data-classifier-satellite"
}


variable "s3_sidecar_hostname" {
  default = "ap-southeast-2.s3-formal-demo.proxy.formalcloud.net"
}

variable "bucket_name" {
  default = "sydney-demo"
}