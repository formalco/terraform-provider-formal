# Required variables
variable "formal_api_key" {
  description = "Formal Control Plane API Key"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "AWS region for deployment"
  type        = string
}

# Networking variables (user-provided)
variable "vpc_id" {
  description = "ID of the VPC where the data discovery satellite will be deployed"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for the ECS service (should be private subnets with NAT gateway access)"
  type        = list(string)
}

variable "security_group_ids" {
  description = "List of security group IDs to attach to the ECS service"
  type        = list(string)
}

# ECS cluster variable (user-provided)
variable "ecs_cluster_arn" {
  description = "ARN of the ECS cluster where the data discovery satellite will be deployed"
  type        = string
}

# Optional variables
variable "name" {
  description = "Name prefix for resources"
  type        = string
  default     = "formal-data-discovery-satellite"
}

variable "container_cpu" {
  description = "CPU units for the container (1024 = 1 vCPU)"
  type        = number
  default     = 1024
}

variable "container_memory" {
  description = "Memory for the container in MB"
  type        = number
  default     = 2048
}

variable "desired_count" {
  description = "Desired number of tasks to run"
  type        = number
  default     = 1
}

variable "image_tag" {
  description = "Tag of the data discovery satellite image"
  type        = string
  default     = "latest"
}

variable "assign_public_ip" {
  description = "Whether to assign a public IP to the ECS tasks"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# Optional: Use existing IAM roles instead of creating new ones
variable "execution_role_arn" {
  description = "ARN of an existing ECS task execution role. If not provided, a new role will be created."
  type        = string
  default     = null
}

variable "task_role_arn" {
  description = "ARN of an existing ECS task role. If not provided, a new role will be created."
  type        = string
  default     = null
}
