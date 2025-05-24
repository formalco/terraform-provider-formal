# Formal Connector EKS / Kubernetes Setup

This directory contains a demonstration of how to set up a Formal Connector in an EKS cluster, with a Kubernetes cluster as a target resource.

It contains a Terraform configuration to set up the required AWS and Formal resources, and managing the Helm chart deployment.

You can use it as-is to set up a Formal Connector in your EKS cluster, or as a starting point to integrate the Connector in your existing infrastructure.

Formal general purpose documentation is available at [docs.joinformal.com](https://docs.joinformal.com).

## Prerequisites

On your side, you need:
* Helm and Terraform installed
* AWS CLI configured and authenticated. You (or your Terraform runner) need to have write access to both the AWS account and the EKS cluster

On the Formal side, you need:
* A Formal API key, that you can create in the Formal Console
* Access to Formal ECR repository (ask you Formal contact for access).

## Setup

1. Create a `terraform.tfvars` file:
```hcl
# Required variables
region = "your-region"
cluster_name = "your-cluster-name"
formal_api_key = "your-formal-api-key"
formal_org_name = "your-formal-org-name"

# Optional variables
namespace = "your-cluster-namespace"  # default: "default"
helm_values = "your-values.yaml"  # default: "values.yaml"
```

2. Initialize and apply the Terraform configuration:
```bash
terraform init
terraform apply
```

The Connector will be deployed in your EKS cluster and will be ready to use.

## Troubleshooting

If you encounter issues, here are a few things you can check:

* Check the Connector pod status for any issues (e.g. `kubectl describe pod ...`)
* Check the logs of the Connector pod (e.g. `kubectl logs ...`)
* Check Kubernetes events (e.g. `kubectl get events`)

If you still encounter issues, please reach out to us!
