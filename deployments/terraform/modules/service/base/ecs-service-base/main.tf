# サービス固有のIAMタスクロール
resource "aws_iam_role" "ecs_task_role" {
  name = "${var.environment}-${var.service_name}${var.name_suffix}-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.environment}-${var.service_name}-task-role"
    Environment = var.environment
    Service     = var.service_name
  }
}

# サービス固有のセキュリティグループ
resource "aws_security_group" "ecs_tasks" {
  name        = "${var.environment}-${var.service_name}${var.name_suffix}-tasks-sg"
  description = "Allow inbound traffic to ECS tasks for ${var.service_name}"
  vpc_id      = var.vpc_id

  # アプリケーションポート
  ingress {
    from_port   = var.container_port
    to_port     = var.container_port
    protocol    = "tcp"
    description = "Allow app port traffic"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # 追加のポートマッピング（条件付き）
  dynamic "ingress" {
    for_each = var.additional_port_mappings
    content {
      from_port   = ingress.value.port
      to_port     = ingress.value.port
      protocol    = "tcp"
      description = ingress.value.description
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = {
    Name        = "${var.environment}-${var.service_name}-tasks-sg"
    Environment = var.environment
    Service     = var.service_name
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.environment}-${var.service_name}${var.name_suffix}"
  retention_in_days = var.log_retention_days

  tags = {
    Environment = var.environment
    Service     = var.service_name
  }
}

# ECSタスク定義
resource "aws_ecs_task_definition" "app" {
  family                   = "${var.environment}-${var.service_name}${var.name_suffix}"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = var.task_execution_role_arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  # ARM64アーキテクチャを明示的に指定
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }

  container_definitions = jsonencode([
    {
      name      = var.service_name
      image     = var.image_uri
      essential = true

      portMappings = concat(
        [
          {
            containerPort = var.container_port
            hostPort      = var.container_port
            protocol      = "tcp"
          }
        ],
        var.additional_container_port_mappings
      )

      environment = concat(
        [
          for key, value in var.environment_variables : {
            name  = key
            value = value
          }
        ],
        [
          {
            name  = "DB_HOST"
            value = var.db_host
          },
          {
            name  = "DB_PORT"
            value = var.db_port
          },
          {
            name  = "DB_NAME"
            value = var.db_name
          },
          {
            name  = "DB_USER"
            value = var.db_user
          },
          {
            name  = "DB_SSLMODE"
            value = var.db_sslmode
          }
        ],
        var.additional_environment_variables
      )

      secrets = concat(
        [
          {
            name      = "DB_PASSWORD"
            valueFrom = var.db_password_arn
          }
        ],
        var.jwt_secret_arn != "" ? [
          {
            name      = "JWT_SECRET"
            valueFrom = var.jwt_secret_arn
          }
        ] : []
      )

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.app.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = {
    Name        = "${var.environment}-${var.service_name}"
    Environment = var.environment
    Service     = var.service_name
  }
}

# ECSサービス
resource "aws_ecs_service" "app" {
  name                               = "${var.environment}-${var.service_name}${var.name_suffix}"
  cluster                            = var.cluster_id
  task_definition                    = aws_ecs_task_definition.app.arn
  desired_count                      = var.desired_count
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  platform_version                   = "LATEST"
  health_check_grace_period_seconds  = var.health_check_grace_period_seconds
  force_new_deployment               = true
  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200

  # 動的ブロックを使用して複数のロードバランサー設定を可能に
  dynamic "load_balancer" {
    for_each = var.load_balancers != null ? var.load_balancers : []
    content {
      target_group_arn = load_balancer.value.target_group_arn
      container_name   = load_balancer.value.container_name
      container_port   = load_balancer.value.container_port
    }
  }

  # 下位互換性のための単一ロードバランサーサポート
  dynamic "load_balancer" {
    for_each = var.target_group_arn != "" && var.load_balancers == null ? [1] : []
    content {
      target_group_arn = var.target_group_arn
      container_name   = var.service_name
      container_port   = var.container_port
    }
  }

  network_configuration {
    security_groups  = [aws_security_group.ecs_tasks.id]
    subnets          = var.subnet_ids
    assign_public_ip = false
  }

  lifecycle {
    ignore_changes = [desired_count]
  }

  tags = {
    Name        = "${var.environment}-${var.service_name}"
    Environment = var.environment
    Service     = var.service_name
  }
}

# Auto Scaling Target
resource "aws_appautoscaling_target" "app" {
  service_namespace  = "ecs"
  resource_id        = "service/${var.cluster_name}/${aws_ecs_service.app.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = var.desired_count
  max_capacity       = var.max_capacity
}

# CPU Utilization Auto Scaling Policy
resource "aws_appautoscaling_policy" "cpu" {
  name               = "${var.environment}-${var.service_name}${var.name_suffix}-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.app.resource_id
  scalable_dimension = aws_appautoscaling_target.app.scalable_dimension
  service_namespace  = aws_appautoscaling_target.app.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = var.cpu_scaling_target_value
    scale_in_cooldown  = var.scale_in_cooldown
    scale_out_cooldown = var.scale_out_cooldown
  }
}

# Memory Utilization Auto Scaling Policy
resource "aws_appautoscaling_policy" "memory" {
  name               = "${var.environment}-${var.service_name}${var.name_suffix}-memory-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.app.resource_id
  scalable_dimension = aws_appautoscaling_target.app.scalable_dimension
  service_namespace  = aws_appautoscaling_target.app.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value       = var.memory_scaling_target_value
    scale_in_cooldown  = var.scale_in_cooldown
    scale_out_cooldown = var.scale_out_cooldown
  }
}
