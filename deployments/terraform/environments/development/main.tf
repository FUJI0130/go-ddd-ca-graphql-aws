terraform {
  required_version = ">= 1.0.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
  # バックエンド設定
  # S3バケットとDynamoDBテーブルは作成済み
  backend "s3" {
    bucket         = "terraform-state-testmgmt"
    key            = "development/terraform.tfstate"
    region         = "ap-northeast-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = "development"
      Project     = "test-management"
      ManagedBy   = "terraform"
    }
  }
}

# VPCとネットワークモジュール（既存のまま）
module "networking" {
  source             = "../../modules/networking"
  environment        = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  public_subnets     = var.public_subnets
  private_subnets    = var.private_subnets
}

# データベースモジュール（既存のまま）
module "database" {
  source                   = "../../modules/database"
  environment              = var.environment
  vpc_id                   = module.networking.vpc_id
  private_subnet_ids       = module.networking.private_subnet_ids
  db_name                  = var.db_name
  db_instance_class        = var.db_instance_class
  db_allocated_storage     = var.db_allocated_storage
  db_max_allocated_storage = var.db_max_allocated_storage
  db_username              = var.db_username
  db_password              = var.db_password
  db_backup_retention      = var.db_backup_retention
  depends_on               = [module.networking]
}

# 共有ECSクラスターモジュール（新規）
module "shared_ecs_cluster" {
  source      = "../../modules/shared/ecs-cluster"
  environment = var.environment
  region      = var.aws_region
}

# APIサービス
# module "service_api" {
#   source                  = "../../modules/service/ecs-service"
#   environment             = var.environment
#   service_name            = "api"
#   cluster_id              = module.shared_ecs_cluster.cluster_id
#   cluster_name            = module.shared_ecs_cluster.cluster_name
#   task_execution_role_arn = module.shared_ecs_cluster.task_execution_role_arn
#   vpc_id                  = module.networking.vpc_id
#   subnet_ids              = module.networking.private_subnet_ids
#   region                  = var.aws_region

#   image_uri      = var.api_image
#   container_port = var.api_port
#   desired_count  = var.api_count
#   cpu            = var.app_cpu
#   memory         = var.app_memory

#   # 既存の環境変数設定
#   environment_variables = merge(var.app_environment, {
#     APP_ENVIRONMENT = var.environment
#     SERVICE_TYPE    = "api"
#   })

#   # データベース接続パラメータを追加
#   db_host         = module.database.db_instance_address
#   db_port         = "5432"
#   db_name         = var.db_name
#   db_user         = var.db_username
#   db_sslmode      = "require"
#   db_password_arn = module.secrets.db_password_arn
#   # target_group_arn = module.loadbalancer_api.target_group_arn
#   target_group_arn = module.target_group_api.target_group_arn
#   depends_on       = [module.shared_ecs_cluster, module.database, module.target_group_api]
# }

# GraphQLサービス
# module "service_graphql" {
#   source                  = "../../modules/service/ecs-service"
#   environment             = var.environment
#   service_name            = "graphql"
#   cluster_id              = module.shared_ecs_cluster.cluster_id
#   cluster_name            = module.shared_ecs_cluster.cluster_name
#   task_execution_role_arn = module.shared_ecs_cluster.task_execution_role_arn
#   vpc_id                  = module.networking.vpc_id
#   subnet_ids              = module.networking.private_subnet_ids
#   region                  = var.aws_region

#   image_uri      = var.graphql_image
#   container_port = var.graphql_port
#   desired_count  = var.graphql_count
#   cpu            = var.app_cpu
#   memory         = var.app_memory

#   # 既存の環境変数設定
#   environment_variables = merge(var.app_environment, {
#     APP_ENVIRONMENT = var.environment
#     SERVICE_TYPE    = "graphql"
#   })

#   # データベース接続パラメータを追加
#   db_host         = module.database.db_instance_address
#   db_port         = "5432"
#   db_name         = var.db_name
#   db_user         = var.db_username
#   db_sslmode      = "require"
#   db_password_arn = module.secrets.db_password_arn

#   target_group_arn = module.target_group_graphql.target_group_arn
#   depends_on       = [module.shared_ecs_cluster, module.database, module.target_group_graphql]
# }


# # gRPCサービス
# module "service_grpc" {
#   source                  = "../../modules/service/ecs-service"
#   environment             = var.environment
#   service_name            = "grpc"
#   cluster_id              = module.shared_ecs_cluster.cluster_id
#   cluster_name            = module.shared_ecs_cluster.cluster_name # エラー解決：追加
#   task_execution_role_arn = module.shared_ecs_cluster.task_execution_role_arn
#   vpc_id                  = module.networking.vpc_id
#   subnet_ids              = module.networking.private_subnet_ids
#   region                  = var.aws_region # エラー解決：追加

#   # 名前の不一致を解決（container_image → image_uri）
#   image_uri      = var.grpc_image # エラー解決：名前変更
#   container_port = var.grpc_port
#   desired_count  = var.grpc_count
#   cpu            = var.grpc_cpu
#   memory         = var.grpc_memory

#   # 環境変数設定の追加（他のサービスと同様）
#   environment_variables = merge(var.app_environment, { # エラー解決：追加
#     APP_ENVIRONMENT = var.environment
#     SERVICE_TYPE    = "grpc"
#   })

#   # データベース接続パラメータを追加（他のサービスと同様）
#   db_host         = module.database.db_instance_address # エラー解決：追加
#   db_port         = "5432"                              # エラー解決：追加
#   db_name         = var.db_name                         # エラー解決：追加
#   db_user         = var.db_username                     # エラー解決：追加
#   db_sslmode      = "require"                           # エラー解決：追加
#   db_password_arn = module.secrets.db_password_arn      # エラー解決：追加


#   # ロードバランサー設定（変更なし）
#   load_balancers = [
#     {
#       target_group_arn = module.target_group_grpc.target_group_arn
#       container_name   = "grpc"
#       container_port   = 8080 # HTTP用ポート
#     },
#     {
#       target_group_arn = module.target_group_grpc_native.target_group_arn
#       container_name   = "grpc"
#       container_port   = var.grpc_port # gRPC用ポート (50051)
#     }
#   ]

#   depends_on = [module.shared_ecs_cluster, module.database, module.target_group_grpc, module.target_group_grpc_native]
# }


# APIサービス用ロードバランサー
# module "loadbalancer_api" {
#   source                           = "../../modules/service/load-balancer"
#   environment                      = var.environment
#   service_name                     = "api"
#   service_type                     = "api"
#   vpc_id                           = module.networking.vpc_id
#   public_subnet_ids                = module.networking.public_subnet_ids
#   container_port                   = var.api_port
#   health_check_path                = "/health"
#   health_check_protocol            = "HTTP"
#   health_check_port                = "traffic-port"
#   health_check_interval            = 30
#   health_check_timeout             = 5
#   health_check_healthy_threshold   = 3
#   health_check_unhealthy_threshold = 3
#   health_check_matcher             = "200"
#   enable_https                     = false
#   certificate_arn                  = ""
#   target_group_arn                 = module.target_group_api.target_group_arn
#   depends_on                       = [module.target_group_api]
# }

# GraphQLサービス用ロードバランサー
# module "loadbalancer_graphql" {
#   source                           = "../../modules/service/load-balancer"
#   environment                      = var.environment
#   service_name                     = "graphql"
#   service_type                     = "graphql"
#   vpc_id                           = module.networking.vpc_id
#   public_subnet_ids                = module.networking.public_subnet_ids
#   container_port                   = var.graphql_port
#   health_check_path                = "/health"
#   health_check_protocol            = "HTTP"
#   health_check_port                = "traffic-port"
#   health_check_interval            = 45 # GraphQL固有の設定
#   health_check_timeout             = 8  # GraphQL固有の設定
#   health_check_healthy_threshold   = 3
#   health_check_unhealthy_threshold = 3
#   health_check_matcher             = "200"
#   enable_https                     = false
#   certificate_arn                  = ""

#   target_group_arn = module.target_group_graphql.target_group_arn
#   depends_on       = [module.target_group_graphql]
# }

# gRPCサービス用ロードバランサー
# module "loadbalancer_grpc" {
#   source                           = "../../modules/service/load-balancer"
#   environment                      = var.environment
#   service_name                     = "grpc"
#   service_type                     = "grpc" # 明示的に指定
#   vpc_id                           = module.networking.vpc_id
#   public_subnet_ids                = module.networking.public_subnet_ids
#   container_port                   = var.grpc_port
#   health_check_path                = "/health-http" # gRPC向けの特別なHTTPヘルスチェックパス
#   health_check_port                = "8080"         # HTTPヘルスチェック用ポートを明示的に指定 (追加)
#   health_check_timeout             = 10             # gRPC固有の設定
#   health_check_interval            = 60             # gRPC固有の設定
#   health_check_unhealthy_threshold = 5              # gRPC固有の設定
#   enable_https                     = false
#   certificate_arn                  = ""

#   target_group_arn = module.target_group_grpc.target_group_arn

#   depends_on = [module.target_group_grpc]
# }

# 共有シークレットモジュールの追加
module "secrets" {
  source                   = "../../modules/shared/secrets"
  environment              = var.environment
  region                   = var.region
  db_password              = var.db_password
  task_execution_role_name = module.shared_ecs_cluster.task_execution_role_name
  jwt_secret               = var.jwt_secret # ← この1行を追加
  depends_on               = [module.shared_ecs_cluster]
}

# APIサービス用ターゲットグループの追加
# module "target_group_api" {
#   source                           = "../../modules/service/target-group"
#   environment                      = var.environment
#   service_name                     = "api"
#   vpc_id                           = module.networking.vpc_id
#   container_port                   = var.api_port
#   health_check_path                = "/health"
#   health_check_interval            = 30
#   health_check_timeout             = 5
#   health_check_healthy_threshold   = 3
#   health_check_unhealthy_threshold = 3
#   health_check_matcher             = "200"
# }

# GraphQLサービス用ターゲットグループの追加
# module "target_group_graphql" {
#   source                           = "../../modules/service/target-group"
#   environment                      = var.environment
#   service_name                     = "graphql"
#   vpc_id                           = module.networking.vpc_id
#   container_port                   = var.graphql_port
#   health_check_path                = "/health"
#   health_check_interval            = 60 # GraphQLは長めの間隔
#   health_check_timeout             = 10
#   health_check_healthy_threshold   = 2
#   health_check_unhealthy_threshold = 3
#   health_check_matcher             = "200"
# }

# gRPCサービス用ターゲットグループの追加
# module "target_group_grpc" {
#   source                           = "../../modules/service/target-group"
#   environment                      = var.environment
#   service_name                     = "grpc"
#   vpc_id                           = module.networking.vpc_id
#   container_port                   = var.grpc_port
#   protocol                         = "HTTP" # gRPCはHTTP/2だがALBではHTTPを指定
#   health_check_path                = "/health-http"
#   health_check_interval            = 60
#   health_check_timeout             = 15
#   health_check_healthy_threshold   = 2
#   health_check_unhealthy_threshold = 3
#   health_check_matcher             = "200"
# }

# 証明書の設定（実際のドメインがない場合は自己署名証明書を使用）
# module "certificates" {
#   source       = "../../modules/certificates"
#   environment  = var.environment
#   service_name = "grpc"
#   domain_name  = "grpc.example.com" # 実際のドメインまたはテスト用ドメイン

#   # Route53を使用しない場合はコメントアウト
#   route53_zone_id = var.route53_zone_id

# }

# 既存証明書の参照を追加
# data "aws_acm_certificate" "existing" {
#   domain      = "grpc.grpc-dev-fuji0130.com" # 実際のドメイン名を指定
#   statuses    = ["ISSUED"]
#   most_recent = true
# }

# gRPCネイティブ用ターゲットグループ（既存のHTTPベースに追加する形）
# module "target_group_grpc_native" {
#   source                           = "../../modules/service/grpc-target-group"
#   environment                      = var.environment
#   service_name                     = "grpc"
#   vpc_id                           = module.networking.vpc_id
#   container_port                   = var.grpc_port
#   health_check_path                = "/grpc.health.v1.Health/Check"
#   health_check_interval            = 60
#   health_check_timeout             = 15
#   health_check_healthy_threshold   = 2
#   health_check_unhealthy_threshold = 3
# }

# gRPCネイティブ用HTTPSリスナー（既存のロードバランサーに追加）
# resource "aws_lb_listener" "grpc_https" {
#   load_balancer_arn = module.loadbalancer_grpc.load_balancer_arn
#   port              = 443
#   protocol          = "HTTPS"
#   ssl_policy        = "ELBSecurityPolicy-2016-08"
#   # certificate_arn   = module.certificates.certificate_arn
#   # certificate_arn = data.aws_acm_certificate.existing.arn
#   certificate_arn = var.certificate_arn # 直接変数から参照
#   default_action {
#     type             = "forward"
#     target_group_arn = module.target_group_grpc_native.target_group_arn
#   }
# }

# 新しいgRPCネイティブサービスモジュールを追加
# module "service_grpc_native" {
#   source              = "../../modules/service/ecs-service-grpc-native"
#   environment         = var.environment
#   service_name        = "grpc"
#   cluster_id          = module.shared_ecs_cluster.cluster_id
#   task_definition_arn = module.service_grpc.task_definition_arn # 既存のタスク定義を再利用
#   desired_count       = var.grpc_count
#   target_group_arn    = module.target_group_grpc_native.target_group_arn
#   container_name      = "grpc"
#   container_port      = var.grpc_port
#   security_group_ids  = [module.service_grpc.security_group_id]
#   subnet_ids          = module.networking.private_subnet_ids

#   depends_on = [module.service_grpc, module.target_group_grpc_native]
# }

# 削除　リソースからモジュールに置き換える
# gRPCネイティブ用のECSサービス連携設定（追加）
# 既存のgRPCサービスに対して2つ目のロードバランサー設定を追加
# resource "aws_ecs_service" "grpc_native" {
#   # 基本的には既存のgrpcサービスと同じ設定を利用
#   name            = "${var.environment}-grpc-native"
#   cluster         = module.shared_ecs_cluster.cluster_id
#   task_definition = module.service_grpc.task_definition_arn # 既存のタスク定義を利用
#   desired_count   = 1

#   # gRPCネイティブターゲットグループへの連携
#   load_balancer {
#     target_group_arn = module.target_group_grpc_native.target_group_arn
#     container_name   = "grpc"
#     container_port   = var.grpc_port # gRPCポート（50051）
#   }

#   # 既存のgRPCサービスと同じネットワーク設定を利用
#   network_configuration {
#     security_groups  = [module.service_grpc.security_group_id]
#     subnets          = module.networking.private_subnet_ids
#     assign_public_ip = false
#   }

#   # 既存サービスが起動していることを前提にするため、依存関係を設定
#   depends_on = [module.service_grpc]

#   # これはダミーサービスなので、変更を無視
#   lifecycle {
#     ignore_changes = all
#   }
# }


# APIサービス用ターゲットグループ（新モジュール）
module "target_group_api_new" {
  source = "../../modules/service/api/target-group"

  environment    = var.environment
  service_name   = "api"
  vpc_id         = module.networking.vpc_id
  container_port = var.api_port

  # REST API固有の設定値を明示的に指定（変更が必要な場合のみ）
  health_check_path                = "/health"
  health_check_port                = "traffic-port"
  health_check_interval            = 30
  health_check_timeout             = 5
  health_check_healthy_threshold   = 3
  health_check_unhealthy_threshold = 3
  health_check_matcher             = "200"
}

# APIサービス用ロードバランサー（新モジュール）
module "loadbalancer_api_new" {
  source = "../../modules/service/api/load-balancer"

  environment       = var.environment
  service_name      = "api"
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  container_port    = var.api_port
  enable_https      = false
  certificate_arn   = ""
  target_group_arn  = module.target_group_api_new.target_group_arn

  depends_on = [module.target_group_api_new]
}

# APIサービス（新モジュール）
module "service_api_new" {
  source = "../../modules/service/api/ecs-service"

  environment             = var.environment
  service_name            = "api"
  cluster_id              = module.shared_ecs_cluster.cluster_id
  cluster_name            = module.shared_ecs_cluster.cluster_name
  task_execution_role_arn = module.shared_ecs_cluster.task_execution_role_arn
  vpc_id                  = module.networking.vpc_id
  subnet_ids              = module.networking.private_subnet_ids
  region                  = var.aws_region

  image_uri      = var.api_image
  container_port = var.api_port
  desired_count  = var.api_count
  cpu            = var.app_cpu
  memory         = var.app_memory

  # 既存の環境変数設定
  environment_variables = merge(var.app_environment, {
    APP_ENVIRONMENT = var.environment
  })

  # データベース接続パラメータ
  db_host         = module.database.db_instance_address
  db_port         = "5432"
  db_name         = var.db_name
  db_user         = var.db_username
  db_sslmode      = "require"
  db_password_arn = module.secrets.db_password_arn

  target_group_arn = module.target_group_api_new.target_group_arn

  depends_on = [module.shared_ecs_cluster, module.database, module.target_group_api_new]
}


# GraphQLサービス用ターゲットグループ（新モジュール）
module "target_group_graphql_new" {
  source = "../../modules/service/graphql/target-group"

  environment    = var.environment
  service_name   = "graphql"
  vpc_id         = module.networking.vpc_id
  container_port = var.graphql_port

  # GraphQL固有の設定値を明示的に指定
  health_check_path                = "/health"
  health_check_interval            = 45
  health_check_timeout             = 8
  health_check_healthy_threshold   = 2
  health_check_unhealthy_threshold = 3
  health_check_matcher             = "200"
}

# GraphQLサービス用ロードバランサー（新モジュール）
module "loadbalancer_graphql_new" {
  source = "../../modules/service/graphql/load-balancer"

  environment       = var.environment
  service_name      = "graphql"
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  enable_https      = true
  certificate_arn   = var.certificate_arn
  target_group_arn  = module.target_group_graphql_new.target_group_arn

  depends_on = [module.target_group_graphql_new]
}

# GraphQLサービス（新モジュール）
module "service_graphql_new" {
  source = "../../modules/service/graphql/ecs-service"

  environment             = var.environment
  service_name            = "graphql"
  cluster_id              = module.shared_ecs_cluster.cluster_id
  cluster_name            = module.shared_ecs_cluster.cluster_name
  task_execution_role_arn = module.shared_ecs_cluster.task_execution_role_arn
  vpc_id                  = module.networking.vpc_id
  subnet_ids              = module.networking.private_subnet_ids
  region                  = var.aws_region

  image_uri      = var.graphql_image
  container_port = var.graphql_port
  desired_count  = var.graphql_count
  cpu            = var.app_cpu
  memory         = var.app_memory

  # ログ保持期間の設定
  log_retention_days = 30

  # 既存の環境変数設定
  environment_variables = merge(var.app_environment, {
    APP_ENVIRONMENT = var.environment
  })

  # データベース接続パラメータ
  db_host         = module.database.db_instance_address
  db_port         = "5432"
  db_name         = var.db_name
  db_user         = var.db_username
  db_sslmode      = "require"
  db_password_arn = module.secrets.db_password_arn

  target_group_arn = module.target_group_graphql_new.target_group_arn

  depends_on = [module.shared_ecs_cluster, module.database, module.target_group_graphql_new]
}



# gRPC HTTP互換ターゲットグループ（新モジュール）
module "target_group_grpc_new" {
  source = "../../modules/service/grpc/target-group"

  environment  = var.environment
  service_name = "grpc"
  vpc_id       = module.networking.vpc_id
  # container_port = 8080 # HTTP互換用ポート

  # HTTP互換用設定
  health_check_path                = "/health-http"
  health_check_interval            = 60
  health_check_timeout             = 10
  health_check_healthy_threshold   = 2
  health_check_unhealthy_threshold = 3
  health_check_matcher             = "200"
}

# gRPCネイティブターゲットグループ（新モジュール）
module "target_group_grpc_native_new" {
  source = "../../modules/service/grpc/target-group-native"

  environment  = var.environment
  service_name = "grpc"
  vpc_id       = module.networking.vpc_id
  # container_port = var.grpc_port # gRPCネイティブ用ポート

  # gRPCネイティブ用設定
  health_check_path                = "/grpc.health.v1.Health/Check"
  health_check_interval            = 60
  health_check_timeout             = 15
  health_check_healthy_threshold   = 2
  health_check_unhealthy_threshold = 3
  health_check_matcher             = "0-99" # gRPC成功コード範囲
}

# gRPCサービス用ロードバランサー（新モジュール）
module "loadbalancer_grpc_new" {
  source = "../../modules/service/grpc/load-balancer"

  environment       = var.environment
  service_name      = "grpc"
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  target_group_arn  = module.target_group_grpc_new.target_group_arn # HTTP互換用

  depends_on = [module.target_group_grpc_new]
}

# gRPCネイティブ用HTTPSリスナー（直接定義）
resource "aws_lb_listener" "grpc_https_new" {
  load_balancer_arn = module.loadbalancer_grpc_new.load_balancer_arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = var.certificate_arn # 変数から証明書ARNを参照

  default_action {
    type             = "forward"
    target_group_arn = module.target_group_grpc_native_new.target_group_arn # gRPCネイティブ用
  }
}

# gRPCサービス（新モジュール）
module "service_grpc_new" {
  source = "../../modules/service/grpc/ecs-service"

  environment             = var.environment
  service_name            = "grpc"
  cluster_id              = module.shared_ecs_cluster.cluster_id
  cluster_name            = module.shared_ecs_cluster.cluster_name
  task_execution_role_arn = module.shared_ecs_cluster.task_execution_role_arn
  vpc_id                  = module.networking.vpc_id
  subnet_ids              = module.networking.private_subnet_ids
  region                  = var.aws_region

  # 追加：log_retention_days を指定
  log_retention_days = 30

  image_uri = var.grpc_image
  # container_port = var.grpc_port # 50051
  desired_count = var.grpc_count
  cpu           = var.grpc_cpu
  memory        = var.grpc_memory

  # 環境変数設定
  environment_variables = merge(var.app_environment, {
    APP_ENVIRONMENT = var.environment
  })

  # データベース接続設定
  db_host         = module.database.db_instance_address
  db_port         = "5432"
  db_name         = var.db_name
  db_user         = var.db_username
  db_sslmode      = "require"
  db_password_arn = module.secrets.db_password_arn

  # 複数ロードバランサー設定
  load_balancers = [
    {
      target_group_arn = module.target_group_grpc_new.target_group_arn
      container_name   = "grpc"
      container_port   = 8080 # HTTP互換用ポート
    },
    {
      target_group_arn = module.target_group_grpc_native_new.target_group_arn
      container_name   = "grpc"
      container_port   = var.grpc_port # gRPCネイティブ用ポート
    }
  ]

  depends_on = [
    module.shared_ecs_cluster,
    module.database,
    module.target_group_grpc_new,
    module.target_group_grpc_native_new,
    module.loadbalancer_grpc_new
  ]
}

# Output定義の追加
output "vpc_id" {
  description = "VPC ID"
  value       = module.networking.vpc_id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.networking.private_subnet_ids
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.networking.public_subnet_ids
}

output "db_instance_address" {
  description = "RDS instance address"
  value       = module.database.db_instance_address
}

output "db_name" {
  description = "Database name"
  value       = var.db_name
}

output "graphql_alb_dns_name" {
  description = "GraphQL ALB DNS name"
  value       = module.loadbalancer_graphql_new.load_balancer_dns_name
}
