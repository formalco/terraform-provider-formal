resource "aws_db_instance" "default" {
  lifecycle {
    ignore_changes = [
      allocated_storage,
      engine_version,
      instance_class,
      username,
      password,
      multi_az,
      publicly_accessible,
      parameter_group_name,
      skip_final_snapshot,
      deletion_protection,
      backup_retention_period,
      vpc_security_group_ids,
      storage_encrypted,
      db_subnet_group_name,
      enabled_cloudwatch_logs_exports,
      performance_insights_enabled,
    ]
  }

  identifier                      = "postgres-rds"
  allocated_storage               = 20
  db_name                         = "main"
  engine                          = "postgres"
  engine_version                  = "14.7"
  instance_class                  = "db.t3.micro"
  username                        = var.postgres_username
  password                        = var.postgres_password
  multi_az                        = true
  publicly_accessible             = false
  parameter_group_name            = "default.postgres14"
  skip_final_snapshot             = false
  deletion_protection             = true
  backup_retention_period         = 35
  vpc_security_group_ids          = [aws_security_group.default.id]
  storage_encrypted               = true
  db_subnet_group_name            = aws_db_subnet_group.default.name
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  performance_insights_enabled    = true
}

resource "aws_db_subnet_group" "default" {
  name       = "main"
  subnet_ids = var.private_subnets

  tags = {
    Name = "RDS"
  }
}

resource "aws_security_group" "default" {
  vpc_id      = var.vpc_id
  name        = "main-rds"
  description = "Allow all inbound for Postgres"
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

output "rds_endpoint" {
  value = aws_db_instance.default.endpoint
}