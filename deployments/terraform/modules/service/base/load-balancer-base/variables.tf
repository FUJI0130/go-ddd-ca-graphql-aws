variable "environment" {
  description = "Environment name"
  type        = string
}

variable "service_name" {
  description = "Service name (e.g., api, graphql, grpc)"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "public_subnet_ids" {
  description = "IDs of the public subnets"
  type        = list(string)
}

variable "enable_https" {
  description = "Whether to enable HTTPS"
  type        = bool
}

variable "certificate_arn" {
  description = "ARN of the SSL certificate"
  type        = string
}

variable "target_group_arn" {
  description = "The ARN of the target group"
  type        = string
}

variable "enable_deletion_protection" {
  description = "Whether to enable deletion protection for the ALB"
  type        = bool
}

variable "name_suffix" {
  description = "サフィックス（リソース名の競合を避けるため）"
  type        = string
  default     = ""
}
