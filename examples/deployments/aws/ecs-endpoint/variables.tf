variable "containers" {
  description = "Application container definitions, in the ECS task definition JSON shape. Each one gets the Formal CA and waits for the Endpoint to be healthy."
  type        = any
}

variable "api_key_secret_arn" {
  description = "ARN of the Secrets Manager secret (or SSM parameter) holding the Formal machine user API key. The task execution role must be able to read it."
  type        = string
}

variable "image" {
  description = "Formal Endpoint image."
  type        = string
  default     = "654654333078.dkr.ecr.us-east-1.amazonaws.com/formalco-prod-kubernetes-egress:latest"
}

variable "cpu" {
  description = "CPU units reserved for the Endpoint container."
  type        = number
  default     = 1024
}

variable "memory_reservation" {
  description = "Memory (MiB) reserved for the Endpoint container."
  type        = number
  default     = 2048
}

variable "log_configuration" {
  description = "Optional logConfiguration for the Endpoint container."
  type        = any
  default     = null
}
