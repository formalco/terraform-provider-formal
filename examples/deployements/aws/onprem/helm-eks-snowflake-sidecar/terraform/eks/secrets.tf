resource "aws_secretsmanager_secret" "formal_snowflake_api_key" {
  name = "formal-snowflake-proxy-api-key"
}

resource "aws_secretsmanager_secret_version" "formal_snowflake_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_snowflake_api_key.id
  secret_string = var.formal_snowflake_api_key
}