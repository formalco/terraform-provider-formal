# Formal Connector on a new AKS cluster

This example creates a one-node AKS cluster in an existing resource group and
deploys a Formal Connector with Helm. The Connector controls access to the AKS
cluster itself.

The cluster uses the Free AKS tier, omits a Linux profile so Terraform does not
configure an SSH key, and enables the AKS OIDC issuer and Azure Workload
Identity.

## Identity model

Terraform creates a user-assigned managed identity for the Connector and links
it to the `formal-connector` Kubernetes service account. The Formal Kubernetes
resource uses an `iam_azure` native user, so the Connector mints AKS tokens from
that workload identity without a client secret.

The `azure-gar-cred` chart uses a separate, existing managed identity that
Formal has authorized to pull images from its Google Artifact Registry. This
configuration reads that identity as a data source and creates only an
AKS-specific federated credential. It does not manage the identity or alter its
existing federated credentials.

## Prerequisites

- Terraform 1.11 or later.
- Azure CLI authentication with permission to create AKS clusters, managed
  identities, and federated identity credentials in the resource group.
- A Formal API key with permission to create Connectors, Resources, and Native
  Users.
- An existing Azure managed identity authorized by Formal for GAR pulls.
- The GCP provider ID and service account email supplied by Formal for that
  GAR identity.

## Configure

Create `terraform.tfvars`:

```hcl
resource_group_name       = "my-existing-resource-group"
cluster_name              = "formal-aks"
formal_api_key            = "<your-formal-api-key>"
gar_pull_identity_name    = "my-formal-gar-pull-identity"
gar_provider_id           = "<provided-by-formal>"
gar_service_account_email = "<provided-by-formal>"
```

Authenticate to Azure with a credential supported by the Azure provider. Keep
`terraform.tfvars` out of version control because it contains the Formal API
key.

## Deploy

```sh
terraform init
terraform apply
```

The apply order is:

1. Create AKS and the Connector managed identity.
2. Configure workload identity and Kubernetes RBAC.
3. Run `azure-gar-cred` to create the short-lived GAR pull secret.
4. Deploy the Connector using that secret.
5. Register the Connector, AKS resource, and managed-identity native user in
   Formal.

Verify the deployments:

```sh
az aks get-credentials \
  --resource-group my-existing-resource-group \
  --name formal-aks
kubectl --namespace formal get pods
kubectl --namespace formal rollout status deployment/formal-connector
kubectl --namespace formal get cronjob/formal-azure-gar-cred-refresh
```
