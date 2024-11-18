resource "aws_ecs_task_definition" "main" {
  family                   = var.name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 2048
  memory                   = 4096
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = var.name
      image     = var.container_image
      essential = true
      portMappings = [
        {
          protocol      = "tcp"
          containerPort = var.health_check_port
          hostPort      = var.health_check_port
      }]
      secrets = [
        {
          name      = "FORMAL_CONTROL_PLANE_API_KEY"
          valueFrom = aws_secretsmanager_secret_version.satellite_data_classifier_api_key.arn
        }
      ],
      logConfiguration = {
        logDriver = "awsfirelens"
        options = {
          "Name"       = "datadog",
          "Host"       = "http-intake.logs.datadoghq.eu",
          "TLS"        = "on",
          "dd_source"  = var.name,
          "provider"   = "ecs",
          "dd_service" = var.name,
          "apikey"     = var.datadog_api_key
        }
      }
      dependsOn = [
        { "containerName" : "log_router", "condition" : "START" },
        { "condition" = "HEALTHY", "containerName" = "datadog-agent" }
      ]
    },
    {
      name              = "log_router"
      image             = "public.ecr.aws/aws-observability/aws-for-fluent-bit:stable"
      memoryReservation = 50,
      firelensConfiguration = {
        "type" = "fluentbit",
        "options" = {
          "enable-ecs-log-metadata" = "true"
        }
      },
    },
    {
      name  = "datadog-agent",
      image = "public.ecr.aws/datadog/agent:latest",
      portMappings = [
        {
          "containerPort" = 8126,
          "hostPort"      = 8126,
          "protocol"      = "tcp"
        }
      ],
      environment = [{
        "name"  = "ECS_FARGATE",
        "value" = "true"
        },
        {
          "name"  = "DD_APM_ENABLED",
          "value" = "true"
        },
        {
          "name"  = "DD_LOGS_ENABLED",
          "value" = "true"
        },
        {
          "name"  = "DD_LOGS_CONFIG_CONTAINER_COLLECT_ALL",
          "value" = "true"
        },
        {
          "name"  = "DD_APM_NON_LOCAL_TRAFFIC",
          "value" = "true"
        },
        {
          "name"  = "DD_API_KEY",
          "value" = var.datadog_api_key
        },
        {
          "name"  = "DD_SITE",
          "value" = "datadoghq.eu"
      }],
      healthCheck = {
        "command" = [
          "CMD-SHELL",
          "agent health"
        ],
        "interval" = 30,
        "timeout"  = 5,
        "retries"  = 3
      }
    }
  ])

  tags = {
    Name        = var.name
    Environment = var.environment
  }
}

resource "aws_security_group" "main" {
  name        = var.name
  description = "Allow outbound traffic"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

}

resource "aws_ecs_service" "main" {
  name                               = var.name
  cluster                            = var.ecs_cluster_id
  task_definition                    = aws_ecs_task_definition.main.arn
  desired_count                      = 1
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent         = 200
  health_check_grace_period_seconds  = 60
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  platform_version                   = "1.4.0"

  network_configuration {
    security_groups  = [aws_security_group.main.id]
    subnets          = var.private_subnets
    assign_public_ip = false
  }

  deployment_controller {
    type = "ECS"
  }

  # we ignore task_definition changes as the revision changes on deploy
  # of a new version of the application
  # desired_count is ignored as it can change due to autoscaling policy
  lifecycle {
    ignore_changes = [task_definition, desired_count]
  }
}
