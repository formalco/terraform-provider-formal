variable "name" {}

variable "environment" {}

variable "health_check_port" {}
variable "main_port" {}

variable "container_image" {}

variable "container_cpu" {}

variable "container_memory" {}

variable "vpc_id" {}
variable "formal_api_key" {}

variable "ecs_cluster_id" {}
variable "ecs_cluster_name" {}

variable "private_subnets" {}
variable "public_subnets" {}

variable "snowflake_hostname" {}

variable "snowflake_sidecar_hostname" {}

variable "snowflake_username" {}
variable "snowflake_password" {}

variable "log_configuration" {
    default = null
}

variable "sidecar_container_dependencies" {
  type = list(object({
    containerName = string
    condition     = string
  }))
  default = []
}

variable "sidecar_container_definitions" {
  type = list(object({
    name              = string
    image             = string
    memoryReservation = number
    firelensConfiguration = optional(object({
      type    = string
      options = map(string)
    }))
    portMappings = optional(list(object({
      containerPort = number
      hostPort      = number
      protocol      = string
    })))
    environment = optional(list(object({
      name  = string
      value = string
    })))
    healthCheck = optional(object({
      command  = list(string)
      interval = number
      timeout  = number
      retries  = number
    }))
  }))
  default = []
}

variable "ecs_enviroment_variables" {
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}