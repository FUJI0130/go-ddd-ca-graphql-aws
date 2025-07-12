variable "environment" {
  description = "Environment name"
  type        = string
}

variable "service_name" {
  description = "Service name"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID"
  type        = string
}

variable "container_port" {
  description = "Container port for gRPC native"
  type        = number
  default     = 50051
}

variable "health_check_path" {
  description = "Path for health check"
  type        = string
  default     = "/grpc.health.v1.Health/Check"
}

variable "health_check_port" {
  description = "Port for health check"
  type        = string
  default     = "50051"
}

variable "health_check_interval" {
  description = "Interval for health check"
  type        = number
  default     = 60
}

variable "health_check_timeout" {
  description = "Timeout for health check"
  type        = number
  default     = 15
}

variable "health_check_healthy_threshold" {
  description = "Healthy threshold for health check"
  type        = number
  default     = 2
}

variable "health_check_unhealthy_threshold" {
  description = "Unhealthy threshold for health check"
  type        = number
  default     = 3
}

variable "health_check_matcher" {
  description = "Matcher for health check"
  type        = string
  default     = "0-99" # gRPC特有の成功コード範囲
}
