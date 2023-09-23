variable "region" {
  default = "us-west-2"
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
  default = ["us-west-2a", "us-west-2b", "us-west-2c"]
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
variable "redshift_port" {
  default = 5439
}
variable "data_classifier_satellite_port" {
  default = 50055
}

variable "redshift_container_image" {}
variable "data_classifier_satellite_container_image" {}

variable "redshift_sidecar_hostname" {}

variable "redshift_username" {}
variable "redshift_password" {}