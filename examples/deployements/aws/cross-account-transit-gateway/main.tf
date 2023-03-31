terraform {
  required_version = ">=1.1.8"
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 3.0.15"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

/*
 * AWS ACCOUNT A - Redshift Cluster
 */

provider "aws" {
  region     = var.region
  alias      = "accountA"
  access_key = var.aws_access_key_account_1
  secret_key = var.aws_secret_key_account_1
}

resource "aws_vpc" "vpc_accountA" {
  provider = aws.accountA

  cidr_block = "10.10.0.0/16"
}

resource "aws_subnet" "accountA_priv_subnet_1" {
  provider = aws.accountA

  vpc_id                  = aws_vpc.vpc_accountA.id
  cidr_block              = "10.10.3.0/24"
  map_public_ip_on_launch = false
  availability_zone       = "${var.region}c"
}

resource "aws_subnet" "accountA_priv_subnet_2" {
  provider = aws.accountA

  vpc_id                  = aws_vpc.vpc_accountA.id
  cidr_block              = "10.10.1.0/24"
  map_public_ip_on_launch = false
  availability_zone       = "${var.region}a"
}

resource "aws_ec2_transit_gateway_vpc_attachment" "tgw_attach" {
  provider = aws.accountA

  subnet_ids         = [aws_subnet.accountA_priv_subnet_1.id, aws_subnet.accountA_priv_subnet_2.id]
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  vpc_id             = aws_vpc.vpc_accountA.id
}

resource "aws_route_table" "tgw" {
  provider = aws.accountA

  vpc_id = aws_vpc.vpc_accountA.id
}

resource "aws_route" "vpc1_edge_tgw_access" {
  provider = aws.accountA

  route_table_id         = aws_route_table.tgw.id
  destination_cidr_block = "10.0.0.0/8"
  transit_gateway_id     = aws_ec2_transit_gateway.tgw.id
}

# Route Table Associations
resource "aws_route_table_association" "prv_sub_1a_association" {
  provider = aws.accountA

  subnet_id      = aws_subnet.accountA_priv_subnet_1.id
  route_table_id = aws_route_table.tgw.id
}

resource "aws_route_table_association" "prv_sub_1c_association" {
  provider = aws.accountA

  subnet_id      = aws_subnet.accountA_priv_subnet_2.id
  route_table_id = aws_route_table.tgw.id
}

#Create TGW
resource "aws_ec2_transit_gateway" "tgw" {
  provider = aws.accountA

  description = "tgw_formal"
}

resource "aws_ram_resource_share" "tgw" {
  provider = aws.accountA

  name                      = "tgw"
  allow_external_principals = true
}

resource "aws_ram_principal_association" "tgw" {
  provider = aws.accountA

  principal          = var.aws_account_2_id
  resource_share_arn = aws_ram_resource_share.tgw.arn
}

resource "aws_ram_resource_association" "tgw" {
  provider = aws.accountA

  resource_arn       = aws_ec2_transit_gateway.tgw.arn
  resource_share_arn = aws_ram_resource_share.tgw.arn
}


resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "tgw" {
  provider = aws.accountA

  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.tgw_attach_formal.id
}

resource "aws_redshift_subnet_group" "main" {
  provider = aws.accountA

  name       = "main"
  subnet_ids = [aws_subnet.accountA_priv_subnet_1.id, aws_subnet.accountA_priv_subnet_2.id]
}

resource "aws_security_group" "allow_ingress_traffic_to_redshift" {
  provider = aws.accountA

  name        = "allow_ingress_traffic_to_redshift"
  description = "Allow inbound traffic to redshift"
  vpc_id      = aws_vpc.vpc_accountA.id

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
  provider = aws.accountA

  cluster_identifier        = "tf-redshift-cluster"
  database_name             = "mydb"
  master_username           = var.redshift_username
  master_password           = var.redshift_password
  node_type                 = "dc2.large"
  cluster_type              = "single-node"
  publicly_accessible       = false
  cluster_subnet_group_name = aws_redshift_subnet_group.main.name
  skip_final_snapshot       = true
  vpc_security_group_ids    = [aws_security_group.allow_ingress_traffic_to_redshift.id]
}
/*
 * END OF AWS ACCOUNT A
 */


/*
 * AWS ACCOUNT B - TGW ATTACHMENT AND FORMAL DATAPLANE AND SIDECAR
 */
provider "aws" {
  region     = var.region
  alias      = "accountB"
  access_key = var.aws_access_key_account_2
  secret_key = var.aws_secret_key_account_2
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
  provider = aws.accountB

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


resource "aws_ram_resource_share_accepter" "tgw" {
  provider = aws.accountB

  share_arn = aws_ram_principal_association.tgw.resource_share_arn
}

resource "aws_vpc" "vpc_accountB" {
  provider = aws.accountB

  cidr_block = "172.31.0.0/16"
}

resource "formal_dataplane" "tgw" {
  name               = var.name
  cloud_region       = var.region
  cloud_account_id   = formal_cloud_account.integrated_aws_account.id
  availability_zones = 2

  depends_on = [
    aws_cloudformation_stack.integrate_with_formal
  ]
}

resource "formal_dataplane_routes" "name" {
  destination_cidr_block = "10.0.0.0/8"
  transit_gateway_id     = aws_ec2_transit_gateway.tgw.id
  dataplane_id           = formal_dataplane.tgw.id

  depends_on = [
    aws_ec2_transit_gateway_vpc_attachment_accepter.tgw
  ]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "tgw_attach_formal" {
  provider = aws.accountB

  subnet_ids         = formal_dataplane.tgw.formal_private_subnets
  transit_gateway_id = aws_ec2_transit_gateway.tgw.id
  vpc_id             = formal_dataplane.tgw.formal_vpc_id

  depends_on = [
    formal_dataplane.tgw
  ]
}

resource "formal_datastore" "demo" {
  technology              = "redshift"
  name                    = var.name
  hostname                = aws_redshift_cluster.demo.dns_name
  port                    = aws_redshift_cluster.demo.port
  default_access_behavior = "allow"
}

resource "formal_sidecar" "main-redshift" {
  name               = "${var.name}-sidecar"
  deployment_type    = "managed"
  cloud_provider     = "aws"
  cloud_region       = var.region
  cloud_account_id   = formal_cloud_account.integrated_aws_account.id
  fail_open          = false
  dataplane_id       = formal_dataplane.main.id
  global_kms_decrypt = true
  network_type       = "internet-facing" //internal, internet-and-internal
  datastore_id       = formal_datastore.main-redshift.id
}

resource "formal_native_role" "main_redshift" {
  datastore_id       = formal_datastore.demo.id
  native_role_id     = var.redshift_username
  native_role_secret = var.redshift_password
  use_as_default     = true // per sidecar, exactly one native role must be marked as the default.
}