resource "aws_lb" "main" {
  name               = var.name
  internal           = false
  load_balancer_type = "network"
  subnets            = aws_subnet.public.*.id

  enable_deletion_protection = true

  tags = {
    Name        = var.name
    Environment = var.environment
  }
}

resource "aws_lb_target_group" "main" {
  name        = var.name
  port        = var.main_port
  protocol    = "TCP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

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
resource "aws_lb_listener" "main" {
  load_balancer_arn = aws_lb.main.id
  port              = var.main_port
  protocol          = "TCP"

  ssl_policy      = null 
  certificate_arn = null 
  alpn_policy     = null

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }

  lifecycle {
    ignore_changes = [default_action]
  }
}