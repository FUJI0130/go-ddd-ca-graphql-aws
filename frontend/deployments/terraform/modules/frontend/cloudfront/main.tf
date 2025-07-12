# frontend/deployments/terraform/modules/frontend/cloudfront/main.tf

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

# CloudFrontディストリビューション
resource "aws_cloudfront_distribution" "frontend" {
  origin {
    domain_name = var.s3_bucket_domain_name
    origin_id   = "S3-${var.s3_bucket_id}"

    s3_origin_config {
      origin_access_identity = var.cloudfront_oai_path
    }
  }

  enabled             = true
  is_ipv6_enabled     = true
  comment             = "${var.environment}-${var.app_name}-frontend CloudFront distribution"
  default_root_object = "index.html"

  # キャッシュビヘイビア（デフォルト）
  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-${var.s3_bucket_id}"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = var.default_cache_ttl
    max_ttl                = var.max_cache_ttl
    compress               = true
  }

  # 静的アセット用キャッシュビヘイビア（長期キャッシュ）
  ordered_cache_behavior {
    path_pattern     = "/assets/*"
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD", "OPTIONS"]
    target_origin_id = "S3-${var.s3_bucket_id}"

    forwarded_values {
      query_string = false
      headers      = ["Origin"]
      cookies {
        forward = "none"
      }
    }

    min_ttl                = 0
    default_ttl            = var.static_cache_ttl
    max_ttl                = var.static_cache_ttl
    compress               = true
    viewer_protocol_policy = "redirect-to-https"
  }

  # SPA対応のカスタムエラーページ
  custom_error_response {
    error_caching_min_ttl = 10
    error_code            = 403
    response_code         = 200
    response_page_path    = "/index.html"
  }

  custom_error_response {
    error_caching_min_ttl = 10
    error_code            = 404
    response_code         = 200
    response_page_path    = "/index.html"
  }

  # 地理的制限（必要に応じて設定）
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  # SSL/TLS設定
  viewer_certificate {
    cloudfront_default_certificate = var.certificate_arn == null || var.certificate_arn == ""
    acm_certificate_arn            = var.certificate_arn != null && var.certificate_arn != "" ? var.certificate_arn : null
    ssl_support_method             = var.certificate_arn != null && var.certificate_arn != "" ? "sni-only" : null
    minimum_protocol_version       = var.certificate_arn != null && var.certificate_arn != "" ? "TLSv1.2_2021" : null
  }

  # Alias設定（カスタムドメイン使用時）
  # dynamic "aliases" {
  #   for_each = var.domain_aliases != null ? [1] : []
  #   content {
  #     aliases = var.domain_aliases
  #   }
  # }

  # aliases = var.domain_aliases != null && length(var.domain_aliases) > 0 ? var.domain_aliases : null
  aliases = var.domain_aliases != null && var.domain_aliases != [] ? var.domain_aliases : null

  # Price Class（コスト最適化）
  price_class = var.price_class

  tags = {
    Name        = "${var.environment}-${var.app_name}-frontend-cf"
    Environment = var.environment
    Service     = "frontend"
    ManagedBy   = "terraform"
  }
}

# CloudFrontキャッシュ無効化（オプション）
# resource "aws_cloudfront_invalidation" "frontend" {
#   count           = var.enable_automatic_invalidation ? 1 : 0
#   distribution_id = aws_cloudfront_distribution.frontend.id
#   paths           = ["/*"]
# }
