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

# variable "container_port" {
#   description = "Container port for HTTP compatibility"
#   type        = number
#   default     = 8080
# }

variable "protocol" {
  description = "Protocol for the target group"
  type        = string
  default     = "HTTP"
}

variable "health_check_path" {
  description = "Path for health check"
  type        = string
  default     = "/health-http"
}

variable "health_check_port" {
  description = "Port for health check"
  type        = string
  default     = "8080"
}

variable "health_check_protocol" {
  description = "Protocol for health check"
  type        = string
  default     = "HTTP"
}

variable "health_check_interval" {
  description = "Interval for health check"
  type        = number
  default     = 60
}

variable "health_check_timeout" {
  description = "Timeout for health check"
  type        = number
  default     = 10
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
  default     = "200"
}
