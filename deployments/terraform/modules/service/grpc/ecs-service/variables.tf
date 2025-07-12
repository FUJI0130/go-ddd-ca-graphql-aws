variable "environment" {
  description = "Environment name"
  type        = string
}

variable "service_name" {
  description = "Service name"
  type        = string
}

variable "cluster_id" {
  description = "ECS cluster ID"
  type        = string
}

variable "cluster_name" {
  description = "ECS cluster name"
  type        = string
}

variable "task_execution_role_arn" {
  description = "Task execution role ARN"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs"
  type        = list(string)
}

variable "region" {
  description = "AWS region"
  type        = string
}

variable "image_uri" {
  description = "Container image URI"
  type        = string
}

variable "desired_count" {
  description = "Desired count of tasks"
  type        = number
  default     = 1
}

variable "cpu" {
  description = "CPU units"
  type        = string
  default     = "256"
}

variable "memory" {
  description = "Memory"
  type        = string
  default     = "512"
}

variable "environment_variables" {
  description = "Environment variables"
  type        = map(string)
  default     = {}
}

variable "max_capacity" {
  description = "Maximum capacity for auto scaling"
  type        = number
  default     = 2
}

variable "db_host" {
  description = "Database host"
  type        = string
}

variable "db_port" {
  description = "Database port"
  type        = string
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_user" {
  description = "Database user"
  type        = string
}

variable "db_sslmode" {
  description = "Database SSL mode"
  type        = string
}

variable "db_password_arn" {
  description = "Database password ARN"
  type        = string
}

variable "load_balancers" {
  description = "Load balancers configuration"
  type = list(object({
    target_group_arn = string
    container_name   = string
    container_port   = number
  }))
  default = []
}

variable "log_retention_days" {
  description = "Log retention days"
  type        = number
  default     = 30
}
