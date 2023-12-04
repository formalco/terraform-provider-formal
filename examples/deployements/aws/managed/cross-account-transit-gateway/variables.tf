variable "region" {}

variable "formal_api_key" {
  type      = string
  sensitive = true
}

variable "aws_access_key_account_1" {
  type      = string
  sensitive = true
}
variable "aws_secret_key_account_1" {
  type      = string
  sensitive = true
}

variable "aws_access_key_account_2" {
  type      = string
  sensitive = true
}
variable "aws_secret_key_account_2" {
  type      = string
  sensitive = true
}

variable "aws_account_2_id" {}

variable "name" {}

variable "redshift_username" {}

variable "redshift_password" {}
