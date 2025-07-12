output "target_group_arn" {
  description = "The ARN of the target group"
  value       = module.base_target_group.target_group_arn
}

output "target_group_name" {
  description = "The name of the target group"
  value       = module.base_target_group.target_group_name
}

output "target_group_id" {
  description = "The ID of the target group"
  value       = module.base_target_group.target_group_id
}
