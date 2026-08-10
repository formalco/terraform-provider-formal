variable "formal_api_key" {
  description = "Formal Control Plane API Key"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "AWS region for deployment"
  type        = string
}

variable "ecr_region" {
  description = "Region to pull the Formal Connector and Satellite images from (Formal's ECR account 654654333078)."
  type        = string
  default     = "us-east-1"
}

variable "name" {
  description = "Base name for all created resources."
  type        = string
  default     = "formal"
}

# --- Existing infrastructure to reuse (nothing below is created) ---

variable "vpc_id" {
  description = "ID of the existing VPC to deploy into."
  type        = string
}

variable "subnet_ids" {
  description = "Subnets for the Connector and Satellite ECS tasks. Use private subnets with a NAT gateway, or public subnets with assign_public_ip = true."
  type        = list(string)
}

variable "nlb_subnet_ids" {
  description = "Subnets for the Connector's network load balancer. Public subnets for an internet-facing NLB, private subnets when nlb_internal = true."
  type        = list(string)
}

variable "ecs_cluster_arn" {
  description = "ARN of the existing ECS cluster to run both services in."
  type        = string
}

# --- Connector ---

variable "connector_ports" {
  description = "Ports the Connector listens on and exposes through the NLB."
  type        = list(number)
  default     = [443]
}

variable "connector_ingress_cidr_blocks" {
  description = "CIDR blocks allowed to reach the Connector's proxied ports."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "connector_log_level" {
  description = "Connector log level."
  type        = string
  default     = "info"
}

variable "connector_image_tag" {
  description = "Tag of the Connector image to deploy."
  type        = string
  default     = "latest"
}

variable "connector_cpu" {
  description = "Fargate CPU units for the Connector task."
  type        = number
  default     = 1024
}

variable "connector_memory" {
  description = "Fargate memory (MiB) for the Connector task."
  type        = number
  default     = 2048
}

variable "connector_desired_count" {
  description = "Number of Connector tasks to run."
  type        = number
  default     = 3
}

variable "nlb_internal" {
  description = "Whether the Connector NLB is internal (true) or internet-facing (false)."
  type        = bool
  default     = false
}

# --- Satellite ---

variable "satellite_image_tag" {
  description = "Tag of the policy-data-loader Satellite image to deploy."
  type        = string
  default     = "latest"
}

variable "satellite_cpu" {
  description = "Fargate CPU units for the Satellite task."
  type        = number
  default     = 1024
}

variable "satellite_memory" {
  description = "Fargate memory (MiB) for the Satellite task."
  type        = number
  default     = 2048
}

variable "assign_public_ip" {
  description = "Assign public IPs to the ECS task ENIs. Set true only when running tasks in public subnets."
  type        = bool
  default     = false
}

# --- Example policy data loader ---

variable "example_policy_data_loader_schedule" {
  description = "Second-based cron for the example policy data loader. Defaults to daily at 03:00."
  type        = string
  default     = "0 0 3 * * *"
}

variable "tags" {
  description = "Tags applied to created resources."
  type        = map(string)
  default     = {}
}
