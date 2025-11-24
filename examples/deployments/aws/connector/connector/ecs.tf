data "aws_region" "current" {}

resource "aws_cloudwatch_log_group" "connector" {
  name              = "/ecs/${var.name}"
  retention_in_days = 7

  tags = {
    Name        = var.name
    Environment = var.environment
  }
}

resource "aws_ecs_task_definition" "main" {
  family                   = var.name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.container_cpu
  memory                   = var.container_memory
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn
  container_definitions = jsonencode([
    {
      name      = var.name
      image     = var.container_image
      essential = true
      portMappings = concat([
        {
          protocol      = "tcp"
          containerPort = 8080
          hostPort      = 8080
        }
        ], [for port in var.connector_ports : {
          protocol      = "tcp"
          containerPort = port
          hostPort      = port
      }])
      environment = [
        {
          name  = "SERVER_CONNECT_TLS"
          value = "true"
        },
        {
          name  = "CLIENT_LISTEN_TLS"
          value = "true"
        },
      ]
      secrets = [
        {
          name      = "FORMAL_CONTROL_PLANE_API_KEY"
          valueFrom = aws_secretsmanager_secret_version.connector_api_key.arn
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.connector.name
          "awslogs-region"        = data.aws_region.current.id
          "awslogs-stream-prefix" = "ecs"
        }
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

  # Health check port
  ingress {
    description = "Health check port"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Dynamic connector ports
  dynamic "ingress" {
    for_each = toset([for port in var.connector_ports : tostring(port)])
    content {
      description = "Connector port ${ingress.key}"
      from_port   = tonumber(ingress.key)
      to_port     = tonumber(ingress.key)
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  # Allow distributed state port (7946) from within security group
  ingress {
    description = "distributed state port"
    from_port   = 7946
    to_port     = 7946
    protocol    = "tcp"
    self        = true
  }

  # Allow distributed state port (7946 UDP) from within security group
  ingress {
    description = "distributed state port UDP"
    from_port   = 7946
    to_port     = 7946
    protocol    = "udp"
    self        = true
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
  task_definition                    = "${aws_ecs_task_definition.main.family}:${aws_ecs_task_definition.main.revision}"
  desired_count                      = 3
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

  dynamic "load_balancer" {
    for_each = aws_lb_target_group.connector
    content {
      target_group_arn = load_balancer.value.arn
      container_name   = var.name
      container_port   = tonumber(load_balancer.key)
    }
  }

  deployment_controller {
    type = "ECS"
  }

  # desired_count is ignored as it can change due to autoscaling policy  
  lifecycle {
    ignore_changes = [desired_count]
  }
}

resource "aws_appautoscaling_target" "ecs" {
  max_capacity       = 20
  min_capacity       = 1
  resource_id        = "service/${var.ecs_cluster_name}/${aws_ecs_service.main.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "ecs_memory" {
  name               = "memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }

    target_value       = 70
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
  }
}

resource "aws_appautoscaling_policy" "ecs_cpu" {
  name               = "cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }

    target_value       = 60
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
  }
}
