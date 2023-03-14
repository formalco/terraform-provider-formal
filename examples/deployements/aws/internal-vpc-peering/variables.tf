variable "region" {
  default = "eu-west-2"
}

variable "formal_client_id" {}

variable "formal_secret_key" {}

variable "aws_access_key" {}

variable "aws_secret_key" {}

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