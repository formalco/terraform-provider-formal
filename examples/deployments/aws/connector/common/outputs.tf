output "vpc_id" {
  value = aws_vpc.main.id
}

output "docker_hub_secret_arn" {
  value = aws_secretsmanager_secret.dockerhub_credentials.arn
}

output "ecs_cluster_id" {
  value = aws_ecs_cluster.main.id
}

output "ecs_cluster_name" {
  value = aws_ecs_cluster.main.name
}

output "private_subnets" {
  value = aws_subnet.private.*.id
}

output "public_subnets" {
  value = aws_subnet.public.*.id
}

output "url" {
  value = aws_lb.main.dns_name
}

output "nlb_id" {
  value = aws_lb.main.id
}

output "ecs_task_execution_role_arn" {
  value = aws_iam_role.ecs_task_execution_role.arn
}

output "ecs_task_role_arn" {
  value = aws_iam_role.ecs_task_role.arn
}

output "ecs_task_role" {
  value = aws_iam_role.ecs_task_role
}