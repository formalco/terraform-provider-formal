resource "aws_secretsmanager_secret" "connector_api_key" {
  name = "${var.name}-api-key"
}

resource "aws_secretsmanager_secret_version" "connector_api_key" {
  secret_id     = aws_secretsmanager_secret.connector_api_key.id
  secret_string = formal_connector.main.api_key
}