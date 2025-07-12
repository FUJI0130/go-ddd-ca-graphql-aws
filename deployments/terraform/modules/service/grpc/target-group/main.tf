module "base_target_group" {
  source = "../../base/target-group-base"

  environment    = var.environment
  service_name   = var.service_name
  name_suffix    = "-new" # サフィックス追加
  vpc_id         = var.vpc_id
  container_port = 8080   # HTTP用固定
  protocol       = "HTTP" # 明示的に指定


  # 追加：health_check_protocol を明示的に指定
  health_check_protocol = "HTTP"

  # gRPC HTTP用ヘルスチェック設定
  health_check_path                = "/health-http"
  health_check_port                = "8080"
  health_check_interval            = 60
  health_check_timeout             = 10
  health_check_healthy_threshold   = 2
  health_check_unhealthy_threshold = 3
  health_check_matcher             = "200"
}
