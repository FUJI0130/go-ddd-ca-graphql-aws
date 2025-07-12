# frontend/deployments/terraform/environments/development/main.tf

terraform {
  required_version = ">= 1.0.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
  # バックエンド設定（リモートステート）
  backend "s3" {
    bucket         = "terraform-state-testmgmt"
    key            = "frontend/development/terraform.tfstate"
    region         = "ap-northeast-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = "test-management-frontend"
      ManagedBy   = "terraform"
    }
  }
}

# バックエンドGraphQL ALB DNS名の取得（データソース）
data "terraform_remote_state" "backend" {
  backend = "s3"
  config = {
    bucket = "terraform-state-testmgmt"
    key    = "development/terraform.tfstate"
    region = var.aws_region
  }
}

# S3ホスティングモジュール
module "s3_hosting" {
  source = "../../modules/frontend/s3-hosting"

  environment          = var.environment
  app_name             = var.app_name
  enable_versioning    = var.enable_versioning
  cors_allowed_origins = var.cors_allowed_origins
  deletion_protection  = var.deletion_protection
}

# CloudFrontモジュール
module "cloudfront" {
  source = "../../modules/frontend/cloudfront"

  environment                   = var.environment
  app_name                      = var.app_name
  s3_bucket_id                  = module.s3_hosting.bucket_id
  s3_bucket_domain_name         = module.s3_hosting.bucket_domain_name
  cloudfront_oai_path           = module.s3_hosting.cloudfront_oai_path
  certificate_arn               = var.certificate_arn
  domain_aliases                = var.domain_aliases
  default_cache_ttl             = var.default_cache_ttl
  max_cache_ttl                 = var.max_cache_ttl
  static_cache_ttl              = var.static_cache_ttl
  price_class                   = var.price_class
  enable_automatic_invalidation = var.enable_automatic_invalidation

  depends_on = [module.s3_hosting]
}

# ローカルファイルでバックエンドDNS名を出力（ビルド時使用）
resource "local_file" "backend_config" {
  content = jsonencode({
    graphql_alb_dns_name = try(data.terraform_remote_state.backend.outputs.graphql_alb_dns_name, "")
    cloudfront_url       = module.cloudfront.cloudfront_url
    s3_bucket_name       = module.s3_hosting.bucket_id
    environment          = var.environment
  })
  filename = "${path.module}/backend-config.json"
}
