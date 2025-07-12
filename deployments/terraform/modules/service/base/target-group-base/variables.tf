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
}

variable "health_check_interval" {
  description = "The approximate amount of time, in seconds, between health checks"
  type        = number
}

variable "health_check_path" {
  description = "The destination for the health check request"
  type        = string
}

variable "health_check_port" {
  description = "The port to use to connect with the target"
  type        = string
}

variable "health_check_protocol" {
  description = "The protocol to use to connect with the target"
  type        = string
}

variable "health_check_timeout" {
  description = "The amount of time, in seconds, to wait before marking a health check as failed"
  type        = number
}

variable "health_check_healthy_threshold" {
  description = "The number of consecutive health checks successes required before considering an unhealthy target healthy"
  type        = number
}

variable "health_check_unhealthy_threshold" {
  description = "The number of consecutive health check failures required before considering the target unhealthy"
  type        = number
}

variable "health_check_matcher" {
  description = "The HTTP codes to use when checking for a successful response from a target"
  type        = string
}

variable "name_suffix" {
  description = "サフィックス（リソース名の競合を避けるため）"
  type        = string
  default     = ""
}

# # 変数定義を追加（variables.tf）
# variable "target_group_arn" {
#   description = "ARN of the target group"
#   type        = string
# }

# variable "target_group_name" {
#   description = "Name of the target group"
#   type        = string
#   default     = ""
# }

# variable "target_group_id" {
#   description = "ID of the target group"
#   type        = string
#   default     = ""
# }
