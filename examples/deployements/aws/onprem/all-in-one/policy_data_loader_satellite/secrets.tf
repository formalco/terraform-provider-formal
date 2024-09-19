resource "aws_secretsmanager_secret" "formal_api_key" {
  name = "${var.name}-formal-api-policy-data-loader-demo"
}

resource "aws_secretsmanager_secret_version" "formal_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_api_key.id
  secret_string = formal_satellite.main.api_key
}

resource "aws_secretsmanager_secret" "zendesk_api_token" {
  name = "${var.name}-zendesk-api-token"
}

resource "aws_secretsmanager_secret_version" "zendesk_api_token" {
  secret_id     = aws_secretsmanager_secret.zendesk_api_token.id
  secret_string = var.zendesk_api_token
}
