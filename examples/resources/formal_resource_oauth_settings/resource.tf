resource "formal_resource" "linear" {
  name       = "linear"
  technology = "mcp"
  hostname   = "mcp.linear.app"
  port       = 443
  tags = {
    "mcp-endpoint" = "/mcp"
  }
}

# Users link their Linear accounts with Formal's OAuth app.
resource "formal_resource_oauth_settings" "linear" {
  resource_id = formal_resource.linear.id
}

resource "formal_resource" "github" {
  name       = "github"
  technology = "http"
  hostname   = "api.github.com"
  port       = 443
}

# Users link their GitHub accounts with your own GitHub OAuth app. GitHub
# publishes no protected resource metadata, so set its issuer.
resource "formal_resource_oauth_settings" "github" {
  resource_id         = formal_resource.github.id
  upstream_oauth_mode = "pre_registered"
  client_id           = var.github_oauth_app_client_id
  client_secret       = var.github_oauth_app_client_secret
  issuer              = "https://github.com/login/oauth"
  requested_scopes    = ["repo", "read:org"]
}

variable "github_oauth_app_client_id" {
  type = string
}

variable "github_oauth_app_client_secret" {
  type      = string
  sensitive = true
}
