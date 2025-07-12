# environments/development/variables.tf

# 基本設定
variable "aws_region" {
  description = "AWS region to deploy to"
  type        = string
  default     = "ap-northeast-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
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
  default     = ["ap-northeast-1a", "ap-northeast-1c"]
}

variable "public_subnets" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnets" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.11.0/24"]
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
  default     = "db.t3.small"
}

variable "db_allocated_storage" {
  description = "Allocated storage for RDS instance (GB)"
  type        = number
  default     = 20
}

variable "db_max_allocated_storage" {
  description = "Maximum allocated storage for RDS instance (GB)"
  type        = number
  default     = 100
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
  default     = 7
}

# 共通アプリケーション設定
variable "app_name" {
  description = "Application name"
  type        = string
  default     = "test-management"
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

# API（REST API）サービス設定
variable "api_image" {
  description = "Docker image for the API service"
  type        = string
}

variable "api_port" {
  description = "Port the API service runs on"
  type        = number
  default     = 8080
}

variable "api_count" {
  description = "Number of API service instances to run"
  type        = number
  default     = 1
}

# GraphQLサービス設定
variable "graphql_image" {
  description = "Docker image for the GraphQL service"
  type        = string
}

variable "graphql_port" {
  description = "Port the GraphQL service runs on"
  type        = number
  default     = 8080
}

variable "graphql_count" {
  description = "Number of GraphQL service instances to run"
  type        = number
  default     = 1
}

# gRPCサービス設定
variable "grpc_image" {
  description = "Docker image for the gRPC service"
  type        = string
}

variable "grpc_port" {
  description = "Port the gRPC service runs on"
  type        = number
  default     = 50051
}

variable "grpc_count" {
  description = "Number of gRPC service instances to run"
  type        = number
  default     = 1
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}
variable "route53_zone_id" {
  description = "Route53のホストゾーンID（証明書のDNS検証に使用）"
  type        = string
  default     = "" # デフォルト値は空文字列
}

variable "certificate_arn" {
  description = "ACM証明書のARN（マネジメントコンソールで作成した証明書）"
  type        = string
  default     = ""
}

variable "grpc_cpu" {
  description = "gRPCサービスのCPU割り当て"
  type        = number
  default     = 256
}

variable "grpc_memory" {
  description = "gRPCサービスのメモリ割り当て"
  type        = number
  default     = 512
}

variable "jwt_secret" {
  description = "JWT secret for authentication"
  type        = string
  sensitive   = true
}
