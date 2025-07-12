output "db_password_arn" {
  description = "ARN of the database password parameter"
  value       = aws_ssm_parameter.db_password.arn
}
output "jwt_secret_arn" {
  description = "ARN of the JWT secret parameter"
  value       = aws_ssm_parameter.jwt_secret.arn
}
