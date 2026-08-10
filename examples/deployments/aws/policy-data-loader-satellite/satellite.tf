resource "aws_secretsmanager_secret" "satellite_api_key" {
  name                    = "${var.name}-satellite-api-key"
  recovery_window_in_days = 0
  tags                    = var.tags
}

resource "aws_secretsmanager_secret_version" "satellite_api_key" {
  secret_id     = aws_secretsmanager_secret.satellite_api_key.id
  secret_string = formal_satellite.policy_data_loader.api_key
}

resource "aws_security_group" "satellite" {
  name_prefix = "${var.name}-satellite-"
  vpc_id      = var.vpc_id

  ingress {
    description     = "Policy data loader gRPC from the Connector"
    from_port       = 50056
    to_port         = 50056
    protocol        = "tcp"
    security_groups = [aws_security_group.connector.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

# The Connector dials the Satellite at <hostname>:50056 and verifies its TLS
# cert against that hostname. Service discovery makes the name resolve to the
# Satellite in-VPC; registering it as the satellite hostname makes the control
# plane issue a cert whose SAN matches.
resource "aws_service_discovery_private_dns_namespace" "satellite" {
  name = "${var.name}.internal"
  vpc  = var.vpc_id
  tags = var.tags
}

resource "aws_service_discovery_service" "satellite" {
  name = "policy-data-loader"

  dns_config {
    namespace_id   = aws_service_discovery_private_dns_namespace.satellite.id
    routing_policy = "MULTIVALUE"

    dns_records {
      type = "A"
      ttl  = 10
    }
  }

  tags = var.tags
}

resource "formal_satellite_hostname" "satellite" {
  satellite_id = formal_satellite.policy_data_loader.id
  hostname     = "${aws_service_discovery_service.satellite.name}.${aws_service_discovery_private_dns_namespace.satellite.name}"
}

resource "aws_cloudwatch_log_group" "satellite" {
  name              = "/ecs/${var.name}-satellite"
  retention_in_days = 7
  tags              = var.tags
}

resource "aws_ecs_task_definition" "satellite" {
  family                   = "${var.name}-satellite"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.satellite_cpu
  memory                   = var.satellite_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.satellite_task_role.arn

  container_definitions = jsonencode([
    {
      name      = "${var.name}-satellite"
      image     = local.satellite_image
      essential = true

      secrets = [
        {
          name      = "FORMAL_CONTROL_PLANE_API_KEY"
          valueFrom = aws_secretsmanager_secret.satellite_api_key.arn
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.satellite.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = var.tags
}

resource "aws_ecs_service" "satellite" {
  name             = "${var.name}-satellite"
  cluster          = var.ecs_cluster_arn
  task_definition  = aws_ecs_task_definition.satellite.arn
  desired_count    = 1
  launch_type      = "FARGATE"
  platform_version = "1.4.0"

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = [aws_security_group.satellite.id]
    assign_public_ip = var.assign_public_ip
  }

  service_registries {
    registry_arn = aws_service_discovery_service.satellite.arn
  }

  # Satellite fetches its TLS cert once at startup, so register the hostname first.
  depends_on = [formal_satellite_hostname.satellite]

  tags = var.tags
}
