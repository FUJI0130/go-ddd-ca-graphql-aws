module "base_load_balancer" {
  source = "../../base/load-balancer-base"

  environment                = var.environment
  service_name               = var.service_name
  name_suffix                = "-new" # サフィックス追加
  vpc_id                     = var.vpc_id
  public_subnet_ids          = var.public_subnet_ids
  enable_https               = false                # HTTPSリスナーは基底モジュールではなく直接定義
  certificate_arn            = ""                   # 証明書は直接リスナーで設定
  target_group_arn           = var.target_group_arn # HTTP用ターゲットグループ
  enable_deletion_protection = var.environment == "production" ? true : false
}

# 注: HTTPSリスナー (aws_lb_listener.grpc_https_new) は
# environment/development/main.tf 内で定義します
