name        = ""
environment = ""

formal_api_key = ""

region             = "us-west-2"
availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
private_subnets    = ["172.0.0.0/20", "172.0.32.0/20", "172.0.64.0/20"]
public_subnets     = ["172.0.16.0/20", "172.0.48.0/20", "172.0.80.0/20"]

datadog_api_key = ""

dockerhub_username = ""
dockerhub_password = ""

health_check_port              = 8080
snowflake_port                 = 443
data_classifier_satellite_port = 50055

snowflake_container_image                 = ""
data_classifier_satellite_container_image = ""

snowflake_sidecar_hostname = ""
snowflake_hostname         = ""

postgres_port                 = 5432
postgres_container_image                 = ""

postgres_sidecar_hostname = ""
postgres_hostname         = ""

postgres_username = ""
postgres_password = ""

http_port                 = 443
http_container_image                 = ""

http_sidecar_hostname = ""
http_hostname = ""