terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 5.0"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key
}

variable "formal_api_key" {
  type      = string
  sensitive = true
}

resource "formal_user" "ci_machine" {
  type = "machine"
  name = "github-actions-ci"
}

resource "formal_integration_oidc" "github_actions" {
  name            = "github-actions"
  issuer          = "https://token.actions.githubusercontent.com"
  machine_user_id = formal_user.ci_machine.id
  claim_condition = <<-CEL
    claims.repository == "formalco/monorepo"
  CEL
  # Optional: derive a Formal human end user by email from claims.
  # Examples:
  #   claims.owner_email
  #   claims.sub.split('/').last()  # AWSReservedSSO session name when it is an email
  # end_user_email_expression = "claims.owner_email"
  status = "active"
}

output "oidc_audience" {
  value = formal_integration_oidc.github_actions.audience
}
