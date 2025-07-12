variable "environment" {
  description = "The name of the environment (e.g. development, production)"
  type        = string
}

variable "service_name" {
  description = "The name of the service"
  type        = string
}

variable "vpc_id" {
  description = "The ID of the VPC"
  type        = string
}

variable "container_port" {
  description = "The port on which the container listens"
  type        = number
}

variable "protocol" {
  description = "The protocol to use for routing traffic to the targets"
  type        = string
  default     = "HTTP"
}

variable "health_check_interval" {
  description = "The approximate amount of time, in seconds, between health checks"
  type        = number
  default     = 30
}

variable "health_check_path" {
  description = "The destination for the health check request"
  type        = string
  default     = "/health"
}

variable "health_check_port" {
  description = "The port to use to connect with the target"
  type        = string
  default     = "traffic-port"
}

variable "health_check_protocol" {
  description = "The protocol to use to connect with the target"
  type        = string
  default     = "HTTP"
}

variable "health_check_timeout" {
  description = "The amount of time, in seconds, to wait before marking a health check as failed"
  type        = number
  default     = 5
}

variable "health_check_healthy_threshold" {
  description = "The number of consecutive health checks successes required before considering an unhealthy target healthy"
  type        = number
  default     = 3
}

variable "health_check_unhealthy_threshold" {
  description = "The number of consecutive health check failures required before considering the target unhealthy"
  type        = number
  default     = 3
}

variable "health_check_matcher" {
  description = "The HTTP codes to use when checking for a successful response from a target"
  type        = string
  default     = "200"
}
