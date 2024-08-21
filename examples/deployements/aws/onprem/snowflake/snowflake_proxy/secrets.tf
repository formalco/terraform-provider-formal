resource "aws_secretsmanager_secret" "formal_snowflake_api_key" {
  name = "${var.name}-formal-snowflake-api-key"
}

resource "aws_secretsmanager_secret_version" "formal_snowflake_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_snowflake_api_key.id
  secret_string = formal_sidecar.main.api_key
}

resource "aws_secretsmanager_secret" "formal_snowflake_pwd" {
  name = "${var.name}-formal-snowflake-pwd"
}

resource "aws_secretsmanager_secret_version" "formal_snowflake_pwd" {
  secret_id     = aws_secretsmanager_secret.formal_snowflake_pwd.id
  secret_string = var.snowflake_password
}