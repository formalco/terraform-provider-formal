# OAuth settings are imported by the ID of their Resource. client_secret is
# never returned; keeping it in config after import doesn't replace the settings.
terraform import formal_resource_oauth_settings.github <resource_id>
