# ECSタスク実行ロール（共有）
resource "aws_iam_role" "ecs_task_execution_role" {
  name = "${var.environment}-shared-task-execution-role"

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
    Name = "${var.environment}-shared-task-execution-role"
  }
}

# ECSタスク実行ロールにポリシーをアタッチ
resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# ECSクラスター（共有）
resource "aws_ecs_cluster" "main" {
  name = "${var.environment}-shared-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Name = "${var.environment}-shared-cluster"
  }
}

# CloudWatch Log Group（共有、ログの構造化のため）
resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/${var.environment}-shared-cluster"
  retention_in_days = var.log_retention_days

}
