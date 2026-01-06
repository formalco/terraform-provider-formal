terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.27.0"
    }
  }
}

provider "aws" {
  region = var.region
}

locals {
  image              = "654654333078.dkr.ecr.us-east-1.amazonaws.com/formalco-prod-data-discovery-satellite:${var.image_tag}"
  create_iam_roles   = var.execution_role_arn == null
  execution_role_arn = var.execution_role_arn != null ? var.execution_role_arn : aws_iam_role.ecs_task_execution_role[0].arn
  task_role_arn      = var.task_role_arn != null ? var.task_role_arn : aws_iam_role.ecs_task_role[0].arn
}

# Data source to get current AWS account ID
data "aws_caller_identity" "current" {}

# IAM Role for ECS Task Execution (only created if execution_role_arn not provided)
resource "aws_iam_role" "ecs_task_execution_role" {
  count = local.create_iam_roles ? 1 : 0
  name  = "${var.name}-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  count      = local.create_iam_roles ? 1 : 0
  role       = aws_iam_role.ecs_task_execution_role[0].name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Policy for ECR cross-account access (image is in Formal's account)
resource "aws_iam_role_policy" "ecr_cross_account" {
  count = local.create_iam_roles ? 1 : 0
  name  = "${var.name}-ecr-cross-account"
  role  = aws_iam_role.ecs_task_execution_role[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability"
        ]
        Resource = "arn:aws:ecr:us-east-1:654654333078:repository/formalco-prod-data-discovery-satellite"
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken"
        ]
        Resource = "*"
      }
    ]
  })
}

# Policy for Secrets Manager access
resource "aws_iam_role_policy" "secrets_manager" {
  count = local.create_iam_roles ? 1 : 0
  name  = "${var.name}-secrets-manager"
  role  = aws_iam_role.ecs_task_execution_role[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = aws_secretsmanager_secret.formal_api_key.arn
      }
    ]
  })
}

# IAM Role for ECS Task (only created if task_role_arn not provided)
resource "aws_iam_role" "ecs_task_role" {
  count = local.create_iam_roles ? 1 : 0
  name  = "${var.name}-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# Secrets Manager for Formal API Key
resource "aws_secretsmanager_secret" "formal_api_key" {
  name                    = "${var.name}-api-key"
  recovery_window_in_days = 0

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "formal_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_api_key.id
  secret_string = var.formal_api_key
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "main" {
  name              = "/ecs/${var.name}"
  retention_in_days = 7

  tags = var.tags
}

# ECS Task Definition
resource "aws_ecs_task_definition" "main" {
  family                   = var.name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.container_cpu
  memory                   = var.container_memory
  execution_role_arn       = local.execution_role_arn
  task_role_arn            = local.task_role_arn

  container_definitions = jsonencode([
    {
      name      = var.name
      image     = local.image
      essential = true

      secrets = [
        {
          name      = "FORMAL_CONTROL_PLANE_API_KEY"
          valueFrom = aws_secretsmanager_secret.formal_api_key.arn
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.main.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = var.tags
}

# ECS Service
resource "aws_ecs_service" "main" {
  name            = var.name
  cluster         = var.ecs_cluster_arn
  task_definition = aws_ecs_task_definition.main.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = var.security_group_ids
    assign_public_ip = var.assign_public_ip
  }

  tags = var.tags
}
