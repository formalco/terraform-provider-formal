terraform {
  required_providers {
    formal = {
      source  = "formalco/formal"
      version = "~> 4.12.8"
    }
  }
}

provider "formal" {
  api_key = var.formal_api_key # you can also use env variable FORMAL_API_KEY
}

resource "formal_permission" "read_only" {
  name        = "logs read-only"
  description = "read only permission for logs"
  code        = <<-EOF
package formal.app

import future.keywords.if
import future.keywords.in

default_app_set := {"Logs"}

# Allow full access to Security Team
allow if {
	"Security Team" in input.user.groups
}
EOF
  status      = "draft"
}
