# SSM Parameter for database password
resource "aws_ssm_parameter" "db_password" {
  name        = "/${var.environment}/database/password"
  description = "Database password for ${var.environment} environment"
  type        = "SecureString"
  value       = var.db_password
  overwrite   = true # 既存パラメータを上書きするため追加

  tags = {
    Environment = var.environment
    Service     = "database"
  }
}

# SSM Parameter for JWT secret
resource "aws_ssm_parameter" "jwt_secret" {
  name        = "/${var.environment}/app/jwt_secret"
  description = "JWT secret for ${var.environment} environment"
  type        = "SecureString"
  value       = var.jwt_secret
  overwrite   = true

  tags = {
    Environment = var.environment
    Service     = "app"
  }
}

# IAM policy for accessing SSM parameters
resource "aws_iam_policy" "ssm_parameter_access" {
  name        = "${var.environment}-ssm-parameter-access"
  description = "Allow ECS tasks to access SSM parameters"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ssm:GetParameters",
          "ssm:GetParameter"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:ssm:${var.region}:*:parameter/${var.environment}/*"
      }
    ]
  })
}

# Attach the policy to the task execution role
resource "aws_iam_role_policy_attachment" "task_exec_ssm_policy" {
  role       = var.task_execution_role_name
  policy_arn = aws_iam_policy.ssm_parameter_access.arn
}
