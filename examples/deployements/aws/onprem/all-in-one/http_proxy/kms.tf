data "aws_caller_identity" "current" {}

resource "aws_kms_key" "http_sidecar_key" {
  description             = "Customer master key for the HTTP sidecar"
  deletion_window_in_days = 10
  key_usage = "ENCRYPT_DECRYPT"
  customer_master_key_spec = "SYMMETRIC_DEFAULT"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid       = "Enable IAM User Permissions",
        Effect    = "Allow",
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        Action    = "kms:*",
        Resource  = "*"
      }
    ]
  })
}


