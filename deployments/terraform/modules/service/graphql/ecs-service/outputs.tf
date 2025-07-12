output "service_id" {
  description = "ID of the ECS service"
  value       = module.base_ecs_service.service_id
}

output "service_name" {
  description = "Name of the ECS service"
  value       = module.base_ecs_service.service_name
}

output "task_definition_arn" {
  description = "ARN of the task definition"
  value       = module.base_ecs_service.task_definition_arn
}

output "security_group_id" {
  description = "ID of the security group"
  value       = module.base_ecs_service.security_group_id
}
