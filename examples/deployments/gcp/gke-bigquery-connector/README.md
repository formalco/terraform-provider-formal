# Formal Connector GKE / BigQuery Setup

This directory contains a demonstration of how to set up a Formal Connector in GKE, connecting to BigQuery with Workload Identity Federation.

It contains:
* A Helm chart that defines the Connector deployment within Kubernetes
* A Terraform configuration to set up the required GCP and Formal resources, and managing the Helm chart deployment

You can use it as-is to set up a Formal Connector in your GKE cluster, or as a starting point to integrate the Connector in your existing deployment pipeline.

Formal general purpose documentation is available at [docs.joinformal.com](https://docs.joinformal.com).


## Prerequisites

On your side, you need:
* Helm, Terraform and `gcloud` CLI installed.
* `gcloud` CLI configured and authenticated. You (or your Terraform runner) need to have write access to both the GCP project and the GKE cluster.
* The last step below will require a Google Workspace admin to manually configure domain-wide delegation.

On the Formal side, you need:
* A Formal API key, that you can create in the Formal Console.
* A Formal ECR repository (ask your Formal contact for access).


## Setup

1. Create a `terraform.tfvars` file:
```hcl
# Required variables
project_id = "your-project-id"
region = "your-region"
cluster_name = "your-cluster-name"
formal_api_key = "your-formal-api-key"
ecr_access_key_id = "your-ecr-access-key-id"
ecr_secret_access_key = "your-ecr-secret-access-key"

# Optional variables
namespace = "your-cluster-namespace"  # default: "default"
helm_values = "my-values.yaml"  # default: "helm/values.yaml"
```

2. Initialize and apply the Terraform configuration:
```bash
terraform init
terraform apply
```

3. Manually configure domain-wide delegation. This step requires Google Workspace admin access.

  * Go to your [Google Workspace Admin Console](https://admin.google.com/)
  * Navigate to **Security > Access and data control > API Controls > Manage domain-wide delegation**
  * Click **Add new**
    * Client ID: Your service account's identifier (from `terraform output google_service_account_id`)
    * OAuth Scopes:
      * https://www.googleapis.com/auth/bigquery
  * Click **Authorize**

You should now be able to connect to BigQuery using the Connector.

For example, using the `bq` CLI:

```bash
CONNECTOR_IP=$(terraform output -raw kubernetes_service_external_ip)
bq query --api http://${CONNECTOR_IP}:7777 'SELECT 1'
```

Or from within your GKE cluster, using the internal hostname:

```bash
CONNECTOR_HOSTNAME=$(terraform output -raw kubernetes_service_internal_hostname)
bq query --api http://${CONNECTOR_HOSTNAME}:7777 'SELECT 1'
```

Or using port forward:

```bash
kubectl port-forward services/formal-connector 7777
bq query --api http://localhost:7777 'SELECT 1'
```

> 💡 **Note:** If you don't want Terraform to manage the Connector deployment in your GKE cluster, you can remove the `helm_release` resource from `main.tf`. You will need to run Terraform first, then Helm configured with the appropriate values.


## Troubleshooting

If you encounter issues, here are a few things you can check:

* Check the Connector pod status for any issues (e.g. `kubectl describe pod ...`)
* Check the logs of the Connector pod (e.g. `kubectl logs ...`)
* Check Kubernetes events (e.g. `kubectl get events`)

If you still encounter issues, please reach out to us!
