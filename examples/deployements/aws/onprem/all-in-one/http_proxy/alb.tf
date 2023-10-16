resource "aws_lb" "main" {
  name               = var.name
  internal           = false
  load_balancer_type = "application"
  subnets            = var.public_subnets

  security_groups = ["${aws_security_group.public_alb.id}"]

  enable_deletion_protection = false

  tags = {
    Name        = var.name
    Environment = var.environment
  }
}

resource "aws_lb_target_group" "main" {
  name        = var.name
  port        = var.main_port
  protocol    = "HTTPS"
  vpc_id      = var.vpc_id
  proxy_protocol_v2 = true
  target_type = "ip"

  health_check {
    healthy_threshold   = "3"
    interval            = "10"
    protocol            = "HTTPS"
    matcher             = "200-399"
    port                = "traffic-port"
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
  protocol          = "HTTPS"
  ssl_policy = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn = var.certificate_arn_acm


  alpn_policy     = null

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }

  lifecycle {
    ignore_changes = [default_action]
  }
}

resource "aws_security_group" "public_alb" {
  name        = "${var.name}_public_alb"
  description = "Allow public traffic for Application Load Balancer."
  vpc_id      = "${var.vpc_id}"

  # for allowing health check traffic
  ingress {
    from_port = 32768 # ephemeral port range: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_PortMapping.html
    # to_port     = 61000
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] // anywhere
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] // anywhere
  }

  ingress {
    # TLS (change to whatever ports you need)
    from_port = 443
    to_port   = 443
    protocol  = "tcp"

    # Please restrict your ingress to only necessary IPs and ports.
    # Opening to 0.0.0.0/0 can lead to security vulnerabilities.
    cidr_blocks = ["0.0.0.0/0"] # add a CIDR block here
  }

  # allow all traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}