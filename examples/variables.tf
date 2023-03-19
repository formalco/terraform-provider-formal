variable "formal_api_key" {
  type      = string
  sensitive = true
}

variable "datastore_technology" {
  type = string
}
variable "datastore_username" {
  type      = string
  sensitive = true
}
variable "datastore_password" {
  type      = string
  sensitive = true
}

variable "datastore_hostname" {
  type = string
}

variable "cloud_account_id" {
  type = string
}

variable "datastore_name" {
  type = string
}

variable "datastore_region" {
  type = string
}

variable "datastore_port" {
  type = number
}

variable "dataplane_id" {
  type = string
}

variable "native_role_secret" {
  type = string
}

