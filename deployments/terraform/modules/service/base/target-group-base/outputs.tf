# output "target_group_arn" {
#   description = "The ARN of the target group"
#   value       = aws_lb_target_group.main.arn
# }

# output "target_group_name" {
#   description = "The name of the target group"
#   value       = aws_lb_target_group.main.name
# }

# output "target_group_id" {
#   description = "The ID of the target group"
#   value       = aws_lb_target_group.main.id
# }
output "target_group_arn" {
  description = "The ARN of the target group"
  value       = aws_lb_target_group.this.arn # 自身で作成したリソースを参照
}

output "target_group_name" {
  description = "The name of the target group"
  value       = aws_lb_target_group.this.name # 自身で作成したリソースを参照
}

output "target_group_id" {
  description = "The ID of the target group"
  value       = aws_lb_target_group.this.id # 自身で作成したリソースを参照
}
