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

# Target groups for connector ports
resource "aws_lb_target_group" "connector" {
  for_each = toset([for port in var.connector_ports : tostring(port)])

  name        = "${var.name}-${each.key}"
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

  tags = {
    Name        = "${var.name}-${each.key}"
    Environment = var.environment
  }
}

# Listeners for connector ports
resource "aws_lb_listener" "connector" {
  for_each = toset([for port in var.connector_ports : tostring(port)])

  load_balancer_arn = aws_lb.main.arn
  port              = each.key
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.connector[each.key].arn
  }
}
