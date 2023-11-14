resource "aws_ecs_task_definition" "ecs_task" {
  family                   = var.name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.container_cpu
  memory                   = var.container_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name  = var.name
      image = var.container_image
      repositoryCredentials = {
        credentialsParameter = var.docker_hub_secret_arn
      }
      essential = true
      portMappings = [
        {
          protocol      = "tcp"
          containerPort = var.main_port
          hostPort      = var.main_port
        },
        {
          protocol      = "tcp"
          containerPort = var.health_check_port
          hostPort      = var.health_check_port
      }]
      environment = [
        {
          name  = "DATA_CLASSIFIER_SATELLITE_URI"
          value = "${var.data_classifier_satellite_url}:${var.data_classifier_satellite_port}"
        },
        {
          name  = "SERVER_CONNECT_TLS"
          value = "true"
        },
        {
          name  = "CLIENT_LISTEN_TLS"
          value = "true"
        },
        {
          name = "LOG_LEVEL",
          value = "debug"
        },
        {
          name  = "DD_VERSION"
          value = "1.0.0"
        },
        {
          name  = "DD_ENV",
          value = "prod"
        },
        {
          name = "ENVIRONMENT",
          value = "prod"
        },
        {
          name  = "DD_SERVICE"
          value = var.name
        },
        {
          name  = "MANAGED_TLS_CERTS"
          value = "true"
        },
        {
          name  = "PII_SAMPLING_RATE"
          value = "8"
        }
      ],
      secrets = [
        {
          name      = "FORMAL_CONTROL_PLANE_API_KEY"
          valueFrom = aws_secretsmanager_secret_version.formal_tls_cert.arn
        },
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
  description = "Allow inbound traffic"
  vpc_id      = var.vpc_id

  ingress {
    description = "Allow inbound traffic"
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

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
  task_definition                    = aws_ecs_task_definition.ecs_task.arn
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

  load_balancer {
    target_group_arn = aws_lb_target_group.main.arn
    container_name   = var.name
    container_port   = var.main_port
  }

  # we ignore task_definition changes as the revision changes on deploy
  # of a new version of the application
  # desired_count is ignored as it can change due to autoscaling policy
  lifecycle {
    ignore_changes = [task_definition, desired_count, load_balancer]
  }
}

resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 20
  min_capacity       = 1
  resource_id        = "service/${var.ecs_cluster_name}/${aws_ecs_service.main.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "ecs_policy_memory" {
  name               = "memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }

    target_value       = 70
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
  }
}

resource "aws_appautoscaling_policy" "ecs_policy_cpu" {
  name               = "cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }

    target_value       = 60
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
  }
}
