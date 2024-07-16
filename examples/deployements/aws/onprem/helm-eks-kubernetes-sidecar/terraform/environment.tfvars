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
chart_oci = "oci://public.ecr.aws/p2k2c1w3/formalco-kubernetes-helm-chart"

kubernetes_port = 443

kubernetes_sidecar_hostname = ""
kubernetes_hostname         = ""
kubernetes_username         = ""
kubernetes_password         = ""
