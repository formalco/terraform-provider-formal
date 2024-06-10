# Override default if needed
region = "us-east-1"

# Provide actual values for these
formal_api_key = "your_actual_api_key_here"
name           = "your_project_name_here"
environment    = "your_environment_here" # e.g., "development", "staging", "production"

cidr               = "172.0.0.0/16"
private_subnets    = ["172.0.0.0/20", "172.0.32.0/20", "172.0.64.0/20"]
public_subnets     = ["172.0.16.0/20", "172.0.48.0/20", "172.0.80.0/20"]
availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]

# Override default if needed
chart_oci = "oci://public.ecr.aws/d6o8b0b1/formal-bigquery-helm-chart"

# Default is usually fine, but can be overridden
bigquery_port = 3306

bigquery_sidecar_hostname = ""
bigquery_username         = ""
bigquery_password         = ""
