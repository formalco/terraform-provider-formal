variable "region" {
  default = "ap-southeast-1"
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
  default = ["ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"]
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

variable "snowflake_container_image" {
  default = "formalco/docker-s-prod-snow-sidecar"
}
variable "data_classifier_satellite_container_image" {
  default = "formalco/docker-prod-data-classifier-satellite"
}

variable "snowflake_sidecar_hostname" {
  default = "snow-lolo.proxy.formalcloud.net"
}
variable "snowflake_hostname" {}

variable "snowflake_username" {}
variable "snowflake_password" {}

variable "http_port" {
  default = 443
}

variable "postgres_port" {
  default = 5432
}
variable "s3_port" {
  default = 443
}
variable "ssh_port" {
  default = 2022
}
variable "redshift_port" {
  default = 5439
}

variable "postgres_container_image" {
  default = "formalco/docker-prod-pg-sidecar"
}
variable "postgres_sidecar_hostname" {
  default = "pg-lolo.proxy.formalcloud.net"
}

variable "postgres_username" {}
variable "postgres_password" {}

variable "http_container_image" {
  default = "formalco/docker-prod-http-sidecar"
}

variable "http_sidecar_hostname" {
  default = "http-lolo.proxy.formalcloud.net"
}
variable "http_hostname" {
  default = "http-lolo.proxy.formalcloud.net"
}

variable "s3_sidecar_hostname" {
  default = "s3-lolo.proxy.formalcloud.net"
}
variable "s3_container_image" {
  default = "formalco/docker-prod-s3-sidecar"
}

variable "bucket_name" {}

variable "redshift_sidecar_hostname" {
  default = "rd-lolo.proxy.formalcloud.net"
}
variable "redshift_container_image" {
  default = "formalco/docker-prod-redshift-sidecar"
}

variable "redshift_username" {}
variable "redshift_password" {}