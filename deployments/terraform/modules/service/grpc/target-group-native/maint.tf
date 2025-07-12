# gRPCネイティブ用ターゲットグループ（直接リソース定義）
resource "aws_lb_target_group" "this" {
  name             = "${var.environment}-${var.service_name}-native-new-tg"
  port             = var.container_port
  protocol         = "HTTP" # gRPCでもベースはHTTPを指定
  protocol_version = "GRPC" # プロトコルバージョンでgRPCを指定
  vpc_id           = var.vpc_id
  target_type      = "ip"

  health_check {
    enabled             = true
    interval            = var.health_check_interval
    path                = var.health_check_path # 例: "/grpc.health.v1.Health/Check"
    port                = var.health_check_port # 明示的にgRPCポートを指定
    timeout             = var.health_check_timeout
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold
    matcher             = var.health_check_matcher # gRPCでは0-99を使用
  }

  tags = {
    Name        = "${var.environment}-${var.service_name}-native-new-tg"
    Environment = var.environment
    Service     = var.service_name
  }
}
