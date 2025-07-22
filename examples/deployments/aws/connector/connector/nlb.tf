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



