# frontend/deployments/terraform/environments/development/terraform.tfvars

# 基本設定
aws_region  = "ap-northeast-1"
environment = "development"
app_name    = "test-management"

# S3設定（開発環境用）
enable_versioning    = false
deletion_protection  = false
cors_allowed_origins = ["*"]

# CloudFront設定（開発環境用）
certificate_arn               = null
domain_aliases                = null
default_cache_ttl             = 300              # 5分（開発環境では短めに設定）
max_cache_ttl                 = 3600             # 1時間
static_cache_ttl              = 86400            # 24時間
price_class                   = "PriceClass_100" # コスト最適化
enable_automatic_invalidation = false

# ビルド設定
build_command    = "npm run build"
build_output_dir = "dist"
