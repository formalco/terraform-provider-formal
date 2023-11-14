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