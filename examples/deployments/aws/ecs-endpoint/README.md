# Formal Endpoint for ECS on EC2

Adds the Formal Endpoint as a sidecar to an ECS task, the ECS equivalent of the
[Kubernetes Egress chart](https://github.com/formalco/helm-charts/tree/main/charts/kubernetes-egress).
The Endpoint intercepts the task's outbound traffic and applies your Formal policies.

The module only computes container definitions. It creates no AWS resources.

## Requirements

- **EC2 launch type.** Fargate cannot add `NET_ADMIN`, which the Endpoint needs to intercept traffic.
- **`awsvpc` network mode**, so all containers in the task share one network namespace.
- A Formal machine user API key in Secrets Manager or SSM Parameter Store, readable by the task execution role.

## Usage

```hcl
module "formal_endpoint" {
  source = "github.com/formalco/terraform-provider-formal//examples/deployments/aws/ecs-endpoint"

  api_key_secret_arn = aws_secretsmanager_secret.formal_api_key.arn
  containers = [{
    name      = "app"
    image     = "my-app:1.2.3"
    essential = true
  }]
}

resource "aws_ecs_task_definition" "app" {
  family                   = "app"
  requires_compatibilities = ["EC2"]
  network_mode             = "awsvpc"
  execution_role_arn       = aws_iam_role.execution.arn
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

## Tests

```bash
terraform init -backend=false && terraform test
```
