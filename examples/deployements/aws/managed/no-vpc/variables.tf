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

variable "snowflake_username" {}
variable "snowflake_password" {}

variable "snowflake_hostname" {}
variable "snowflake_port" {}