output "load_balancer_arn" {
  description = "The ARN of the load balancer"
  value       = module.base_load_balancer.load_balancer_arn
}

output "load_balancer_dns_name" {
  description = "The DNS name of the load balancer"
  value       = module.base_load_balancer.load_balancer_dns_name
}

output "load_balancer_zone_id" {
  description = "The zone ID of the load balancer"
  value       = module.base_load_balancer.load_balancer_zone_id
}

output "load_balancer_security_group_id" {
  description = "The security group ID of the load balancer"
  value       = module.base_load_balancer.security_group_id
}
