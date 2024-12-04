variable "region" {}

variable "formal_api_key" {
  type      = string
  sensitive = true
}

variable "name" {
  default = "demo-env"
}
