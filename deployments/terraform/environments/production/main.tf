# environments/production/main.tf

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }

  # バックエンド設定は後で有効化する
  # S3バケットとDynamoDBテーブルを事前に作成する必要がある
  # backend "s3" {
  #   bucket         = "test-management-terraform-state"
  #   key            = "production/terraform.tfstate"
  #   region         = "ap-northeast-1"
  #   encrypt        = true
  #   dynamodb_table = "test-management-terraform-lock"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = "production"
      Project     = "test-management"
      ManagedBy   = "terraform"
    }
  }
}

# VPCとネットワークモジュール
module "networking" {
  source = "${path.module}../../modules/networking"

  environment        = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  public_subnets     = var.public_subnets
  private_subnets    = var.private_subnets
}

# データベースモジュール
module "database" {
  source = "../../modules/database"

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
  multi_az                 = true # 本番環境ではマルチAZ設定を有効化

  depends_on = [module.networking]
}

# ECSモジュール
module "ecs" {
  source = "../../modules/ecs"

  environment        = var.environment
  vpc_id             = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  app_name           = var.app_name
  app_image          = var.app_image
  app_port           = var.app_port
  app_count          = var.app_count
  app_cpu            = var.app_cpu
  app_memory         = var.app_memory
  app_environment    = var.app_environment
  app_secrets        = var.app_secrets
  db_host            = module.database.db_endpoint
  db_name            = var.db_name
  db_username        = var.db_username
  db_password        = var.db_password

  depends_on = [module.database]
}

# ロードバランサーモジュール
module "loadbalancer" {
  source = "../../modules/loadbalancer"

  environment       = var.environment
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  app_port          = var.app_port
  ecs_service_name  = module.ecs.ecs_service_name
  ecs_target_group  = module.ecs.ecs_target_group
  enable_https      = true # 本番環境ではHTTPSを有効化
  domain_name       = var.domain_name
  certificate_arn   = var.certificate_arn

  depends_on = [module.ecs]
}