variable "region" {
  default = "us-east-1"
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
  default = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "chart_oci" {
  default = "oci://public.ecr.aws/d6o8b0b1/formal-mysql-helm-chart"
}
variable "mysql_port" {
  default = 3306
}
variable "mysql_sidecar_hostname" {}
variable "mysql_username" {
  type      = string
  sensitive = true
}
variable "mysql_password" {
  type      = string
  sensitive = true
}