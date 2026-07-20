# Formal Connector GKE / BigQuery Setup

This directory contains a demonstration of how to set up a Formal Connector in GKE, connecting to BigQuery with Workload Identity Federation.

It contains:
* A Terraform configuration to set up the required GCP and Formal resources
* Deployment of the official Formal Helm chart (`formal/connector`) with GCP-specific service and Workload Identity annotations
* The Connector image pulled from Formal's public GCP Artifact Registry (`us-docker.pkg.dev/formal-public-assets/...`)

You can use it as-is to set up a Formal Connector in your GKE cluster, or as a starting point to integrate the Connector in your existing deployment pipeline.

Formal general purpose documentation is available at [docs.joinformal.com](https://docs.joinformal.com).
Helm charts are available at [github.com/formalco/helm-charts](https://github.com/formalco/helm-charts).


## Prerequisites

On your side, you need:
* Helm, Terraform and `gcloud` CLI installed.
* `gcloud` CLI configured and authenticated. You (or your Terraform runner) need to have write access to both the GCP project and the GKE cluster.
* A GKE cluster with Workload Identity enabled.
* The last step below will require a Google Workspace admin to manually configure domain-wide delegation.

On the Formal side, you need:
* A Formal API key, that you can create in the Formal Console.


## Setup

1. Create a `terraform.tfvars` file:
```hcl
# Required variables
project_id     = "your-project-id"
region         = "your-region"
cluster_name   = "your-cluster-name"
formal_api_key = "your-formal-api-key"

# Optional variables
namespace = "default"
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

By default, the Connector is deployed with an internal load balancer, which is only accessible from within the VPC. You can connect from within your GKE cluster using the internal hostname:

```bash
CONNECTOR_HOSTNAME=$(terraform output -raw kubernetes_service_internal_hostname)
bq query --api http://${CONNECTOR_HOSTNAME}:7777 'SELECT 1'
```

For local development or testing, you can use port forwarding:

```bash
kubectl port-forward services/formal-connector 7777
bq query --api http://localhost:7777 'SELECT 1'
```

If you need external access from outside the VPC, you can modify the service configuration in `main.tf` to use an external load balancer by removing the `cloud.google.com/load-balancer-type` annotation.

> **Note:** If you don't want Terraform to manage the Connector deployment in your GKE cluster, you can remove the `helm_release` resource from `main.tf`. You will need to run Terraform first, then Helm configured with the appropriate values.


## Customization

The connector is deployed from the public Formal Helm repository (`https://formalco.github.io/helm-charts`). GCP-specific values are set in `main.tf`:

* `image.repository` set to Formal's public GCP Artifact Registry image
* `serviceAccount.annotations["iam.gke.io/gcp-service-account"]` for Workload Identity (BigQuery access at runtime)
* `service.annotations["cloud.google.com/load-balancer-type"] = "Internal"` for an internal load balancer

To inspect the full set of chart defaults:

```bash
helm repo add formal https://formalco.github.io/helm-charts
helm show values formal/connector
```


## Troubleshooting

If you encounter issues, here are a few things you can check:

* Check the Connector pod status for any issues (e.g. `kubectl describe pod ...`)
* Check the logs of the Connector pod (e.g. `kubectl logs ...`)
* Check Kubernetes events (e.g. `kubectl get events`)
* Verify the Workload Identity binding is correctly configured

If you still encounter issues, please reach out to us!
