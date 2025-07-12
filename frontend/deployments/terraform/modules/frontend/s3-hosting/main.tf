# frontend/deployments/terraform/modules/frontend/s3-hosting/main.tf

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

# S3バケット（静的ファイルホスティング用）
resource "aws_s3_bucket" "frontend" {
  bucket = "${var.environment}-${var.app_name}-frontend"

  tags = {
    Name        = "${var.environment}-${var.app_name}-frontend"
    Environment = var.environment
    Service     = "frontend"
    ManagedBy   = "terraform"
  }
}

# S3バケット公開アクセス設定（CloudFront経由のみ許可）
resource "aws_s3_bucket_public_access_block" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3バケット静的ウェブサイトホスティング設定
resource "aws_s3_bucket_website_configuration" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "index.html" # SPA対応: すべてのエラーをindex.htmlに
  }
}

# S3バケットバージョニング設定
resource "aws_s3_bucket_versioning" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  versioning_configuration {
    status = var.enable_versioning ? "Enabled" : "Suspended"
  }
}

# S3バケット暗号化設定
resource "aws_s3_bucket_server_side_encryption_configuration" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# CloudFront用Origin Access Identity（OAI）
resource "aws_cloudfront_origin_access_identity" "frontend" {
  comment = "${var.environment}-${var.app_name}-frontend OAI"
}

# S3バケットポリシー（CloudFrontからのアクセスのみ許可）
resource "aws_s3_bucket_policy" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCloudFrontAccess"
        Effect = "Allow"
        Principal = {
          AWS = aws_cloudfront_origin_access_identity.frontend.iam_arn
        }
        Action   = "s3:GetObject"
        Resource = "${aws_s3_bucket.frontend.arn}/*"
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.frontend]
}

# CORS設定（開発時のローカルアクセス用）
resource "aws_s3_bucket_cors_configuration" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = var.cors_allowed_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
