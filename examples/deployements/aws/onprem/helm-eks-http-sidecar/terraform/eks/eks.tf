module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "19.15.3"

  cluster_name    = var.name
  cluster_version = "1.27"

  vpc_id                         = var.vpc_id
  subnet_ids                     = var.private_subnets
  cluster_endpoint_public_access = true

  eks_managed_node_group_defaults = {
    ami_type = "AL2_x86_64"
  }

   cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
  }

  eks_managed_node_groups = {
    one = {
      name = "node-group-1"

      instance_types = ["t3.small"]

      min_size     = 1
      max_size     = 3
      desired_size = 2
    }
  }

  tags = {
    environment = var.environment
  }
}

output "aws_eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "aws_eks_cluster_ca_cert" {
  value = module.eks.cluster_certificate_authority_data
}

output "aws_eks_cluster_name" {
  value = module.eks.cluster_name
}