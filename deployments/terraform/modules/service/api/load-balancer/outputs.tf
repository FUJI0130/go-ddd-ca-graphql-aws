output "load_balancer_arn" {
  description = "The ARN of the load balancer"
  value       = module.base_load_balancer.load_balancer_arn
}

output "load_balancer_dns_name" {
  description = "The DNS name of the load balancer"
  value       = module.base_load_balancer.load_balancer_dns_name
}

output "load_balancer_zone_id" {
  description = "The canonical hosted zone ID of the load balancer"
  value       = module.base_load_balancer.load_balancer_zone_id
}

output "security_group_id" {
  description = "The ID of the security group"
  value       = module.base_load_balancer.security_group_id
}

output "http_listener_arn" {
  description = "The ARN of the HTTP listener"
  value       = module.base_load_balancer.http_listener_arn
}

output "https_listener_arn" {
  description = "The ARN of the HTTPS listener"
  value       = module.base_load_balancer.https_listener_arn
}
