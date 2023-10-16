# resource "aws_ecr_repository" "my_helm_chart_repo" {
#   name = "formal-helm-chart"
# }

# resource "aws_iam_role" "ecr_push_role" {
#   name = "ecr-push-role"

#   assume_role_policy = jsonencode({
#     Version = "2012-10-17",
#     Statement = [
#       {
#         Action = "sts:AssumeRole",
#         Effect = "Allow",
#         Principal = {
#           Service = "ecr.amazonaws.com"
#         }
#       }
#     ]
#   })
# }

# resource "aws_iam_policy" "ecr_push_policy" {
#   name = "ecr-push-policy"

#   description = "Policy to allow pushing Helm charts to ECR"

#   policy = jsonencode({
#     Version = "2012-10-17",
#     Statement = [
#       {
#         Action   = [
#           "ecr:GetDownloadUrlForLayer",
#           "ecr:BatchCheckLayerAvailability",
#           "ecr:GetRepositoryPolicy",
#           "ecr:DescribeRepositories",
#           "ecr:GetRepositoryPolicy",
#           "ecr:ListImages",
#           "ecr:DescribeImageScanFindings",
#           "ecr:GetLifecyclePolicy",
#           "ecr:GetLifecyclePolicyPreview",
#           "ecr:GetImage",
#           "ecr:GetImageTagMutability",
#           "ecr:BatchGetImage",
#           "ecr:GetAuthorizationToken"
#         ],
#         Effect   = "Allow",
#         Resource = aws_ecr_repository.my_helm_chart_repo.arn
#       }
#     ]
#   })
# }

# resource "aws_iam_role_policy_attachment" "attach_policy" {
#   policy_arn = aws_iam_policy.ecr_push_policy.arn
#   role       = aws_iam_role.ecr_push_role.name
# }

# output "ecr_repository_url" {
#   value = aws_ecr_repository.my_helm_chart_repo.repository_url
# }