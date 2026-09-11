resource "formal_resource" "db" {
  name       = "analytics-postgres"
  hostname   = "analytics.internal"
  technology = "postgres"
  port       = 5432

  native_users_v3_enabled = true
}

resource "formal_native_user_v3" "admin" {
  resource_id = formal_resource.db.id
  label       = "admin"

  basic {
    username = "app_admin"
    password {
      environment_variable = "ANALYTICS_ADMIN_PASSWORD"
    }
  }
}

resource "formal_native_user_v3" "read_only" {
  resource_id = formal_resource.db.id
  label       = "read-only"

  basic {
    username = "app_readonly"
    password {
      environment_variable = "ANALYTICS_READONLY_PASSWORD"
    }
  }
}

# Sessions from the admins group connect as app_admin; everyone else connects as
# app_readonly.
resource "formal_resource_native_user_selection" "db" {
  resource_id = formal_resource.db.id

  cel = <<-CEL
    "admins" in user.groups ? "${formal_native_user_v3.admin.id}"
    : "${formal_native_user_v3.read_only.id}"
  CEL
}
