resource "aws_iam_role" "ecs_task_execution_role" {
  name = "ecs_execution_role_snowflake"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution_role_policy_attachment" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_policy" "ecs_secrets" {
  name        = "ECSAccessToSecrets_snowflake"
  description = "Grant ECS tasks access to secrets"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action   = ["secretsmanager:GetSecretValue"],
        Effect   = "Allow",
        Resource = var.docker_hub_secret_arn

      },
      {
        Action   = ["secretsmanager:GetSecretValue"],
        Effect   = "Allow",
        Resource = aws_secretsmanager_secret.formal_tls_cert.arn

      }
    ]
  })
}

resource "aws_iam_policy_attachment" "ecs_secrets_attachment" {
  name       = "ECSAccessToSecretsAttachment"
  roles      = [aws_iam_role.ecs_task_execution_role.name]
  policy_arn = aws_iam_policy.ecs_secrets.arn
}


resource "aws_iam_role" "ecs_task_role" {
  name = "ecs_task_role_snowflake"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}


# resource "aws_iam_policy" "full-secrets-access" {
#   name = "full-secrets-access-snowflake"

#   policy = <<EOF
# {
#   "Version": "2012-10-17",
#   "Statement": [
#     {
#         "Effect": "Allow",
#         "Action": [
#             "ssm:StartSession",
#             "ssm:DescribeInstanceInformation",
#             "sts:GetCallerIdentity"
#         ],
#         "Resource": "*"
#     },
#     {
#         "Effect": "Allow",
#         "Action": [
#             "ecs:ExecuteCommand",
#             "ecs:ListClusters",
#             "ecs:DescribeClusters",
#             "ecs:ListServices",
#             "ecs:DescribeServices",
#             "ecs:DescribeTasks",
#             "ecs:ListTasks"
#         ],
#         "Resource": "*"
#     },
#     {
#         "Effect": "Allow",
#         "Action": [
#             "ec2:DescribeRegions",
#             "ec2:DescribeInstances"
#         ],
#         "Resource": "*"
#     }
#   ]
# }
# EOF
# }


# resource "aws_iam_role_policy_attachment" "ecs-task-role-policy-attachment-sidecar-destroyer" {
#   role       = aws_iam_role.ecs_task_role.name
#   policy_arn = aws_iam_policy.full-secrets-access.arn

#   depends_on = [
#     aws_iam_role.ecs_task_role
#   ]
# }
