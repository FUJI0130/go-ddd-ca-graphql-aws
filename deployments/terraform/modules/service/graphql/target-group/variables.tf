variable "environment" {
  description = "Environment name (e.g. development, production)"
  type        = string
}

variable "service_name" {
  description = "Name of the service (e.g. graphql)"
  type        = string
}

variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "container_port" {
  description = "Port on which the container will listen"
  type        = number
  default     = 8080
}

variable "health_check_path" {
  description = "Path for health checks"
  type        = string
  default     = "/health"
}

variable "health_check_port" {
  description = "Port for health checks"
  type        = string
  default     = "traffic-port"
}

variable "health_check_protocol" {
  description = "Protocol for health checks"
  type        = string
  default     = "HTTP"
}

variable "health_check_interval" {
  description = "Interval for health checks (seconds)"
  type        = number
  default     = 45
}

variable "health_check_timeout" {
  description = "Timeout for health checks (seconds)"
  type        = number
  default     = 8
}

variable "health_check_healthy_threshold" {
  description = "Number of successful health checks before considering healthy"
  type        = number
  default     = 2
}

variable "health_check_unhealthy_threshold" {
  description = "Number of failed health checks before considering unhealthy"
  type        = number
  default     = 3
}

variable "health_check_matcher" {
  description = "HTTP response codes to consider healthy"
  type        = string
  default     = "200"
}
