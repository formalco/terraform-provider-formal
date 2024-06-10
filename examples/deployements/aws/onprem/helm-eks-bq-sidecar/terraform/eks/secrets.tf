resource "aws_secretsmanager_secret" "formal_bigquery_api_key" {
  name = "formal-bigquery-proxy-api-key"
}

resource "aws_secretsmanager_secret_version" "formal_bigquery_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_bigquery_api_key.id
  secret_string = var.formal_bigquery_api_key
}