resource "aws_secretsmanager_secret" "formal_connector_api_key" {
  name = "${var.name}-formal-api-key-connector-demo"
}

resource "aws_secretsmanager_secret_version" "formal_connector_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_connector_api_key.id
  secret_string = formal_connector.main.api_key
}