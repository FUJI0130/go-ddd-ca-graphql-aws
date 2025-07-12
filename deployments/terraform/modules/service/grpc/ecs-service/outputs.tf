output "task_definition_arn" {
  description = "Task definition ARN"
  value       = module.base_ecs_service.task_definition_arn
}

output "service_id" {
  description = "Service ID"
  value       = module.base_ecs_service.service_id
}

output "service_name" {
  description = "Service name"
  value       = module.base_ecs_service.service_name
}

output "security_group_id" {
  description = "Security group ID"
  value       = module.base_ecs_service.security_group_id
}
