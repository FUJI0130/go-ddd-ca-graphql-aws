module "base_target_group" {
  source = "../../base/target-group-base"

  environment    = var.environment
  service_name   = var.service_name
  name_suffix    = "-new" # サフィックス追加
  vpc_id         = var.vpc_id
  container_port = var.container_port
  protocol       = var.protocol

  # REST API固有のデフォルト値
  health_check_path                = var.health_check_path
  health_check_port                = var.health_check_port
  health_check_protocol            = var.health_check_protocol
  health_check_interval            = var.health_check_interval
  health_check_timeout             = var.health_check_timeout
  health_check_healthy_threshold   = var.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.health_check_unhealthy_threshold
  health_check_matcher             = var.health_check_matcher
}
