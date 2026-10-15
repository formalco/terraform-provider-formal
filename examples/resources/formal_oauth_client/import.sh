# The client secret is only returned at creation or rotation, so client_secret
# is empty after import until the next rotation.
terraform import formal_oauth_client.cloud_agent <client_id>
