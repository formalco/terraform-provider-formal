resource "aws_rds_cluster" "default" {
  cluster_identifier              = "aurora-cluster-demo"
  availability_zones              = var.availability_zones
  database_name                   = "main"
  engine                          = "aurora-mysql"
  engine_version                  = "5.7.mysql_aurora.2.07.9"
  master_username                 = var.mysql_username
  master_password                 = var.mysql_password
  db_subnet_group_name            = aws_db_subnet_group.default.name
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.default.name
  vpc_security_group_ids          = [aws_security_group.default.id]
  lifecycle {
    ignore_changes = [
      engine_version,
    ]
  }
}

resource "aws_rds_cluster_instance" "default" {
  count                = 1
  cluster_identifier   = aws_rds_cluster.default.id
  engine               = aws_rds_cluster.default.engine
  engine_version       = aws_rds_cluster.default.engine_version
  instance_class       = "db.r5.large"
  publicly_accessible  = true
  db_subnet_group_name = aws_db_subnet_group.default.name
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
  value = aws_rds_cluster_instance.default[0].endpoint
}