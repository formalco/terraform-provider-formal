resource "aws_instance" "main" {
  ami           = "ami-093cb9fb2d34920ad"
  instance_type = "t2.micro"

  iam_instance_profile        = aws_iam_instance_profile.ssm_instance_profile.name
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.public.*.id[0]
  vpc_security_group_ids      = [aws_security_group.ec2.id]
}

resource "aws_security_group" "ec2" {
  name   = "ec2"
  vpc_id = aws_vpc.main.id

  ingress {
    protocol         = "tcp"
    from_port        = 22
    to_port          = 22
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    protocol         = "-1"
    from_port        = 0
    to_port          = 0
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}


resource "aws_iam_user" "aws_native_user" {
  name = "ssh-sidecar-user-iam"
}

resource "aws_iam_policy" "ssh_full_access" {
  name = "ssh-policy-iam"

  # AmazonS3FullAccess managed policy ARN
  # You can also create a custom policy with the necessary permissions if needed.
  description = "Ssh proxy policy"
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:ListServices",
        "ec2:DescribeInstances",
        "ssm:DescribeInstanceInformation",
        "ecs:ExecuteCommand",
        "ecs:ListTasks",
        "ec2:DescribeRegions",
        "ecs:DescribeServices",
        "sts:GetCallerIdentity",
        "ecs:DescribeTasks",
        "ecs:DescribeClusters",
        "ecs:ListClusters",
        "ssm:StartSession"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "attach_s3_full_access" {
  name       = "ssh-iam"
  policy_arn = aws_iam_policy.ssh_full_access.arn
  users      = [aws_iam_user.aws_native_user.name]
}

resource "aws_iam_access_key" "example_access_key" {
  user = aws_iam_user.aws_native_user.name
}