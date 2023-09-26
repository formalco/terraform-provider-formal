name        = ""
environment = ""

formal_api_key = ""

region             = "ap-southeast-1"
availability_zones = ["ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"]
private_subnets    = ["172.0.0.0/20", "172.0.32.0/20", "172.0.64.0/20"]
public_subnets     = ["172.0.16.0/20", "172.0.48.0/20", "172.0.80.0/20"]

datadog_api_key = ""

dockerhub_username = ""
dockerhub_password = ""

health_check_port              = 8080
snowflake_port                 = 443
redshift_port                 = 5439
http_port                 = 443
s3_port                 = 443
ssh_port = 2022
postgres_port                 = 5432
data_classifier_satellite_port = 50055

snowflake_container_image                 = ""
data_classifier_satellite_container_image = ""

snowflake_sidecar_hostname = ""
snowflake_hostname         = ""

postgres_container_image                 = ""

postgres_sidecar_hostname = ""
postgres_hostname         = ""

postgres_username = ""
postgres_password = ""


http_container_image                 = ""

http_sidecar_hostname = ""
http_hostname = ""

iam_user_key_id = ""
iam_user_secret_key = ""

bucket_name = ""

s3_sidecar_hostname = ""
s3_hostname         = ""
s3_container_image                 = ""

redshift_container_image                 = ""
redshift_sidecar_hostname = ""
redshift_hostname         = ""

redshift_username = ""
redshift_password = ""
