resource "aws_iam_role" "ecs_task_execution_role" {
  name = "ecs_execution_role_demo"

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
  name        = "ECSAccessToSecrets_ssh"
  description = "Grant ECS tasks access to secrets"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action   = ["secretsmanager:GetSecretValue"],
        Effect   = "Allow",
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_policy_attachment" "ecs_secrets_attachment" {
  name       = "ECSAccessToSecretsAttachment"
  roles      = [aws_iam_role.ecs_task_execution_role.name]
  policy_arn = aws_iam_policy.ecs_secrets.arn
}

resource "aws_iam_policy" "ecr_access" {
  name        = "ECSAccessToECR"
  description = "Grant ECS tasks access to ECR"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action   = ["ecr:BatchGetImage", "ecr:GetDownloadUrlForLayer", "ecr:GetAuthorizationToken"],
        Effect   = "Allow",
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_policy_attachment" "ecr_attachment" {
  name       = "ECSAccessToSecretsAttachment"
  roles      = [aws_iam_role.ecs_task_execution_role.name]
  policy_arn = aws_iam_policy.ecr_access.arn
}



resource "aws_iam_role" "ecs_task_role" {
  name = "ecs_task_role_demo"

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

resource "aws_iam_policy" "rds_iam_policy" {
  name        = "rds-connect-policy"
  description = "Policy to allow RDS IAM"
  policy      = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
      {
         "Effect": "Allow",
         "Action": [
             "rds-db:connect"
         ],
         "Resource": [
             "arn:aws:rds-db:*:*:dbuser:*/*"
         ]
      }
   ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "rds_connect_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.rds_iam_policy.arn
}


resource "aws_iam_policy" "full-secrets-access" {
  name = "full-secrets-access-ssh"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Effect": "Allow",
        "Action": [
            "ssm:StartSession",
            "ssm:DescribeInstanceInformation",
            "sts:GetCallerIdentity"
        ],
        "Resource": "*"
    },
    {
        "Effect": "Allow",
        "Action": [
            "ecs:ExecuteCommand",
            "ecs:ListClusters",
            "ecs:DescribeClusters",
            "ecs:ListServices",
            "ecs:DescribeServices",
            "ecs:DescribeTasks",
            "ecs:ListTasks"
        ],
        "Resource": "*"
    },
    {
        "Effect": "Allow",
        "Action": [
            "ec2:DescribeRegions",
            "ec2:DescribeInstances"
        ],
        "Resource": "*"
    },
    {
      "Sid": "VisualEditor1",
      "Effect": "Allow",
      "Action": [
          "kms:Decrypt",
          "kms:GenerateDataKey"
      ],
      "Resource":  "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs-task-role-policy-attachment-sidecar-destroyer" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.full-secrets-access.arn

  depends_on = [
    aws_iam_role.ecs_task_role
  ]
}
