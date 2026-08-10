resource "aws_secretsmanager_secret" "connector_api_key" {
  name = "${var.name}-connector-api-key"
  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "connector_api_key" {
  secret_id     = aws_secretsmanager_secret.connector_api_key.id
  secret_string = formal_connector.main.api_key
}

resource "aws_security_group" "connector" {
  name        = "${var.name}-connector"
  description = "Allow inbound traffic to the Connector"
  vpc_id      = var.vpc_id

  ingress {
    description = "Health check port"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  dynamic "ingress" {
    for_each = toset([for port in var.connector_ports : tostring(port)])
    content {
      description = "Connector port ${ingress.key}"
      from_port   = tonumber(ingress.key)
      to_port     = tonumber(ingress.key)
      protocol    = "tcp"
      cidr_blocks = var.connector_ingress_cidr_blocks
    }
  }

  ingress {
    description = "distributed state port"
    from_port   = 7946
    to_port     = 7946
    protocol    = "tcp"
    self        = true
  }

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

  tags = var.tags
}

resource "aws_lb" "connector" {
  name                       = substr(var.name, 0, 32)
  internal                   = var.nlb_internal
  load_balancer_type         = "network"
  subnets                    = var.nlb_subnet_ids
  enable_deletion_protection = false
  tags                       = var.tags
}

resource "aws_lb_target_group" "connector" {
  for_each = toset([for port in var.connector_ports : tostring(port)])

  name        = substr("${var.name}-${each.key}", 0, 32)
  port        = tonumber(each.key)
  protocol    = "TCP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "8080"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = var.tags
}

resource "aws_lb_listener" "connector" {
  for_each = toset([for port in var.connector_ports : tostring(port)])

  load_balancer_arn = aws_lb.connector.arn
  port              = each.key
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.connector[each.key].arn
  }
}

resource "aws_cloudwatch_log_group" "connector" {
  name              = "/ecs/${var.name}-connector"
  retention_in_days = 7
  tags              = var.tags
}

resource "aws_ecs_task_definition" "connector" {
  family                   = "${var.name}-connector"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.connector_cpu
  memory                   = var.connector_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.connector_task_role.arn

  container_definitions = jsonencode([
    {
      name      = "${var.name}-connector"
      image     = local.connector_image
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
      secrets = [
        {
          name      = "FORMAL_CONTROL_PLANE_API_KEY"
          valueFrom = aws_secretsmanager_secret_version.connector_api_key.arn
        }
      ]
      # The Connector reads the satellite hostname once at startup. Baking it in
      # here (the Connector does not consume this variable) means any hostname
      # change produces a new task revision, restarting the Connector so it
      # re-reads the satellite.
      environment = [
        {
          name  = "POLICY_DATA_LOADER_SATELLITE_HOSTNAME"
          value = formal_satellite_hostname.satellite.hostname
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.connector.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = var.tags
}

resource "aws_ecs_service" "connector" {
  name                               = "${var.name}-connector"
  cluster                            = var.ecs_cluster_arn
  task_definition                    = "${aws_ecs_task_definition.connector.family}:${aws_ecs_task_definition.connector.revision}"
  desired_count                      = var.connector_desired_count
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent         = 200
  health_check_grace_period_seconds  = 60
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  platform_version                   = "1.4.0"

  network_configuration {
    security_groups  = [aws_security_group.connector.id]
    subnets          = var.subnet_ids
    assign_public_ip = var.assign_public_ip
  }

  dynamic "load_balancer" {
    for_each = aws_lb_target_group.connector
    content {
      target_group_arn = load_balancer.value.arn
      container_name   = "${var.name}-connector"
      container_port   = tonumber(load_balancer.key)
    }
  }

  deployment_controller {
    type = "ECS"
  }

  depends_on = [aws_lb_listener.connector]

  lifecycle {
    ignore_changes = [desired_count]
  }
}

resource "aws_appautoscaling_target" "connector" {
  max_capacity       = 20
  min_capacity       = 1
  resource_id        = "service/${local.ecs_cluster_name}/${aws_ecs_service.connector.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "connector_memory" {
  name               = "${var.name}-connector-memory"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.connector.resource_id
  scalable_dimension = aws_appautoscaling_target.connector.scalable_dimension
  service_namespace  = aws_appautoscaling_target.connector.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = 70
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
  }
}

resource "aws_appautoscaling_policy" "connector_cpu" {
  name               = "${var.name}-connector-cpu"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.connector.resource_id
  scalable_dimension = aws_appautoscaling_target.connector.scalable_dimension
  service_namespace  = aws_appautoscaling_target.connector.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 60
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
  }
}
