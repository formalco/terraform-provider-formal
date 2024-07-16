resource "aws_secretsmanager_secret" "formal_kubernetes_api_key" {
  name = "formal-kubernetes-proxy-api-key"
}

resource "aws_secretsmanager_secret_version" "formal_kubernetes_api_key" {
  secret_id     = aws_secretsmanager_secret.formal_kubernetes_api_key.id
  secret_string = var.formal_kubernetes_api_key
}