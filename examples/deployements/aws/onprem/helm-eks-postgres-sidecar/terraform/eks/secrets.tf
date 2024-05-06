resource "aws_secretsmanager_secret" "formal_postgres_api_key" {
  name = "formal-postgres-proxy-api-key"
}

resource "aws_secretsmanager_secret_version" "formal_postgres_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_postgres_api_key.id
  secret_string = var.formal_postgres_api_key
}