output "task_definition_arn" {
  description = "The ARN of the task definition"
  value       = module.base_ecs_service.task_definition_arn
}

output "security_group_id" {
  description = "The ID of the security group"
  value       = module.base_ecs_service.security_group_id
}

output "service_name" {
  description = "The name of the service"
  value       = module.base_ecs_service.service_name
}

output "service_id" {
  description = "The ID of the service"
  value       = module.base_ecs_service.service_id
}

output "task_role_name" {
  description = "The name of the task role"
  value       = module.base_ecs_service.task_role_name
}

output "task_role_arn" {
  description = "The ARN of the task role"
  value       = module.base_ecs_service.task_role_arn
}

output "cloudwatch_log_group_name" {
  description = "The name of the CloudWatch log group"
  value       = module.base_ecs_service.cloudwatch_log_group_name
}
