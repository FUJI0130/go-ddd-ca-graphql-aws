# AWS ECS複数サービスデプロイのTerraform設定ガイド

## 概要

このドキュメントでは、Terraformを使用して複数サービス（REST API、GraphQL、gRPC）を個別のECSサービスとしてデプロイするための設定方法について説明します。複数サービスをデプロイする際の主要なポイントと設定例を提供します。

## 1. ファイル構造

複数サービスデプロイのための基本的なファイル構造：

```
deployments/terraform/
├── environments/
│   └── development/
│       ├── main.tf           # 環境固有の設定
│       ├── terraform.tfvars  # 変数値の設定
│       └── variables.tf      # 変数定義
└── modules/
    ├── database/             # データベースモジュール
    ├── ecs/                  # ECSモジュール
    ├── loadbalancer/         # ロードバランサーモジュール
    └── networking/           # ネットワークモジュール
```

## 2. main.tf の設定

複数サービスを設定するための`main.tf`の主要部分：

```hcl
# APIサービス（REST API）
module "ecs_api" {
  source             = "../../modules/ecs"
  environment        = var.environment
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  
  app_name        = "${var.app_name}-api"
  app_image       = var.api_image
  app_port        = var.api_port
  app_count       = var.api_count
  app_cpu         = var.app_cpu
  app_memory      = var.app_memory
  
  app_environment = merge(var.app_environment, {
    APP_ENVIRONMENT = var.environment
    SERVICE_TYPE    = "api"
  })
  
  app_secrets     = var.app_secrets
  
  # データベース情報
  db_host         = module.database.db_instance_address
  db_name         = var.db_name
  db_username     = var.db_username
  db_password     = var.db_password

  depends_on = [module.database]
}

# GraphQLサービス
module "ecs_graphql" {
  source             = "../../modules/ecs"
  environment        = var.environment
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  
  app_name        = "${var.app_name}-graphql"
  app_image       = var.graphql_image
  app_port        = var.graphql_port
  app_count       = var.graphql_count
  app_cpu         = var.app_cpu
  app_memory      = var.app_memory
  
  app_environment = merge(var.app_environment, {
    APP_ENVIRONMENT = var.environment
    SERVICE_TYPE    = "graphql"
  })
  
  app_secrets     = var.app_secrets
  
  # データベース情報
  db_host         = module.database.db_instance_address
  db_name         = var.db_name
  db_username     = var.db_username
  db_password     = var.db_password

  depends_on = [module.database]
}

# gRPCサービス
module "ecs_grpc" {
  source             = "../../modules/ecs"
  environment        = var.environment
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  
  app_name        = "${var.app_name}-grpc"
  app_image       = var.grpc_image
  app_port        = var.grpc_port
  app_count       = var.grpc_count
  app_cpu         = var.app_cpu
  app_memory      = var.app_memory
  
  app_environment = merge(var.app_environment, {
    APP_ENVIRONMENT = var.environment
    SERVICE_TYPE    = "grpc"
    GRPC_PORT       = var.grpc_port
  })
  
  app_secrets     = var.app_secrets
  
  # データベース情報
  db_host         = module.database.db_instance_address
  db_name         = var.db_name
  db_username     = var.db_username
  db_password     = var.db_password

  depends_on = [module.database]
}

# ロードバランサーモジュール（API用）
module "loadbalancer_api" {
  source            = "../../modules/loadbalancer"
  environment       = var.environment
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  app_port          = var.api_port
  ecs_service_name  = module.ecs_api.ecs_service_name
  ecs_target_group  = module.ecs_api.ecs_target_group
  enable_https      = false
  certificate_arn   = ""
  depends_on        = [module.ecs_api]
}

# ロードバランサーモジュール（GraphQL用）
module "loadbalancer_graphql" {
  source            = "../../modules/loadbalancer"
  environment       = var.environment
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  app_port          = var.graphql_port
  ecs_service_name  = module.ecs_graphql.ecs_service_name
  ecs_target_group  = module.ecs_graphql.ecs_target_group
  enable_https      = false
  certificate_arn   = ""
  depends_on        = [module.ecs_graphql]
}

# ロードバランサーモジュール（gRPC用）
module "loadbalancer_grpc" {
  source            = "../../modules/loadbalancer"
  environment       = var.environment
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  app_port          = var.grpc_port
  ecs_service_name  = module.ecs_grpc.ecs_service_name
  ecs_target_group  = module.ecs_grpc.ecs_target_group
  enable_https      = false
  certificate_arn   = ""
  depends_on        = [module.ecs_grpc]
}
```

## 3. variables.tf の設定

複数サービスをサポートするための変数定義：

```hcl
# 共通アプリケーション設定
variable "app_name" {
  description = "Application name"
  type        = string
  default     = "test-management"
}

variable "app_cpu" {
  description = "CPU units for the app"
  type        = number
  default     = 256
}

variable "app_memory" {
  description = "Memory for the app in MiB"
  type        = number
  default     = 512
}

variable "app_environment" {
  description = "Environment variables for the app"
  type        = map(string)
  default     = {}
}

variable "app_secrets" {
  description = "Secret environment variables for the app"
  type        = map(string)
  default     = {}
  sensitive   = true
}

# API（REST API）サービス設定
variable "api_image" {
  description = "Docker image for the API service"
  type        = string
}

variable "api_port" {
  description = "Port the API service runs on"
  type        = number
  default     = 8080
}

variable "api_count" {
  description = "Number of API service instances to run"
  type        = number
  default     = 1
}

# GraphQLサービス設定
variable "graphql_image" {
  description = "Docker image for the GraphQL service"
  type        = string
}

variable "graphql_port" {
  description = "Port the GraphQL service runs on"
  type        = number
  default     = 8080
}

variable "graphql_count" {
  description = "Number of GraphQL service instances to run"
  type        = number
  default     = 1
}

# gRPCサービス設定
variable "grpc_image" {
  description = "Docker image for the gRPC service"
  type        = string
}

variable "grpc_port" {
  description = "Port the gRPC service runs on"
  type        = number
  default     = 50051
}

variable "grpc_count" {
  description = "Number of gRPC service instances to run"
  type        = number
  default     = 1
}
```

## 4. terraform.tfvars の設定

サービス固有の変数値を設定する例：

```hcl
# 共通アプリケーション設定
app_name  = "test-management"
app_cpu    = 256
app_memory = 512

# アプリケーション環境変数
app_environment = {
  "ENV"       = "development"
  "LOG_LEVEL" = "debug"
}

# API service
api_image = "YOUR_AWS_ACCOUNT_ID.dkr.ecr.ap-northeast-1.amazonaws.com/development-test-management-api:latest"
api_port  = 8080
api_count = 1

# GraphQL service
graphql_image = "YOUR_AWS_ACCOUNT_ID.dkr.ecr.ap-northeast-1.amazonaws.com/development-test-management-graphql:latest"
graphql_port  = 8080
graphql_count = 1

# gRPC service
grpc_image = "YOUR_AWS_ACCOUNT_ID.dkr.ecr.ap-northeast-1.amazonaws.com/development-test-management-grpc:latest"
grpc_port  = 50051
grpc_count = 1
```

## 5. 重要なポイント

### 5.1 サービス固有の環境変数

各サービスには、そのタイプを識別するための環境変数を設定します：

```hcl
app_environment = merge(var.app_environment, {
  APP_ENVIRONMENT = var.environment
  SERVICE_TYPE    = "api"  # "api", "graphql", "grpc"のいずれか
})
```

これにより、コンテナ内のアプリケーションは自身のサービスタイプを認識できます。

### 5.2 リソース命名規則

各サービスには一貫した命名規則を使用し、区別できるようにします：

```hcl
app_name = "${var.app_name}-api"  # "-api", "-graphql", "-grpc"のサフィックス
```

### 5.3 依存関係の管理

適切な依存関係を設定することで、リソースが正しい順序で作成されるよう保証します：

```hcl
depends_on = [module.database]
```

### 5.4 ポート設定

各サービスのデフォルトポート：
- API: 8080
- GraphQL: 8080
- gRPC: 50051

必要に応じて異なるポート番号を設定できます。

## 6. イメージURI更新の自動化

ECRイメージURIを自動的に更新するための`update-tfvars.sh`スクリプトが提供されています。このスクリプトは以下を行います：

1. ECRリポジトリの存在確認
2. リポジトリURIの取得
3. terraform.tfvarsファイルの更新

使用例：
```bash
# すべてのサービスを更新
make update-tfvars-all TF_ENV=development

# 特定のサービスのみ更新
SERVICE_TYPE=api make update-tfvars TF_ENV=development
```

## 7. デプロイフロー

複数サービスのデプロイには以下のワークフローを使用できます：

```bash
# 完全デプロイワークフロー
make deploy-app-workflow TF_ENV=development
```

このコマンドは以下を実行します：
1. ECRイメージの準備（prepare-all-ecr-images）
2. terraform.tfvarsの更新（update-tfvars-all）
3. ECSとロードバランサーのデプロイ（deploy-ecs-complete）
4. デプロイ結果の検証（verify-all-services）

## 8. 注意点と推奨事項

1. **リソース競合の回避**: 複数サービスを同時にデプロイする場合、依存関係とタイミングに注意
2. **メモリとCPUの割り当て**: サービスの要件に応じて適切なリソースを割り当て
3. **ヘルスチェック設定**: 各サービスに適したヘルスチェックパスとパラメータを設定
4. **ログ管理**: 各サービスのログを区別するための命名規則を採用
5. **スケーリング設定**: サービスごとの負荷特性に応じたAuto Scaling設定を検討