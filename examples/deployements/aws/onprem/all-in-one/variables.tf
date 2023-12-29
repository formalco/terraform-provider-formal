variable "region" {}

variable "formal_api_key" {
  type      = string
  sensitive = true
}

variable "name" {}
variable "environment" {}

variable "cidr" {}
variable "private_subnets" {}
variable "public_subnets" {}
variable "availability_zones" {}

variable "datadog_api_key" {}

variable "dockerhub_username" {}
variable "dockerhub_password" {}

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

variable "snowflake_container_image" {}
variable "data_classifier_satellite_container_image" {}

variable "snowflake_sidecar_hostname" {}
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

variable "postgres_container_image" {}
variable "postgres_sidecar_hostname" {}

variable "postgres_username" {}
variable "postgres_password" {}

variable "http_container_image" {}

variable "http_sidecar_hostname" {}
variable "http_hostname" {}

variable "s3_sidecar_hostname" {}
variable "s3_container_image" {}

variable "ssh_container_image" {}

variable "bucket_name" {}

variable "redshift_sidecar_hostname" {}

variable "ssh_sidecar_hostname" {}

variable "redshift_container_image" {}

variable "redshift_username" {}
variable "redshift_password" {}

variable "mysql_port" {
  default = 3306
}

variable "mysql_sidecar_hostname" {}

variable "mysql_container_image" {}

variable "mysql_username" {}
variable "mysql_password" {}
