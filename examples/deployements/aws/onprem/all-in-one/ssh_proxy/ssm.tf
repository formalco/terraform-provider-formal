resource "aws_ecs_cluster" "ssm-cluster" {
  name = "ssm-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# Creating an ECS task definition
resource "aws_ecs_task_definition" "task" {
  family                   = "service"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE", "EC2"]
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  cpu                      = 512
  memory                   = 2048

  container_definitions = jsonencode([
    {
      name: "nginx",
      image: "nginx:1.23.1",
      cpu: 512,
      memory: 2048,
      essential: true,
      portMappings: [
        {
          containerPort: 80,
          hostPort: 80,
        },
      ],
    },
  ])
}

# Creating an ECS service
resource "aws_ecs_service" "service" {
  name             = "ssm-service"
  cluster          = aws_ecs_cluster.ssm-cluster.id
  task_definition  = aws_ecs_task_definition.task.arn
  desired_count    = 1
  enable_execute_command = true
  launch_type      = "FARGATE"
  platform_version = "LATEST"

  network_configuration {
    assign_public_ip = true
    security_groups  = [aws_security_group.main.id]
    subnets          = var.public_subnets
  }

  lifecycle {
    ignore_changes = [task_definition]
  }
}

resource "aws_iam_policy" "access-ssm" {
  name = "access-ssm"

  policy = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
       {
       "Effect": "Allow",
       "Action": [
            "ssmmessages:CreateControlChannel",
            "ssmmessages:CreateDataChannel",
            "ssmmessages:OpenControlChannel",
            "ssmmessages:OpenDataChannel"
       ],
      "Resource": "*"
      }
   ]
}
EOF
}


resource "aws_iam_role_policy_attachment" "ecs-task-role-policy-attachment-ssm" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.access-ssm.arn

  depends_on = [
    aws_iam_role.ecs_task_role
  ]
}

