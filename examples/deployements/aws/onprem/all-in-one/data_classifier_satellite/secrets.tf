resource "aws_secretsmanager_secret" "formal_tls_cert" {
  name = "${var.name}-formal-tls-cert-data-classifier"
}

resource "aws_secretsmanager_secret_version" "formal_tls_cert" {
  secret_id     = aws_secretsmanager_secret.formal_tls_cert.id
  secret_string = "63xd1kARu5GzTfaM9CmtlYH04BiL72O8"
}