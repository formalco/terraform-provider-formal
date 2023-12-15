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

  identifier                      = "mysql-rds"
  allocated_storage               = 20
  db_name                         = "main"
  engine                          = "mysql"
  engine_version                  = "5.7"
  instance_class                  = "db.t3.micro"
  username                        = var.mysql_username
  password                        = var.mysql_password
  publicly_accessible             = true
  parameter_group_name            = "aurora-mysql-custom"
  skip_final_snapshot             = true
  deletion_protection             = false
  backup_retention_period         = 35
  vpc_security_group_ids          = [aws_security_group.default.id]
  storage_encrypted               = false
  db_subnet_group_name            = aws_db_subnet_group.default.name
  performance_insights_enabled    = false
}

resource "aws_rds_cluster_parameter_group" "default" {
  name        = "aurora-mysql-custom"
  family      = "aurora-mysql5.7"
  description = "Managed by Terraform"

  parameter {
    name  = "binlog_cache_size"
    value = "3276800"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "binlog_format"
    value = "ROW"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "ft_min_word_len"
    value = "1"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "event_scheduler"
    value = "ON"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "explicit_defaults_for_timestamp"
    value = "0"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "general_log"
    value = "0"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "group_concat_max_len"
    value = "1000000"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "group_concat_max_len"
    value = "1000000"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "innodb_flush_log_at_trx_commit"
    value = "1"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "innodb_ft_min_token_size"
    value = "2"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "innodb_lock_wait_timeout"
    value = "10"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "log_bin_trust_function_creators"
    value = "1"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "log_output"
    value = "TABLE"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "server_audit_events"
    value = "connect, query"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "slow_query_log"
    value = "1"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "time_zone"
    value = "Europe/Paris"
    apply_method = "pending-reboot"
  }
}

resource "aws_db_subnet_group" "default" {
  name       = "main-mysql"
  subnet_ids = var.public_subnets

  tags = {
    Name = "MYSQL"
  }
}

resource "aws_security_group" "default" {
  vpc_id      = var.vpc_id
  name        = "main-mysql"
  description = "Allow all inbound for Postgres"
  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

output "rds_hostname" {
  value = aws_db_instance.default.address
}