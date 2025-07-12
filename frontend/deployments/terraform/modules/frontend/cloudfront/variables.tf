# frontend/deployments/terraform/modules/frontend/cloudfront/variables.tf

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "test-management"
}

variable "s3_bucket_id" {
  description = "S3 bucket ID"
  type        = string
}

variable "s3_bucket_domain_name" {
  description = "S3 bucket domain name"
  type        = string
}

variable "cloudfront_oai_path" {
  description = "CloudFront Origin Access Identity path"
  type        = string
}

variable "certificate_arn" {
  description = "ACM certificate ARN for custom domain (optional)"
  type        = string
  default     = null
}

variable "domain_aliases" {
  description = "Custom domain aliases for CloudFront distribution"
  type        = list(string)
  default     = null
}

variable "default_cache_ttl" {
  description = "Default cache TTL in seconds for HTML files"
  type        = number
  default     = 3600 # 1 hour
}

variable "max_cache_ttl" {
  description = "Maximum cache TTL in seconds"
  type        = number
  default     = 86400 # 24 hours
}

variable "static_cache_ttl" {
  description = "Cache TTL in seconds for static assets"
  type        = number
  default     = 31536000 # 1 year
}

variable "price_class" {
  description = "CloudFront price class"
  type        = string
  default     = "PriceClass_100" # コスト最適化
}

variable "enable_automatic_invalidation" {
  description = "Enable automatic cache invalidation on deployment"
  type        = bool
  default     = false
}
