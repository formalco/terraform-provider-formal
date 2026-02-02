# Example: Creating an RDP (Remote Desktop Protocol) Resource

# Create an RDP resource
resource "formal_resource" "windows_server" {
  name       = "windows-server-prod"
  technology = "rdp"
  hostname   = "rdp.example.com"
  port       = 3389
  space_id   = var.space_id

  tags = {
    "environment" = "production"
    "team"        = "platform"
  }

  termination_protection = false
}

# Create a native user for RDP authentication
resource "formal_native_user" "rdp_admin" {
  resource_id            = formal_resource.windows_server.id
  native_user_id         = "administrator"
  native_user_secret     = var.rdp_admin_password # Use var for sensitive data
  use_as_default         = true
  termination_protection = false
}

# Alternative: Use write-only secret (recommended, requires Terraform 1.11+)
resource "formal_native_user" "rdp_user" {
  resource_id                   = formal_resource.windows_server.id
  native_user_id                = "jdoe"
  native_user_secret_wo         = var.rdp_user_password
  native_user_secret_wo_version = 1 # Increment to rotate password
  termination_protection        = false
}

# Variables for secrets
variable "space_id" {
  description = "The ID of the Formal space"
  type        = string
}

variable "rdp_admin_password" {
  description = "Administrator password for RDP"
  type        = string
  sensitive   = true
}

variable "rdp_user_password" {
  description = "User password for RDP"
  type        = string
  sensitive   = true
}

# Output the resource ID
output "rdp_resource_id" {
  value       = formal_resource.windows_server.id
  description = "The ID of the created RDP resource"
}
