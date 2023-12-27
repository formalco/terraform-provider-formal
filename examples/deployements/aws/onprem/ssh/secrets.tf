resource "aws_secretsmanager_secret" "formal_tls_cert" {
  name = "${var.name}-formal-tls-cert-ssh"
}

resource "aws_secretsmanager_secret_version" "formal_tls_cert" {
  secret_id     = aws_secretsmanager_secret.formal_tls_cert.id
  secret_string = formal_sidecar.main.formal_control_plane_tls_certificate
}

resource "aws_secretsmanager_secret" "dockerhub_credentials" {
  name = "${var.name}-dockerhub-credentials-common"
}

resource "aws_secretsmanager_secret_version" "dockerhub_credentials" {
  secret_id     = aws_secretsmanager_secret.dockerhub_credentials.id
  secret_string = jsonencode({ "username" : "${var.dockerhub_username}", "password" : "${var.dockerhub_password}" })
}