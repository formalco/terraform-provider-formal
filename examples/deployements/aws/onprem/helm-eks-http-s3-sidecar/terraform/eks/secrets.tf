resource "aws_secretsmanager_secret" "formal_tls_cert" {
  name = "formal-http-proxy-tls-cert"
}

resource "aws_secretsmanager_secret_version" "formal_tls_cert" {
  secret_id     = aws_secretsmanager_secret.formal_tls_cert.id
  secret_string = ""
}

resource "aws_secretsmanager_secret" "formal_data_classifier_tls_cert" {
  name = "formal-data-classifier-tls-cert"
}

resource "aws_secretsmanager_secret_version" "formal_data_classifier_tls_cert" {
  secret_id     = aws_secretsmanager_secret.formal_data_classifier_tls_cert.id
  secret_string = ""
}