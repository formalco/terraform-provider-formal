# A public client, such as a desktop app, authenticates with PKCE alone.
resource "formal_oauth_client" "desktop_agent" {
  name = "Desktop agent"
  redirect_uris = [
    "http://127.0.0.1/callback",
    "com.example.agent://oauth/callback",
  ]
}

# A confidential client also authenticates with a Formal-generated secret.
# Change client_secret_rotation_triggers to rotate it.
resource "formal_oauth_client" "cloud_agent" {
  name          = "Cloud agent"
  redirect_uris = ["https://agent.example.com/oauth/callback"]
  confidential  = true

  client_secret_rotation_triggers = {
    rotated_on = "2026-10-15"
  }
}

output "cloud_agent_client_id" {
  value = formal_oauth_client.cloud_agent.id
}

output "cloud_agent_client_secret" {
  value     = formal_oauth_client.cloud_agent.client_secret
  sensitive = true
}
