variable "region" {
  default = "eu-west-2"
}

variable "formal_api_key" {
  type      = string
  sensitive = true
}

variable "aws_access_key" {
  type      = string
  sensitive = true
}
variable "aws_secret_key" {
  type      = string
  sensitive = true
}

variable "name" {}

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

variable "postgres_username" {}
variable "postgres_password" {}

variable "redshift_username" {}
variable "redshift_password" {}