# frontend/deployments/terraform/environments/development/variables.tf

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "test-management"
}

# S3関連設定
variable "enable_versioning" {
  description = "Enable S3 bucket versioning"
  type        = bool
  default     = false
}

variable "cors_allowed_origins" {
  description = "CORS allowed origins for the S3 bucket"
  type        = list(string)
  default     = ["*"]
}

variable "deletion_protection" {
  description = "Enable deletion protection for production environments"
  type        = bool
  default     = false
}

# CloudFront関連設定
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
  default     = 3600 # 1 hour for development
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

# ビルド関連設定
variable "build_command" {
  description = "Frontend build command"
  type        = string
  default     = "npm run build"
}

variable "build_output_dir" {
  description = "Frontend build output directory"
  type        = string
  default     = "dist"
}
