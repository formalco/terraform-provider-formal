resource "aws_iam_policy" "eks_secrets" {
  name        = "EKSAccessToSecrets"
  description = "Grant EKS access to secrets"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action   = ["secretsmanager:DescribeSecret", "secretsmanager:GetSecretValue"],
        Effect   = "Allow",
        Resource = aws_secretsmanager_secret.formal_bigquery_api_key.arn
      }
    ]
  })
}