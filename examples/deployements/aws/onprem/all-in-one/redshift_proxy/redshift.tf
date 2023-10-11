resource "aws_redshift_cluster" "default" {
  cluster_identifier = "tf-redshift-cluster"
  database_name      = "main"
  master_username    = var.redshift_username
  master_password    = var.redshift_password
  node_type          = "dc2.large"
  cluster_type       = "single-node"
  vpc_security_group_ids          = [aws_security_group.default.id]
  cluster_subnet_group_name            = aws_redshift_subnet_group.default.name

}

resource "aws_security_group" "default" {
  vpc_id      = var.vpc_id
  name        = "main-redshift"
  description = "Allow all inbound for Redshift"
  ingress {
    from_port   = 5439
    to_port     = 5439
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_redshift_subnet_group" "default" {
  name       = "main-rd"
  subnet_ids = var.public_subnets

  tags = {
    Name = "Redshift"
  }
}

output "redshift_hostname" {
  value = replace(
    try(aws_redshift_cluster.default.endpoint, ""),
    format(":%s", try(aws_redshift_cluster.default.port, "")),
    "",
  )
}
