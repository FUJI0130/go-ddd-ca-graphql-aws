# frontend/deployments/terraform/modules/frontend/s3-hosting/variables.tf

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "test-management"
}

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
