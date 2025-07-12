output "target_group_arn" {
  description = "ARN of the target group"
  value       = module.base_target_group.target_group_arn
}

output "target_group_name" {
  description = "Name of the target group"
  value       = module.base_target_group.target_group_name
}

output "target_group_id" {
  description = "ID of the target group"
  value       = module.base_target_group.target_group_id
}
