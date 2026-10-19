# OAuth settings are imported by the ID of their Resource. The client secret is
# never returned. client_secret_wo stays out of state; increment
# client_secret_wo_version to send it.
terraform import formal_resource_oauth_settings.github <resource_id>
