resource "aws_secretsmanager_secret" "formal_tls_cert" {
  name = "${var.name}-formal-tls-cert-redshift"
}

resource "aws_secretsmanager_secret_version" "formal_tls_cert" {
  secret_id     = aws_secretsmanager_secret.formal_api_key.id
  secret_string = formal_sidecar.main.formal_control_plane_tls_certificate
}