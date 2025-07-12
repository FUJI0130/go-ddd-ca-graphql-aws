resource "aws_lb_target_group" "this" {
  name        = "${var.environment}-${var.service_name}${var.name_suffix}-tg"
  port        = var.container_port
  protocol    = var.protocol
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    interval            = var.health_check_interval
    path                = var.health_check_path
    port                = var.health_check_port
    protocol            = var.health_check_protocol
    timeout             = var.health_check_timeout
    healthy_threshold   = var.health_check_healthy_threshold
    unhealthy_threshold = var.health_check_unhealthy_threshold
    matcher             = var.health_check_matcher
  }

  tags = {
    Name        = "${var.environment}-${var.service_name}-tg"
    Environment = var.environment
    Service     = var.service_name
  }
}
