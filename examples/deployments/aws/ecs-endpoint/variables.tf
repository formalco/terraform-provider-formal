variable "containers" {
  description = "Application container definitions, in the ECS task definition JSON shape. Each one gets the Formal CA and waits for the Endpoint to be healthy."
  type        = any
}

variable "launch_type" {
  description = "Where the task runs: EC2 (transparent interception) or FARGATE (apps reach the Endpoint through HTTPS_PROXY)."
  type        = string
  default     = "EC2"

  validation {
    condition     = contains(["EC2", "FARGATE"], var.launch_type)
    error_message = "launch_type must be EC2 or FARGATE."
  }
}

variable "oidc_integration_id" {
  description = "Formal OIDC integration the Endpoint signs in with through the task role, which needs sts:GetWebIdentityToken. Set this or api_key_secret_arn."
  type        = string
  default     = null

  validation {
    condition     = var.oidc_integration_id == null || can(regex("^integrationoidc_[0-9a-z]+$", var.oidc_integration_id))
    error_message = "oidc_integration_id must be a Formal OIDC integration ID, such as integrationoidc_01abc..."
  }
}

variable "api_key_secret_arn" {
  description = "ARN of the Secrets Manager secret (or SSM parameter) holding a Formal machine user API key, when not using oidc_integration_id. The task execution role must be able to read it."
  type        = string
  default     = null
}

variable "ca_secret_arn" {
  description = "Optional Secrets Manager secret with the CA the Endpoint signs intercepted TLS with, such as an intermediate of your PKI: JSON with tls.crt (the CA, then its issuers) and tls.key. The task execution role must be able to read it. When null, the Endpoint generates a CA per task."
  type        = string
  default     = null

  validation {
    condition     = var.ca_secret_arn == null || can(regex("^arn:[^:]+:secretsmanager:", var.ca_secret_arn))
    error_message = "ca_secret_arn must be a Secrets Manager secret ARN."
  }
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
