variable "region" {
  default = "ap-southeast-3"
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
  default = ["ap-southeast-3a", "ap-southeast-3b", "ap-southeast-3c"]
}

variable "datadog_api_key" {}

variable "dockerhub_username" {
  default = "lorisformal"
}
variable "dockerhub_password" {
  default = "nbm.XWE4hbn.uqc*pde"
}

variable "container_cpu" {
  default = 512
}
variable "container_memory" {
  default = 1024
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
  default = "formalco/docker-s-prod-snow-sidecar:master"
}
variable "data_classifier_satellite_container_image" {
  default = "formalco/docker-prod-data-classifier-satellite"
}

variable "snowflake_sidecar_hostname" {
  default = "snow-formal-demo.proxy.formalcloud.net"
}
variable "snowflake_hostname" {
  default = "gv36203.eu-west-1.snowflakecomputing.com"
}

variable "snowflake_username" {
  default = "ahmb84"
}
variable "snowflake_password" {
  default = "dpb_uhw_drq0amt2MKF"
}

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
  default = "pg-formal-demo.proxy.formalcloud.net"
}

variable "postgres_username" {
  default = "formal"
}
variable "postgres_password" {
  default = "FormalDemo1234"
}

variable "http_container_image" {
  default = "formalco/docker-prod-http-sidecar"
}

variable "http_sidecar_hostname" {
  default = "http-formal-demo.proxy.formalcloud.net"
}
variable "http_hostname" {
  default = "api.stripe.com"
}

variable "s3_sidecar_hostname" {
  default = "s3-formal-demo.proxy.formalcloud.net"
}
variable "s3_container_image" {
  default = "formalco/docker-prod-s3-sidecar:2.1.0"
}

variable "ssh_container_image" {
  default = "formalco/docker-prod-ssh-sidecar:latest"
}

variable "bucket_name" {
  default = "formal-demo-us"
}

variable "redshift_sidecar_hostname" {
  default = "rd-formal-demo.proxy.formalcloud.net"
}

variable "ssh_sidecar_hostname" {
  default = "ssh-formal-demo.proxy.formalcloud.net"
}

variable "redshift_container_image" {
  default = "formalco/docker-prod-redshift-sidecar:1.6.3"
}

variable "redshift_username" {
  default = "formal"
}
variable "redshift_password" {
  default = "Formaltest1234"
}

variable "mysql_port" {
  default = 3306
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

variable "http_certificane_arn" {
  default = "arn:aws:acm:ap-southeast-3:879131189120:certificate/afbec427-348e-4a23-8b51-521f8582c4e6"
}