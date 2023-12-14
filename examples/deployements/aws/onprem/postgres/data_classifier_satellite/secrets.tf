resource "aws_secretsmanager_secret" "formal_api_key" {
  name = "${var.name}-formal-api-key"
}

resource "aws_secretsmanager_secret_version" "formal_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_api_key.id
  secret_string = formal_satellite.main.api_key
}
