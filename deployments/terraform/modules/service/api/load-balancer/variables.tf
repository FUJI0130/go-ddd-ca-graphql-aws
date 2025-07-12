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

variable "container_port" {
  description = "Port the container exposes"
  type        = number
}

variable "enable_https" {
  description = "Whether to enable HTTPS"
  type        = bool
  default     = false
}

variable "certificate_arn" {
  description = "ARN of the SSL certificate"
  type        = string
  default     = ""
}

variable "target_group_arn" {
  description = "The ARN of the target group"
  type        = string
}
