# AWS Terraform設計書

## 1. 概要
このドキュメントでは、テストケース管理システムのAWS環境をTerraformで構築するための設計について記述します。インフラストラクチャをコードとして管理し、再現性のある環境構築を実現します。

## 2. 全体アーキテクチャ
```
                      ┌─────────────────────────┐
                      │       CloudFront        │
                      └──────────────┬──────────┘
                                    │
┌───────────────────────────────────┼───────────────────────────────────┐
│ VPC                              │                                   │
│                                  ▼                                   │
│  ┌────────────────┐      ┌───────────────────┐                       │
│  │   Public       │      │       ALB         │                       │
│  │   Subnet       │      └─────────┬─────────┘                       │
│  └────────────────┘                │                                 │
│                                    │                                 │
│  ┌────────────────┐      ┌─────────▼─────────┐    ┌────────────────┐ │
│  │   Private      │      │    ECS Cluster    │    │    RDS         │ │
│  │   Subnet       ├──────►    (Fargate)      ├────►   (PostgreSQL) │ │
│  └────────────────┘      └───────────────────┘    └────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

## 3. リソース構成

### 3.1 ネットワーク
- **VPC**: プライベートネットワークの基盤
- **サブネット**: 
  - パブリックサブネット (2AZ)
  - プライベートサブネット (2AZ)
- **ゲートウェイ/ルーティング**:
  - インターネットゲートウェイ
  - NATゲートウェイ
  - ルートテーブル
- **セキュリティグループ**:
  - ALB用
  - ECS用
  - RDS用

### 3.2 コンピューティング
- **ECS Cluster**: 
  - Fargateタイプ
  - クラスター名: `test-management-cluster`
- **ECSタスク定義**:
  - バックエンドサービス用タスク定義
  - フロントエンド用タスク定義
- **ECSサービス**:
  - バックエンドサービス
  - フロントエンドサービス

### 3.3 データベース
- **RDS**:
  - エンジン: PostgreSQL 14.13
  - インスタンスクラス: db.t3.small
  - マルチAZ: 非対応（コスト削減）
  - ストレージ: 20GB (GP2)
  - バックアップ保持期間: 7日間

### 3.4 ロードバランサー
- **ALB**:
  - パブリックサブネットに配置
  - HTTPSリスナー（ACM証明書使用）
  - ターゲットグループ:
    - バックエンドサービス用
    - フロントエンド用

### 3.5 CDN/キャッシュ
- **CloudFront**:
  - オリジン: ALB
  - キャッシュ設定: 静的ファイルのみ
  - 証明書: ACM

### 3.6 モニタリング
- **CloudWatch**:
  - ALBメトリクス
  - ECSメトリクス
  - RDSメトリクス
  - カスタムダッシュボード

## 4. Terraformモジュール構成

### 4.1 モジュール設計
```
terraform/
├── environments/
│   ├── development/
│   └── production/
└── modules/
    ├── networking/
    ├── database/
    ├── ecs/
    ├── loadbalancer/
    └── monitoring/
```

### 4.2 主要モジュール詳細
#### ネットワークモジュール (modules/networking)
```hcl
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  
  name = "test-management-vpc"
  cidr = "10.0.0.0/16"
  
  azs             = ["ap-northeast-1a", "ap-northeast-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]
  
  enable_nat_gateway = true
  single_nat_gateway = true
  
  tags = {
    Environment = var.environment
    Project     = "test-management"
  }
}
```

#### データベースモジュール (modules/database)
```hcl
module "db" {
  source  = "terraform-aws-modules/rds/aws"
  
  identifier = "test-management-db"
  
  engine               = "postgres"
  engine_version       = "14.13"
  family               = "postgres14"
  major_engine_version = "14"
  instance_class       = "db.t3.small"
  
  allocated_storage     = 20
  max_allocated_storage = 100
  
  db_name  = "test_management"
  username = var.db_username
  password = var.db_password
  port     = 5432
  
  vpc_security_group_ids = [module.security_group_rds.security_group_id]
  subnet_ids             = module.vpc.private_subnets
  
  backup_retention_period = 7
  
  tags = {
    Environment = var.environment
    Project     = "test-management"
  }
}
```

#### ECSモジュール (modules/ecs)
```hcl
module "ecs" {
  source = "terraform-aws-modules/ecs/aws"
  
  cluster_name = "test-management-cluster"
  
  cluster_configuration = {
    execute_command_configuration = {
      logging = "OVERRIDE"
      log_configuration = {
        cloud_watch_log_group_name = "/aws/ecs/test-management"
      }
    }
  }
  
  fargate_capacity_providers = {
    FARGATE = {
      default_capacity_provider_strategy = {
        weight = 100
      }
    }
  }
  
  tags = {
    Environment = var.environment
    Project     = "test-management"
  }
}
```

## 5. 変数と環境分離
### 5.1 変数定義
```hcl
// common variables
variable "aws_region" {
  description = "AWS region to deploy to"
  default     = "ap-northeast-1"
}

variable "environment" {
  description = "Environment (development, staging, production)"
  type        = string
}

// database variables
variable "db_username" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

// application variables
variable "app_image" {
  description = "Container image for the application"
  type        = string
}

variable "app_port" {
  description = "Port the application runs on"
  type        = number
  default     = 8080
}
```

### 5.2 環境別設定ファイル
#### development/terraform.tfvars
```hcl
environment = "development"
app_image   = "123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/test-management:dev"
```

#### production/terraform.tfvars
```hcl
environment = "production"
app_image   = "123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/test-management:latest"
```

## 6. CI/CD連携
### 6.1 GitLab CI/CD連携
```yaml
terraform_plan:
  stage: plan
  script:
    - cd terraform/environments/${CI_ENVIRONMENT_NAME}
    - terraform init
    - terraform plan -out=tfplan
  artifacts:
    paths:
      - terraform/environments/${CI_ENVIRONMENT_NAME}/tfplan

terraform_apply:
  stage: apply
  script:
    - cd terraform/environments/${CI_ENVIRONMENT_NAME}
    - terraform apply -auto-approve tfplan
  dependencies:
    - terraform_plan
  when: manual
  only:
    - main
```

## 7. 拡張計画
現在のMVP設計から、将来的に以下の拡張を検討:

- マルチAZ対応のRDS構成
- ElastiCache (Redis) の追加
- AWS WAFによるセキュリティ強化
- Route53での独自ドメイン対応
- S3 + CloudFrontでの静的アセット最適化
- AWS Secrets Managerでのシークレット管理

## 8. セキュリティ考慮事項
- VPC設計: プライベートサブネットでの重要サービス稼働
- IAM: 最小権限の原則に基づいたロール設定
- セキュリティグループ: 必要最低限のポート公開
- 暗号化: 転送中と保存時のデータ暗号化
- シークレット管理: 環境変数の安全な管理