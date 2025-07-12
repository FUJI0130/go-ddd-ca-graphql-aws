# modules/ecs/variables.tf

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "private_subnet_ids" {
  description = "IDs of the private subnets"
  type        = list(string)
}

variable "app_name" {
  description = "Application name"
  type        = string
}

variable "app_image" {
  description = "Docker image for the application"
  type        = string
}

variable "app_port" {
  description = "Port the application runs on"
  type        = number
}

variable "app_count" {
  description = "Number of app instances to run"
  type        = number
  default     = 1
}

variable "app_cpu" {
  description = "CPU units for the app"
  type        = number
  default     = 256
}

variable "app_memory" {
  description = "Memory for the app in MiB"
  type        = number
  default     = 512
}

variable "app_environment" {
  description = "Environment variables for the app"
  type        = map(string)
  default     = {}
}

variable "app_secrets" {
  description = "Secret environment variables for the app"
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "db_host" {
  description = "Database host"
  type        = string
}

variable "db_name" {
  description = "Database name"
  type        = string
}

variable "db_username" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}