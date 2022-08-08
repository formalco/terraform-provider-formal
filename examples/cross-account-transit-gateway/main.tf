terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~>1.0.33"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~>3.0"
    }
  }
}

provider "formal" {
  client_id = var.formal_client_id
  secret_key = var.formal_secret_key
}

provider "aws" {
  region  = "${var.region}"
  alias   = "account1"
  access_key = var.aws_access_key_account_1
  secret_key = var.aws_secret_key_account_1
}

provider "aws" {
  region     = "${var.region}"
  alias      = "account2"
  access_key = var.aws_access_key_account_2
  secret_key = var.aws_secret_key_account_2
}

resource "aws_vpc" "vpc_account_1" {
  cidr_block = "10.10.0.0/16"
  provider = aws.account1
}

resource "aws_subnet" "vpc_account_1_private_subnet" {
  vpc_id                  = aws_vpc.vpc_account_1.id
  cidr_block              = "10.10.3.0/24"
  map_public_ip_on_launch = false
  availability_zone       = "${var.region}c"

  provider = aws.account1
}

resource "aws_subnet" "vpc_account_1_private_subnet_2" {
  vpc_id                  = aws_vpc.vpc_account_1.id
  cidr_block              = "10.10.1.0/24"
  map_public_ip_on_launch = false
  availability_zone       = "${var.region}a"

  provider = aws.account1
}

resource "aws_ec2_transit_gateway_vpc_attachment" "tgw_attach" {
  subnet_ids         = [aws_subnet.vpc_account_1_private_subnet.id, aws_subnet.vpc_account_1_private_subnet_2.id]
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  vpc_id             = aws_vpc.vpc_account_1.id

  provider = aws.account1
}

resource "aws_route_table" "tgw" {
  vpc_id = aws_vpc.vpc_account_1.id

  provider = aws.account1
}

// try to delete it
resource "aws_route" "spoke1_internet_access" {
  route_table_id         = aws_route_table.tgw.id
  destination_cidr_block = "0.0.0.0/0"
  transit_gateway_id     = aws_ec2_transit_gateway.tgw.id

  provider = aws.account1
}

resource "aws_route" "vpc1_edge_tgw_access" {
  route_table_id         = aws_route_table.tgw.id
  destination_cidr_block = "10.0.0.0/8"
  transit_gateway_id     = aws_ec2_transit_gateway.tgw.id

  provider = aws.account1
  }

# Route Table Associations
resource "aws_route_table_association" "spoke_1_prv_sub_1a_association" {
  subnet_id      = aws_subnet.vpc_account_1_private_subnet.id
  route_table_id = aws_route_table.tgw.id

  provider = aws.account1
}

resource "aws_route_table_association" "spoke_1_prv_sub_1c_association" {
  subnet_id      = aws_subnet.vpc_account_1_private_subnet_2.id
  route_table_id = aws_route_table.tgw.id

  provider = aws.account1
}


#Create TGW
resource "aws_ec2_transit_gateway" "tgw" {
  description = "tgw_formal"

  provider = aws.account1
}

resource "aws_ram_resource_share" "tgw" {
  name                      = "tgw"
  allow_external_principals = true

  provider = aws.account1
}

resource "aws_ram_principal_association" "tgw" {
  principal          = var.aws_account_2_id
  resource_share_arn = aws_ram_resource_share.tgw.arn

  provider = aws.account1
}

resource "aws_ram_resource_association" "tgw" {
  resource_arn       = aws_ec2_transit_gateway.tgw.arn
  resource_share_arn = aws_ram_resource_share.tgw.arn

  provider = aws.account1
}

resource "aws_ram_resource_share_accepter" "tgw" {
  share_arn = aws_ram_principal_association.tgw.resource_share_arn

  provider = aws.account2
}

resource "formal_dataplane" "tgw" {
  name               = var.name
  cloud_region       = var.region
  cloud_account_id   = var.cloud_account_id // should be optional
  customer_vpc_id    = var.customer_vpc_id
  availability_zones = 2
}

resource "formal_dataplane_routes" "name" {
  destination_cidr_block = "10.0.0.0/8"
  transit_gateway_id     = aws_ec2_transit_gateway.tgw.id
  dataplane_id           = formal_dataplane.tgw.id
}

resource "aws_ec2_transit_gateway_vpc_attachment" "tgw_attach_formal" {
  subnet_ids         = formal_dataplane.tgw.formal_private_subnets
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  vpc_id             = formal_dataplane.tgw.formal_vpc_id

  provider = aws.account2

  depends_on = [
    formal_dataplane.tgw
  ]
}

resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "tgw" {
  provider = aws.account1

  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.tgw_attach_formal.id
}

resource "aws_db_subnet_group" "default" {
  provider = aws.account1

  name       = "main"
  subnet_ids = [aws_subnet.vpc_account_1_private_subnet.id, aws_subnet.vpc_account_1_private_subnet_2.id]
}

resource "aws_db_instance" "demo" {
  provider             = aws.account1
  identifier           = var.name
  allocated_storage    = 10
  engine               = "postgres"
  engine_version       = "13.3"
  instance_class       = "db.t3.micro"
  name                 = "main"
  username             = var.postgres_username
  password             = var.postgres_password
  parameter_group_name = "default.postgres13"
  skip_final_snapshot  = true
  publicly_accessible  = false
  db_subnet_group_name = aws_db_subnet_group.default.name
}

resource "formal_datastore" "demo" {
  technology       = "postgres" # postgres, redshift, snowflake
  name             = var.name
  hostname         = aws_db_instance.demo.address
  port             = aws_db_instance.demo.port
  deployment_type  = "managed"
  cloud_provider   = "aws"
  cloud_region     = var.region
  cloud_account_id = var.cloud_account_id
  customer_vpc_id  = var.customer_vpc_id
  fail_open        = false
  username         = var.postgres_username
  password         = var.postgres_password
  dataplane_id     = formal_dataplane.tgw.id
  global_kms_decrypt = true
}
