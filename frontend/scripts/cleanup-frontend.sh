#!/bin/bash

# フロントエンドAWS環境クリーンアップスクリプト
# バックエンドパターンを踏襲したクリーンアップ実装

set -euo pipefail

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TERRAFORM_DIR="${PROJECT_ROOT}/deployments/terraform/environments/development"

# ログファイル設定
LOG_FILE="${PROJECT_ROOT}/logs/cleanup-frontend-$(date +%Y%m%d_%H%M%S).log"
mkdir -p "$(dirname "${LOG_FILE}")"

# 色付きログ関数
log_info() {
    echo -e "\033[36m[INFO]\033[0m $1" | tee -a "${LOG_FILE}"
}

log_success() {
    echo -e "\033[32m[SUCCESS]\033[0m $1" | tee -a "${LOG_FILE}"
}

log_warning() {
    echo -e "\033[33m[WARNING]\033[0m $1" | tee -a "${LOG_FILE}"
}

log_error() {
    echo -e "\033[31m[ERROR]\033[0m $1" | tee -a "${LOG_FILE}"
}

# 環境変数の読み込み
load_terraform_env() {
    if [ -f ~/.env.terraform ]; then
        log_info "Terraformenv環境変数を読み込み中..."
        set -a
        source ~/.env.terraform
        set +a
        log_success "環境変数読み込み完了"
    else
        log_error "~/.env.terraform が見つかりません"
        exit 1
    fi
}

# Terraformの初期化確認
ensure_terraform_init() {
    log_info "Terraform初期化状況を確認中..."
    
    cd "${TERRAFORM_DIR}"
    
    if [ ! -d ".terraform" ]; then
        log_info "Terraformを初期化中..."
        terraform init
        log_success "Terraform初期化完了"
    else
        log_info "Terraform初期化済み"
    fi
}

# フロントエンドリソース状況確認
check_frontend_resources() {
    log_info "フロントエンドリソース状況を確認中..."
    
    cd "${TERRAFORM_DIR}"
    
    # Terraform状態確認
    if terraform state list | grep -q "module.frontend"; then
        log_info "フロントエンドリソースが存在します"
        
        # 主要リソースの詳細確認
        log_info "S3バケット状況:"
        terraform state show module.frontend.module.s3_hosting.aws_s3_bucket.frontend_bucket 2>/dev/null || log_warning "S3バケット情報取得失敗"
        
        log_info "CloudFront状況:"
        terraform state show module.frontend.module.cloudfront.aws_cloudfront_distribution.frontend_distribution 2>/dev/null || log_warning "CloudFront情報取得失敗"
        
        return 0
    else
        log_info "フロントエンドリソースは存在しません"
        return 1
    fi
}

# S3バケット内容削除
cleanup_s3_content() {
    log_info "S3バケット内容削除中..."
    
    # バケット名取得
    BUCKET_NAME=$(terraform output -raw frontend_s3_bucket_name 2>/dev/null || echo "")
    
    if [ -n "${BUCKET_NAME}" ]; then
        log_info "S3バケット '${BUCKET_NAME}' の内容を削除中..."
        aws s3 rm "s3://${BUCKET_NAME}" --recursive || log_warning "S3削除で一部エラーが発生"
        log_success "S3バケット内容削除完了"
    else
        log_warning "S3バケット名が取得できませんでした"
    fi
}

# CloudFront無効化
invalidate_cloudfront() {
    log_info "CloudFront無効化中..."
    
    # CloudFront Distribution ID取得
    DISTRIBUTION_ID=$(terraform output -raw frontend_cloudfront_distribution_id 2>/dev/null || echo "")
    
    if [ -n "${DISTRIBUTION_ID}" ]; then
        log_info "CloudFront Distribution '${DISTRIBUTION_ID}' を無効化中..."
        aws cloudfront create-invalidation \
            --distribution-id "${DISTRIBUTION_ID}" \
            --paths "/*" \
            || log_warning "CloudFront無効化で一部エラーが発生"
        log_success "CloudFront無効化完了"
    else
        log_warning "CloudFront Distribution IDが取得できませんでした"
    fi
}

# Terraformリソース削除
destroy_terraform_resources() {
    log_info "Terraformリソース削除中..."
    
    cd "${TERRAFORM_DIR}"
    
    # フロントエンド専用のターゲット削除
    log_info "フロントエンドモジュールを削除中..."
    terraform destroy \
        -target="module.frontend" \
        -auto-approve \
        || log_error "Terraformリソース削除失敗"
    
    log_success "Terraformリソース削除完了"
}

# コスト確認
check_remaining_costs() {
    log_info "残存リソースとコスト影響を確認中..."
    
    # 簡易コスト確認（バックエンドパターン踏襲）
    aws ec2 describe-instances --query 'Reservations[*].Instances[?State.Name!=`terminated`].[InstanceId,InstanceType,State.Name]' --output table || log_warning "EC2確認失敗"
    aws rds describe-db-instances --query 'DBInstances[*].[DBInstanceIdentifier,DBInstanceClass,DBInstanceStatus]' --output table || log_warning "RDS確認失敗"
    aws elbv2 describe-load-balancers --query 'LoadBalancers[*].[LoadBalancerName,Type,State.Code]' --output table || log_warning "ALB確認失敗"
    aws s3 ls || log_warning "S3確認失敗"
    
    log_info "コスト影響確認完了"
}

# メイン処理
main() {
    log_info "=== フロントエンドAWS環境クリーンアップ開始 ==="
    log_info "開始時刻: $(date)"
    
    # 環境準備
    load_terraform_env
    ensure_terraform_init
    
    # リソース確認
    if check_frontend_resources; then
        # クリーンアップ実行
        cleanup_s3_content
        invalidate_cloudfront
        destroy_terraform_resources
        
        log_success "=== フロントエンドクリーンアップ完了 ==="
    else
        log_info "=== クリーンアップ対象リソースが見つかりませんでした ==="
    fi
    
    # 最終確認
    check_remaining_costs
    
    log_info "完了時刻: $(date)"
    log_info "ログファイル: ${LOG_FILE}"
}

# エラーハンドリング
trap 'log_error "スクリプト実行中にエラーが発生しました"; exit 1' ERR

# スクリプト実行
main "$@"