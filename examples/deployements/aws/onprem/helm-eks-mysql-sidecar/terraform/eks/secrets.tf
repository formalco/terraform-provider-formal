resource "aws_secretsmanager_secret" "formal_mysql_api_key" {
  name = "formal-mysql-proxy-api-key"
}

resource "aws_secretsmanager_secret_version" "formal_mysql_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_mysql_api_key.id
  secret_string = var.formal_mysql_api_key
}