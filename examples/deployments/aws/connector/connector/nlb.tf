resource "aws_lb" "main" {
  name               = var.name
  internal           = false
  load_balancer_type = "network"
  subnets            = var.public_subnets

  enable_deletion_protection = false

  tags = {
    Name        = var.name
    Environment = var.environment
  }
}

resource "aws_lb_target_group" "connector_postgres" {
  name              = "${var.name}-postgres"
  port              = var.connector_postgres_listener_port
  protocol          = "TCP"
  vpc_id            = var.vpc_id
  proxy_protocol_v2 = false
  target_type       = "ip"

  health_check {
    healthy_threshold   = "3"
    interval            = "10"
    protocol            = "HTTP"
    matcher             = "200-399"
    port                = var.health_check_port
    timeout             = "6"
    path                = "/health"
    unhealthy_threshold = "3"
  }

  tags = {
    Name        = var.name
    Environment = var.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Redirect traffic to target group
resource "aws_lb_listener" "connector_postgres" {
  load_balancer_arn = aws_lb.main.id
  port              = var.connector_postgres_listener_port
  protocol          = "TCP"

  ssl_policy      = null
  certificate_arn = null
  alpn_policy     = null

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.connector_postgres.arn
  }

  lifecycle {
    ignore_changes = [default_action]
  }
}

resource "aws_lb_target_group" "connector_mysql" {
  name              = "${var.name}-mysql"
  port              = var.connector_mysql_port
  protocol          = "TCP"
  vpc_id            = var.vpc_id
  proxy_protocol_v2 = false
  target_type       = "ip"

  health_check {
    healthy_threshold   = "3"
    interval            = "10"
    protocol            = "HTTP"
    matcher             = "200-399"
    port                = var.health_check_port
    timeout             = "6"
    path                = "/health"
    unhealthy_threshold = "3"
  }

  tags = {
    Name        = var.name
    Environment = var.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Redirect traffic to target group
resource "aws_lb_listener" "connector_mysql" {
  load_balancer_arn = aws_lb.main.id
  port              = var.connector_mysql_port
  protocol          = "TCP"

  ssl_policy      = null
  certificate_arn = null
  alpn_policy     = null

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.connector_mysql.arn
  }

  lifecycle {
    ignore_changes = [default_action]
  }
}

resource "aws_lb_target_group" "connector_kubernetes" {
  name              = "${var.name}-kubernetes"
  port              = var.connector_kubernetes_listener_port
  protocol          = "TCP"
  vpc_id            = var.vpc_id
  proxy_protocol_v2 = false
  target_type       = "ip"

  health_check {
    healthy_threshold   = "3"
    interval            = "10"
    protocol            = "HTTP"
    matcher             = "200-399"
    port                = var.health_check_port
    timeout             = "6"
    path                = "/health"
    unhealthy_threshold = "3"
  }

  tags = {
    Name        = var.name
    Environment = var.environment
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Redirect traffic to target group
resource "aws_lb_listener" "connector_kubernetes" {
  load_balancer_arn = aws_lb.main.id
  port              = var.connector_kubernetes_listener_port
  protocol          = "TCP"

  ssl_policy      = null
  certificate_arn = null
  alpn_policy     = null

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.connector_kubernetes.arn
  }

  lifecycle {
    ignore_changes = [default_action]
  }
}