module "base_load_balancer" {
  source = "../../base/load-balancer-base"

  environment       = var.environment
  service_name      = var.service_name
  name_suffix       = "-new" # サフィックス追加
  vpc_id            = var.vpc_id
  public_subnet_ids = var.public_subnet_ids
  enable_https      = var.enable_https
  certificate_arn   = var.certificate_arn
  target_group_arn  = var.target_group_arn
  # REST API固有のデフォルト値
  enable_deletion_protection = var.environment == "production" ? true : false
}
