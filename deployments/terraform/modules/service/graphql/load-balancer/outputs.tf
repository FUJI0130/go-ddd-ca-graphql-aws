output "load_balancer_arn" {
  description = "ARN of the load balancer"
  value       = module.base_load_balancer.load_balancer_arn
}

output "load_balancer_dns_name" {
  description = "DNS name of the load balancer"
  value       = module.base_load_balancer.load_balancer_dns_name
}

output "load_balancer_zone_id" {
  description = "Zone ID of the load balancer"
  value       = module.base_load_balancer.load_balancer_zone_id
}
