resource "aws_lb" "main" {
  name               = var.name
  internal           = true
  load_balancer_type = "network"
  subnets            = var.private_subnets

  enable_deletion_protection = false

  security_groups = ["${aws_security_group.sg_nlb.id}"]

  tags = {
    Name        = var.name
    Environment = var.environment
  }
}

resource "aws_lb_target_group" "main" {
  name        = var.name
  port        = var.main_port
  protocol    = "TCP"
  vpc_id      = var.vpc_id
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

resource "aws_security_group" "sg_nlb" {
  name        = "${var.name}_nlb_data_classifier"
  description = "Allow traffic for Network Load Balancer."
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

}