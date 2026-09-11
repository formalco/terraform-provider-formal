resource "formal_resource" "db" {
  name       = "analytics-postgres"
  hostname   = "analytics.internal"
  technology = "postgres"
  port       = 5432

  native_users_v3_enabled = true
}

# Username and password, read from an environment variable on the connector.
resource "formal_native_user_v3" "basic" {
  resource_id = formal_resource.db.id
  label       = "admin"

  basic {
    username = "app_admin"
    password {
      environment_variable = "ANALYTICS_ADMIN_PASSWORD"
    }
  }
}

# AWS IAM, inheriting the role from the connector's environment.
resource "formal_native_user_v3" "aws_iam" {
  resource_id = formal_resource.db.id
  label       = "iam"

  aws_iam {
    username = "app_iam"
  }
}

resource "formal_resource" "api" {
  name       = "billing-api"
  hostname   = "billing.internal"
  technology = "http"
  port       = 443

  native_users_v3_enabled = true
}

resource "formal_native_user_v3" "bearer_token" {
  resource_id = formal_resource.api.id
  label       = "service-account"

  http_bearer {
    header = "Authorization"
    token {
      environment_variable = "BILLING_API_TOKEN"
    }
  }
}

# Credentials resolved at connection time by a hook running on the connector,
# for example to mint short-lived credentials from a secrets manager.
resource "formal_native_user_v3" "from_hook" {
  resource_id = formal_resource.db.id
  label       = "dynamic"

  hook {
    output_type               = "basic"
    allowlisted_env_variables = ["FORMAL_ENV_USER", "FORMAL_ENV_PASSWORD"]
    code                      = <<-TYPESCRIPT
      export default async function (_input, env) {
        return {
          username: env.FORMAL_ENV_USER,
          password: env.FORMAL_ENV_PASSWORD,
        };
      }
    TYPESCRIPT
  }
}
