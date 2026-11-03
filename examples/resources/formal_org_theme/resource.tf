# The organization's display name and logo, shown to its users in the Catalog at
# catalog.formal.ai.
resource "formal_org_theme" "this" {
  display_name      = "Acme"
  logo              = filebase64("${path.module}/logo.svg")
  logo_content_type = "image/svg+xml"
}
