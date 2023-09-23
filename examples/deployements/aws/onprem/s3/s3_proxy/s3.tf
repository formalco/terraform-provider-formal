resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "public-read"

  tags = {
    Name        = "Test"
    Environment = "S3-test"
  }
}

resource "aws_iam_user" "aws_native_user" {
  name = "example-user"
}

resource "aws_iam_policy" "s3_full_access" {
  name = "s3-full-access-policy"
    
  # AmazonS3FullAccess managed policy ARN
  # You can also create a custom policy with the necessary permissions if needed.
  description = "Full access to Amazon S3"
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "attach_s3_full_access" {
  name = "tf-test-iam"
  policy_arn = aws_iam_policy.s3_full_access.arn
  users      = [aws_iam_user.aws_native_user.name]
}

resource "aws_iam_access_key" "example_access_key" {
  user = aws_iam_user.aws_native_user.name
}

output "iam_access_key_id" {
  value = aws_iam_access_key.example_access_key.id
}

output "iam_secret_access_key" {
  value = aws_iam_access_key.example_access_key.secret
}