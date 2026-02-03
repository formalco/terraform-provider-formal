# Formal Connector GKE / Cloud SQL Setup

This directory contains a demonstration of how to set up a Formal Connector in GKE, connecting to Cloud SQL for PostgreSQL using the Cloud SQL Auth Proxy with IAM authentication.

It contains:
* A Terraform configuration to set up the required GCP and Formal resources
* Deployment of the official Formal Helm charts (`formal/connector` and `formal/ecr-cred`)
* Cloud SQL Proxy configured as a sidecar container for secure connectivity
* GCP IAM authentication handled by the Formal Connector

You can use it as-is to set up a Formal Connector in your GKE cluster, or as a starting point to integrate the Connector in your existing deployment pipeline.

Formal general purpose documentation is available at [docs.joinformal.com](https://docs.joinformal.com).


## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  GKE Pod                                                        │
│  ┌─────────────────────┐      ┌─────────────────────────────┐  │
│  │  Formal Connector   │      │  Cloud SQL Auth Proxy       │  │
│  │                     │      │                             │  │
│  │  Listens on :5432   │─────▶│  Connects to Cloud SQL via  │  │
│  │  (client-facing)    │      │  localhost:5433             │  │
│  │  (IAM auth via WI)  │      │                             │  │
│  └─────────────────────┘      └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                            │
                                            ▼
                                  ┌─────────────────────┐
                                  │  Cloud SQL          │
                                  │  (PostgreSQL)       │
                                  └─────────────────────┘
```

The Cloud SQL Proxy runs as a sidecar and handles secure connectivity to Cloud SQL.
The Formal Connector handles GCP IAM authentication using Workload Identity to generate IAM tokens.


## Prerequisites

On your side, you need:
* Helm, Terraform and `gcloud` CLI installed.
* `gcloud` CLI configured and authenticated. You (or your Terraform runner) need to have write access to both the GCP project and the GKE cluster.
* Cloud SQL Admin API enabled (`gcloud services enable sqladmin.googleapis.com`).
* A GKE cluster with Workload Identity enabled.
* A Cloud SQL for PostgreSQL instance with IAM database authentication enabled.

On the Formal side, you need:
* A Formal API key, that you can create in the Formal Console.
* Formal ECR credentials (ask your Formal contact for access).


## Setup

1. Create a `terraform.tfvars` file:
```hcl
# Required variables
project_id                    = "your-project-id"
region                        = "your-region"
cluster_name                  = "your-cluster-name"
cloud_sql_instance_connection = "your-project:your-region:your-instance"
formal_api_key                = "your-formal-api-key"
ecr_access_key_id             = "your-ecr-access-key-id"
ecr_secret_access_key         = "your-ecr-secret-access-key"

# Optional variables
namespace      = "default"
connector_name = "cloudsql-connector"
postgres_port  = 5432
```

2. Initialize and apply the Terraform configuration:
```bash
terraform init
terraform apply
```

The Terraform configuration automatically creates:
* A GCP service account with the necessary IAM roles
* Workload Identity binding for the Kubernetes service account
* The Cloud SQL IAM database user

You should now be able to connect to Cloud SQL via the Connector.

By default, the Connector is deployed with an internal load balancer, which is only accessible from within the VPC. You can connect from within your GKE cluster using the internal hostname:

```bash
CONNECTOR_HOSTNAME=$(terraform output -raw kubernetes_service_internal_hostname)
psql "host=${CONNECTOR_HOSTNAME} port=5432 user=<formal-user> dbname=<database>@cloudsql-connector-postgres"
```

For local development or testing, you can use port forwarding:

```bash
kubectl port-forward services/formal-connector 5432
psql "host=localhost port=5432 user=<formal-user> dbname=<database>@cloudsql-connector-postgres"
```

If you need external access from outside the VPC, you can modify the service configuration in `main.tf` to use an external load balancer by removing the `cloud.google.com/load-balancer-type` annotation.

> **Note:** If you don't want Terraform to manage the Connector deployment in your GKE cluster, you can remove the `helm_release` resources from `main.tf`. You will need to run Terraform first, then Helm configured with the appropriate values.


## Customization

### Using Private IP

By default, this example uses Cloud SQL's public IP. If your Cloud SQL instance has a private IP and your GKE cluster can reach it (e.g., same VPC or VPC peering), add the `--private-ip` flag to the `sidecars` configuration in `main.tf`:

```hcl
args = [
  var.cloud_sql_instance_connection,
  "--port=5433",
  "--private-ip"
]
```

### Adjusting Resources

You can adjust the Connector and Cloud SQL Proxy resources by modifying the `values` block in `main.tf`.


## Troubleshooting

If you encounter issues, here are a few things you can check:

* Check the Connector pod status for any issues (e.g. `kubectl describe pod ...`)
* Check the logs of both containers:
  * Connector logs: `kubectl logs <pod-name> -c connector`
  * Cloud SQL Proxy logs: `kubectl logs <pod-name> -c cloud-sql-proxy`
* Check Kubernetes events (e.g. `kubectl get events`)
* Verify the GCP service account has the correct IAM roles:
  * `roles/cloudsql.client`
  * `roles/cloudsql.instanceUser`
* Verify the Workload Identity binding is correctly configured

If you still encounter issues, please reach out to us!
