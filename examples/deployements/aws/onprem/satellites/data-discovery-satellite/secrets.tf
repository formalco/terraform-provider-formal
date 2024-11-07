resource "aws_secretsmanager_secret" "satellite_data_classifier_api_key" {
  name = "${var.name}-formal-data-classifier-api-key"
}

resource "aws_secretsmanager_secret_version" "satellite_data_classifier_api_key" {
  secret_id     = aws_secretsmanager_secret.satellite_data_classifier_api_key.id
  secret_string = formal_satellite.main.api_key
}