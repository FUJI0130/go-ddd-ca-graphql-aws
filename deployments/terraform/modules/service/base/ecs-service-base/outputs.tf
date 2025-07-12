output "task_definition_arn" {
  description = "The ARN of the task definition"
  value       = aws_ecs_task_definition.app.arn
}

output "security_group_id" {
  description = "The ID of the security group"
  value       = aws_security_group.ecs_tasks.id
}

output "service_name" {
  description = "The name of the service"
  value       = aws_ecs_service.app.name
}

output "service_id" {
  description = "The ID of the service"
  value       = aws_ecs_service.app.id
}

output "task_role_name" {
  description = "The name of the task role"
  value       = aws_iam_role.ecs_task_role.name
}

output "task_role_arn" {
  description = "The ARN of the task role"
  value       = aws_iam_role.ecs_task_role.arn
}

output "cloudwatch_log_group_name" {
  description = "The name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.app.name
}
