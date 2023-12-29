terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>3.0.23"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~>3.0"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

provider "aws" {
  region     = var.region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
}

# Cloud Account Integration Demo (for Managed Cloud) 
# Note the specified aws_cloud_region is the region the CloudFormation stack will be deployed in, which must be deployed with an aws provider setup for eu-west-1, us-east-1, or us-east-2.
resource "formal_cloud_account" "integrated_aws_account" {
  cloud_account_name = var.name
  cloud_provider     = "aws"
  aws_cloud_region   = var.region
}

# Declare the CloudFormation stack
resource "aws_cloudformation_stack" "integrate_with_formal" {
  name = formal_cloud_account.integrated_aws_account.aws_formal_stack_name
  parameters = {
    FormalID          = formal_cloud_account.integrated_aws_account.aws_formal_id
    FormalIamRole     = formal_cloud_account.integrated_aws_account.aws_formal_iam_role
    FormalHandshakeID = formal_cloud_account.integrated_aws_account.aws_formal_handshake_id
    FormalPingbackArn = formal_cloud_account.integrated_aws_account.aws_formal_pingback_arn
  }
  template_body = formal_cloud_account.integrated_aws_account.aws_formal_template_body
  capabilities  = ["CAPABILITY_NAMED_IAM"]
}

resource "aws_db_subnet_group" "main" {
  name       = var.name
  subnet_ids = toset(aws_subnet.private.*.id)
}

resource "formal_dataplane" "main" {
  name               = var.name
  cloud_region       = var.region
  cloud_account_id   = formal_cloud_account.integrated_aws_account.id
  availability_zones = 2

  depends_on = [
    formal_cloud_account.integrated_aws_account,
    aws_cloudformation_stack.integrate_with_formal
  ]
}

resource "aws_security_group" "rds" {
  name   = "rds"
  vpc_id = aws_vpc.main.id

  ingress {
    protocol         = "tcp"
    from_port        = 5432
    to_port          = 5432
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    protocol         = "-1"
    from_port        = 0
    to_port          = 0
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

// AWS RDS Instance
resource "aws_db_instance" "main" {
  identifier             = var.name
  allocated_storage      = 10
  engine                 = "postgres"
  engine_version         = "13.4"
  instance_class         = "db.t3.micro"
  name                   = "main"
  username               = var.postgres_username
  password               = var.postgres_password
  parameter_group_name   = "default.postgres13"
  skip_final_snapshot    = true
  publicly_accessible    = false
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  depends_on = [
    aws_internet_gateway.main
  ]
}

resource "formal_datastore" "main" {
  technology              = "postgres"
  name                    = "${var.name}-postgres"
  hostname                = aws_db_instance.main.address
  port                    = aws_db_instance.main.port
  default_access_behavior = "allow"
}

resource "formal_sidecar" "main" {
  name               = var.name
  deployment_type    = "managed"
  cloud_provider     = "aws"
  cloud_region       = var.region
  cloud_account_id   = formal_cloud_account.integrated_aws_account.id
  fail_open          = false
  dataplane_id       = formal_dataplane.main.id
  global_kms_decrypt = true
  datastore_id       = formal_datastore.main.id
}

# Native Role
resource "formal_native_role" "main_postgres" {
  datastore_id       = formal_datastore.main.id
  native_role_id     = var.postgres_username
  native_role_secret = var.postgres_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}


resource "aws_redshift_subnet_group" "main" {
  name       = "main"
  subnet_ids = toset(aws_subnet.private.*.id)
}

resource "aws_security_group" "allow_ingress_traffic_to_redshift" {
  name        = "allow_ingress_traffic_to_redshift"
  description = "Allow inbound traffic to redshift"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 5439
    to_port     = 5439
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_redshift_cluster" "demo" {
  cluster_identifier        = "tf-redshift-cluster"
  database_name             = "mydb"
  master_username           = var.postgres_username
  master_password           = var.postgres_password
  node_type                 = "dc2.large"
  cluster_type              = "single-node"
  publicly_accessible       = true
  cluster_subnet_group_name = aws_redshift_subnet_group.main.name
  skip_final_snapshot       = true
  vpc_security_group_ids    = [aws_security_group.allow_ingress_traffic_to_redshift.id]
  depends_on = [
    aws_internet_gateway.main
  ]
}

resource "formal_datastore" "main-redshift" {
  technology              = "redshift"
  name                    = "${var.name}-redshift"
  hostname                = aws_redshift_cluster.demo.dns_name
  port                    = aws_redshift_cluster.demo.port
  default_access_behavior = "allow"
}

resource "formal_sidecar" "main-redshift" {
  name               = "${var.name}-redshift"
  deployment_type    = "managed"
  cloud_provider     = "aws"
  cloud_region       = var.region
  cloud_account_id   = formal_cloud_account.integrated_aws_account.id
  fail_open          = false
  dataplane_id       = formal_dataplane.main.id
  global_kms_decrypt = true
  datastore_id       = formal_datastore.main-redshift.id
}

# Native Role
resource "formal_native_role" "main_redshift" {
  datastore_id       = formal_datastore.main-redshift.id
  native_role_id     = var.redshift_username
  native_role_secret = var.redshift_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}

resource "aws_route53_zone_association" "secondary" {
  zone_id = formal_dataplane.main.formal_r53_private_hosted_zone_id
  vpc_id  = aws_vpc.main.id
}

resource "aws_vpc_peering_connection" "main" {
  peer_vpc_id = formal_dataplane.main.formal_vpc_id
  vpc_id      = aws_vpc.main.id
  auto_accept = true
}

resource "aws_route" "vpc_peering_private" {
  count = length(aws_route_table.private)

  route_table_id            = aws_route_table.private[count.index].id
  destination_cidr_block    = formal_dataplane.main.formal_vpc_cidr_block
  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
}

resource "aws_route" "vpc_peering_public" {
  route_table_id            = aws_route_table.public.id
  destination_cidr_block    = formal_dataplane.main.formal_vpc_cidr_block
  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
}

resource "formal_dataplane_routes" "name" {
  destination_cidr_block    = aws_vpc.main.cidr_block
  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  dataplane_id              = formal_dataplane.main.id
}