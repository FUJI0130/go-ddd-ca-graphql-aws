module "base_ecs_service" {
  source = "../../base/ecs-service-base"

  environment             = var.environment
  service_name            = var.service_name
  name_suffix             = "-new" # サフィックス追加
  cluster_id              = var.cluster_id
  cluster_name            = var.cluster_name
  task_execution_role_arn = var.task_execution_role_arn
  vpc_id                  = var.vpc_id
  subnet_ids              = var.subnet_ids
  region                  = var.region
  image_uri               = var.image_uri
  container_port          = var.container_port
  desired_count           = var.desired_count
  cpu                     = var.cpu
  memory                  = var.memory
  environment_variables   = var.environment_variables
  db_host                 = var.db_host
  db_port                 = var.db_port
  db_name                 = var.db_name
  db_user                 = var.db_user
  db_sslmode              = var.db_sslmode
  db_password_arn         = var.db_password_arn
  log_retention_days      = var.log_retention_days
  target_group_arn        = var.target_group_arn
  load_balancers          = var.load_balancers

  # REST API固有の設定
  additional_environment_variables = [
    {
      name  = "HTTP_PORT"
      value = "8080"
    },
    {
      name  = "SERVICE_TYPE"
      value = "api"
    }
  ]

  # 追加のポートマッピングは必要ない（コンテナポートが既に8080）
  additional_port_mappings           = []
  additional_container_port_mappings = []

  # オートスケーリング設定
  max_capacity                      = var.max_capacity
  cpu_scaling_target_value          = 70
  memory_scaling_target_value       = 70
  scale_in_cooldown                 = 300
  scale_out_cooldown                = 60
  health_check_grace_period_seconds = 60
}
