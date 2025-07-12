variable "environment" {
  description = "Environment name"
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "task_execution_role_name" {
  description = "Name of the ECS task execution role"
  type        = string
}

variable "jwt_secret" {
  description = "JWT secret for authentication"
  type        = string
  sensitive   = true
}
