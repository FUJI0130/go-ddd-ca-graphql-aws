variable "environment" {
  description = "The name of the environment (e.g. development, production)"
  type        = string
}

variable "service_name" {
  description = "The name of the service"
  type        = string
}

variable "cluster_id" {
  description = "The ID of the ECS cluster"
  type        = string
}

variable "cluster_name" {
  description = "The name of the ECS cluster"
  type        = string
}

variable "task_execution_role_arn" {
  description = "The ARN of the IAM role that allows ECS to execute tasks"
  type        = string
}

variable "vpc_id" {
  description = "The ID of the VPC"
  type        = string
}

variable "subnet_ids" {
  description = "The IDs of the subnets"
  type        = list(string)
}

variable "region" {
  description = "The AWS region"
  type        = string
}

variable "image_uri" {
  description = "The URI of the container image"
  type        = string
}

variable "container_port" {
  description = "The port on which the container listens"
  type        = number
}

variable "desired_count" {
  description = "The desired number of tasks"
  type        = number
}

variable "cpu" {
  description = "The number of CPU units used by the task"
  type        = string
}

variable "memory" {
  description = "The amount of memory used by the task"
  type        = string
}

variable "environment_variables" {
  description = "Environment variables for the container"
  type        = map(string)
}

variable "additional_environment_variables" {
  description = "Additional environment variables for the container"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

variable "db_host" {
  description = "The database host"
  type        = string
}

variable "db_port" {
  description = "The database port"
  type        = string
}

variable "db_name" {
  description = "The database name"
  type        = string
}

variable "db_user" {
  description = "The database user"
  type        = string
}

variable "db_sslmode" {
  description = "The database SSL mode"
  type        = string
}

variable "db_password_arn" {
  description = "The ARN of the database password in AWS SSM Parameter Store"
  type        = string
}

variable "log_retention_days" {
  description = "The number of days to retain CloudWatch logs"
  type        = number
}

variable "target_group_arn" {
  description = "The ARN of the target group"
  type        = string
  default     = ""
}

variable "load_balancers" {
  description = "List of load balancer configurations"
  type = list(object({
    target_group_arn = string
    container_name   = string
    container_port   = number
  }))
  default = null
}

variable "additional_port_mappings" {
  description = "Additional port mappings for the security group"
  type = list(object({
    port        = number
    description = string
  }))
  default = []
}

variable "additional_container_port_mappings" {
  description = "Additional port mappings for the container"
  type = list(object({
    containerPort = number
    hostPort      = number
    protocol      = string
  }))
  default = []
}

variable "health_check_grace_period_seconds" {
  description = "The grace period for health checks"
  type        = number
  default     = 60
}

variable "max_capacity" {
  description = "The maximum capacity for auto scaling"
  type        = number
}

variable "cpu_scaling_target_value" {
  description = "The target value for CPU utilization in auto scaling"
  type        = number
  default     = 70
}

variable "memory_scaling_target_value" {
  description = "The target value for memory utilization in auto scaling"
  type        = number
  default     = 70
}

variable "scale_in_cooldown" {
  description = "The cooldown period for scaling in"
  type        = number
  default     = 300
}

variable "scale_out_cooldown" {
  description = "The cooldown period for scaling out"
  type        = number
  default     = 60
}

variable "name_suffix" {
  description = "サフィックス（リソース名の競合を避けるため）"
  type        = string
  default     = ""
}

variable "jwt_secret_arn" {
  description = "ARN of the JWT secret parameter in SSM"
  type        = string
  default     = ""
}
