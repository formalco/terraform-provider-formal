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
  default = ""
}
variable "kubernetes_port" {
  default = 443
}
variable "kubernetes_sidecar_hostname" {}
variable "kubernetes_hostname" {}
variable "kubernetes_username" {
  type      = string
  sensitive = true
}
variable "kubernetes_password" {
  type      = string
  sensitive = true
}