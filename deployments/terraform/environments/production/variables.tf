# environments/production/variables.tf

# 基本設定
variable "aws_region" {
  description = "AWS region to deploy to"
  type        = string
  default     = "ap-northeast-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

# VPC設定
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["ap-northeast-1a", "ap-northeast-1c", "ap-northeast-1d"]
}

variable "public_subnets" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "private_subnets" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.11.0/24", "10.0.12.0/24"]
}

# データベース設定
variable "db_name" {
  description = "Database name"
  type        = string
  default     = "test_management"
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "db_allocated_storage" {
  description = "Allocated storage for RDS instance (GB)"
  type        = number
  default     = 50
}

variable "db_max_allocated_storage" {
  description = "Maximum allocated storage for RDS instance (GB)"
  type        = number
  default     = 200
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

variable "db_backup_retention" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
}

# アプリケーション設定
variable "app_name" {
  description = "Application name"
  type        = string
  default     = "test-management-app"
}

variable "app_image" {
  description = "Docker image for the application"
  type        = string
  default     = "YOUR_AWS_ACCOUNT_ID.dkr.ecr.ap-northeast-1.amazonaws.com/test-management:latest"
}

variable "app_port" {
  description = "Port the application runs on"
  type        = number
  default     = 8080
}

variable "app_count" {
  description = "Number of app instances to run"
  type        = number
  default     = 3
}

variable "app_cpu" {
  description = "CPU units for the app"
  type        = number
  default     = 512
}

variable "app_memory" {
  description = "Memory for the app in MiB"
  type        = number
  default     = 1024
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

# HTTPS設定
variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = "test-management.example.com"
}

variable "certificate_arn" {
  description = "ARN of the SSL certificate for HTTPS"
  type        = string
  default     = ""
}