# Update the variables according to your requirement!

variable "project_name" {
  description = "Project Name - will prefex all generated AWS resource names"
  default     = "tgw-formal"
}

variable "region" {
  default = "eu-central-1"
}

data "aws_availability_zones" "azs" {
}

variable "edge_sg_vpc_cidr" {
  description = "Edge VPC CIDR"
  default     = "10.7.0.0/16"
}

variable "spoke_1_sg_vpc_cidr" {
  description = "Spoke VPC 1 CIDR"
  default     = "10.10.0.0/16"
}

variable "key_name" {
  description = "SSH Key Pair"
  default     = "test"
}

variable "formal_client_id" {}

variable "formal_secret_key" {}

variable "aws_access_key_account_1" {}

variable "aws_secret_key_account_1" {}

variable "aws_access_key_account_2" {}

variable "aws_secret_key_account_2" {}

variable "aws_account_2_id" {}

variable "name" {}

variable "cloud_account_id" {}

variable "customer_vpc_id" {}

variable "postgres_username" {}

variable "postgres_password" {}
