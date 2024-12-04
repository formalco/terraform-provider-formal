resource "aws_ecs_cluster" "main" {
  name = var.name
  tags = {
    Name        = var.name
    Environment = var.environment
  }
}

