resource "aws_lb_target_group" "policy_data_loader" {
  name              = var.name
  port              = var.main_port
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
resource "aws_lb_listener" "policy_data_loader" {
  load_balancer_arn = var.nlb_id
  port              = var.main_port
  protocol          = "TCP"

  ssl_policy      = null
  certificate_arn = null
  alpn_policy     = null

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.policy_data_loader.arn
  }

  lifecycle {
    ignore_changes = [default_action]
  }
}
