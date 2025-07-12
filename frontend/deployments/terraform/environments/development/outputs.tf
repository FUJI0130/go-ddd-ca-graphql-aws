# frontend/deployments/terraform/environments/development/outputs.tf

# S3関連出力
output "s3_bucket_id" {
  description = "S3 bucket ID"
  value       = module.s3_hosting.bucket_id
}

output "s3_bucket_arn" {
  description = "S3 bucket ARN"
  value       = module.s3_hosting.bucket_arn
}

output "s3_website_endpoint" {
  description = "S3 website endpoint"
  value       = module.s3_hosting.website_endpoint
}

# CloudFront関連出力
output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID"
  value       = module.cloudfront.distribution_id
}

output "cloudfront_domain_name" {
  description = "CloudFront domain name"
  value       = module.cloudfront.distribution_domain_name
}

output "cloudfront_url" {
  description = "CloudFront URL (HTTPS)"
  value       = module.cloudfront.cloudfront_url
}

# バックエンド連携情報
output "backend_graphql_alb_dns_name" {
  description = "Backend GraphQL ALB DNS name"
  value       = try(data.terraform_remote_state.backend.outputs.graphql_alb_dns_name, "")
}

# フロントエンドアクセス情報
output "frontend_url" {
  description = "Frontend access URL"
  value       = module.cloudfront.cloudfront_url
}

# デプロイ情報
output "deployment_info" {
  description = "Deployment information"
  value = {
    environment     = var.environment
    s3_bucket       = module.s3_hosting.bucket_id
    cloudfront_url  = module.cloudfront.cloudfront_url
    backend_api_url = try(data.terraform_remote_state.backend.outputs.graphql_alb_dns_name, "")
    deployment_time = timestamp()
  }
}
