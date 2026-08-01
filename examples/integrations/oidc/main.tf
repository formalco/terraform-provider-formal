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
  status          = "active"
}

output "oidc_audience" {
  value = formal_integration_oidc.github_actions.audience
}
