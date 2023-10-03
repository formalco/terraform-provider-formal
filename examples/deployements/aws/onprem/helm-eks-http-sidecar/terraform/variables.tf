variable "region" {
  default = "ap-southeast-2"
}

# variable "formal_api_key" {
#   type      = string
#   sensitive = true
# }

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
  default = ["ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c"]
}

variable "container_cpu" {
  default = 2048
}
variable "container_memory" {
  default = 4096
}

variable "chart_oci" {
  default = "oci://public.ecr.aws/z9b2c2l8/formal-http-helm-chart"
}