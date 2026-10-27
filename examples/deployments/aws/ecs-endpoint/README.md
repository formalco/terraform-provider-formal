# Formal Endpoint for ECS

Adds the Formal Endpoint as a sidecar to an ECS task, the ECS equivalent of the
[Kubernetes Egress chart](https://github.com/formalco/helm-charts/tree/main/charts/kubernetes-egress).
The Endpoint intercepts the task's outbound traffic and applies your Formal policies.

The module only computes container definitions. It creates no AWS resources.

## Requirements

- **`awsvpc` network mode**, so all containers in the task share one network namespace.
- A way for the Endpoint to sign in to Formal. See [Sign-in](#sign-in).

## Launch types

| | `EC2` (default) | `FARGATE` |
|---|---|---|
| Traffic | The Endpoint intercepts all TCP connections | Applications must proxy HTTP and HTTPS traffic through the Endpoint |
| Required capabilities | `NET_ADMIN`, `NET_BIND_SERVICE` | None |
| Protocols | HTTP, Postgres, MySQL, SSH and others | HTTP, HTTPS |

On Fargate, a client that ignores the proxy settings connects directly. Use `EC2` when Formal must enforce egress.

## Usage

```hcl
module "formal_endpoint" {
  source = "github.com/formalco/terraform-provider-formal//examples/deployments/aws/ecs-endpoint"

  launch_type         = "EC2" # or "FARGATE"
  oidc_integration_id = "integrationoidc_..." # or api_key_secret_arn
  containers = [{
    name      = "app"
    image     = "my-app:1.2.3"
    essential = true
  }]
}

resource "aws_ecs_task_definition" "app" {
  family                   = "app"
  requires_compatibilities = ["EC2"] # or ["FARGATE"], with task-level cpu and memory
  network_mode             = "awsvpc"
  execution_role_arn       = aws_iam_role.execution.arn
  task_role_arn            = aws_iam_role.task.arn # the Endpoint signs in with this role
  container_definitions    = jsonencode(module.formal_endpoint.container_definitions)

  volume {
    name = module.formal_endpoint.volume_name
  }
}
```

Each application container:

- waits for the `formal-endpoint` container to be `HEALTHY` before it starts
- mounts the shared CA volume read-only at `/formal-ca`. Its `ca-bundle.pem` holds the Endpoint image's system CAs plus the Formal CA, so TLS works both to hosts Formal intercepts and to hosts it does not
- gets `SSL_CERT_FILE`, `REQUESTS_CA_BUNDLE`, `NODE_EXTRA_CA_CERTS`, `AWS_CA_BUNDLE` and similar variables pointing at that bundle

Runtimes that ignore these variables, such as Java truststores, need their own trust configuration pointed at `/formal-ca/ca-bundle.pem`.

With `FARGATE`, the module points each application container at the Endpoint
(`127.0.0.1:3128`) by setting the proxy variables that Node.js, Java and most
HTTP clients read. A container must not set `HTTPS_PROXY` or `HTTP_PROXY`
itself.

## Sign-in

1. Turn on outbound identity federation, once per AWS account:

   ```bash
   aws iam enable-outbound-web-identity-federation
   aws iam get-outbound-web-identity-federation-info
   ```

   The second command prints your issuer URL.
2. Allow the task role `sts:GetWebIdentityToken`.
3. In Formal, create an [OIDC integration](https://app.formal.ai/oidc-integrations) for that issuer. Pick the machine user the Endpoint signs in as.
4. Restrict the integration to your task role:

   ```cel
   claims.sub == "arn:aws:iam::<ACCOUNT_ID>:role/<TASK_ROLE>"
   ```

5. Pass the integration ID as `oidc_integration_id`.

To use an API key instead, store it in Secrets Manager or SSM Parameter Store, then pass its ARN as `api_key_secret_arn` in the module. The task execution role must be able to read it. Check with:

```bash
aws iam simulate-principal-policy --policy-source-arn <EXECUTION_ROLE_ARN> \
  --action-names secretsmanager:GetSecretValue --resource-arns <SECRET_ARN> \
  --query 'EvaluationResults[0].EvalDecision'
# Expected: "allowed"
```

## Use your own CA

By default, each task's Endpoint generates its own CA. To sign with a CA from your PKI instead, such as an intermediate your images already trust, store it in Secrets Manager as JSON with `tls.crt` (the CA certificate, then its issuers) and `tls.key`:

```bash
aws secretsmanager create-secret --name formal-ca --secret-string \
  "$(jq -n --rawfile crt ca-chain.pem --rawfile key ca-key.pem '{"tls.crt": $crt, "tls.key": $key}')"
```

Then pass its ARN as `ca_secret_arn`. The task execution role must be able to read it. Only the Endpoint container gets the key. If the pair is not a valid CA, the Endpoint never becomes healthy and the application containers do not start.

## Tests

```bash
terraform init -backend=false && terraform test
```
