module "base_target_group" {
  source = "../../base/target-group-base"

  environment    = var.environment
  service_name   = var.service_name
  name_suffix    = "-new" # サフィックス追加
  vpc_id         = var.vpc_id
  container_port = var.container_port
  protocol       = "HTTP"

  # GraphQL固有のデフォルト値
  health_check_path                = "/health"
  health_check_port                = "traffic-port"
  health_check_protocol            = "HTTP"
  health_check_interval            = 45 # GraphQL用に最適化（長いクエリに対応）
  health_check_timeout             = 8  # GraphQL用に最適化
  health_check_healthy_threshold   = 2  # より寛容な設定
  health_check_unhealthy_threshold = 3
  health_check_matcher             = "200"
}
