# output "target_group_arn" {
#   description = "The ARN of the target group"
#   value       = aws_lb_target_group.main.arn
# }

# output "target_group_name" {
#   description = "The name of the target group"
#   value       = aws_lb_target_group.main.name
# }

# output "target_group_id" {
#   description = "The ID of the target group"
#   value       = aws_lb_target_group.main.id
# }

# ALBリソースの追加
resource "aws_security_group" "alb" {
  name        = "${var.environment}-${var.service_name}${var.name_suffix}-alb-sg"
  description = "Allow inbound traffic to ALB for ${var.service_name}"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow HTTP traffic"
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow HTTPS traffic"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = {
    Name        = "${var.environment}-${var.service_name}${var.name_suffix}-alb-sg"
    Environment = var.environment
    Service     = var.service_name
  }
}

resource "aws_lb" "main" {
  name               = "${var.environment}-${var.service_name}${var.name_suffix}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = var.public_subnet_ids

  enable_deletion_protection = var.enable_deletion_protection

  tags = {
    Name        = "${var.environment}-${var.service_name}${var.name_suffix}-alb"
    Environment = var.environment
    Service     = var.service_name
  }
}

# HTTP Listener
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = var.enable_https ? "redirect" : "forward"

    dynamic "redirect" {
      for_each = var.enable_https ? [1] : []
      content {
        port        = "443"
        protocol    = "HTTPS"
        status_code = "HTTP_301"
      }
    }

    dynamic "forward" {
      for_each = var.enable_https ? [] : [1]
      content {
        target_group {
          arn = var.target_group_arn
        }
      }
    }
  }
}

# HTTPS Listener（オプション）
resource "aws_lb_listener" "https" {
  count             = var.enable_https ? 1 : 0
  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = var.target_group_arn
  }
}
