#!/bin/bash
# frontend/scripts/upload-frontend.sh
# フロントエンドファイルS3アップロードスクリプト（バックエンドパターン踏襲）

set -euo pipefail

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ENVIRONMENT=${1:-development}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DIST_DIR="${PROJECT_ROOT}/dist"
TF_DIR="${PROJECT_ROOT}/deployments/terraform/environments/${ENVIRONMENT}"

# ログファイル設定
LOG_FILE="${PROJECT_ROOT}/logs/upload-frontend-$(date +%Y%m%d_%H%M%S).log"
mkdir -p "$(dirname "${LOG_FILE}")"

# 色付きログ関数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "${LOG_FILE}"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "${LOG_FILE}"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "${LOG_FILE}"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "${LOG_FILE}"
}

# 環境変数の読み込み
load_environment() {
    if [ -f ~/.env.terraform ]; then
        log_info "環境変数を読み込み中..."
        set -a
        source ~/.env.terraform
        set +a
        log_success "環境変数読み込み完了"
    else
        log_error "~/.env.terraform が見つかりません"
        exit 1
    fi
}

# 前提条件の確認
check_prerequisites() {
    log_info "前提条件を確認中..."
    
    # AWS CLI確認
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLIがインストールされていません"
        exit 1
    fi
    
    # Terraform確認
    if ! command -v terraform &> /dev/null; then
        log_error "Terraformがインストールされていません"
        exit 1
    fi
    
    # AWS認証確認
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS認証情報が設定されていないか、無効です"
        exit 1
    fi
    
    # ビルド成果物確認
    if [ ! -d "${DIST_DIR}" ]; then
        log_error "ビルド成果物が見つかりません: ${DIST_DIR}"
        log_error "先に 'make build-frontend-assets' を実行してください"
        exit 1
    fi
    
    # index.htmlの存在確認
    if [ ! -f "${DIST_DIR}/index.html" ]; then
        log_error "index.html が見つかりません: ${DIST_DIR}/index.html"
        exit 1
    fi
    
    log_success "前提条件確認完了"
}

# S3バケット名取得
get_s3_bucket_name() {
    log_info "S3バケット名を取得中..."
    
    cd "${TF_DIR}"
    
    if [ ! -f ".terraform/terraform.tfstate" ] && [ ! -f "terraform.tfstate" ]; then
        log_error "Terraform状態が見つかりません"
        log_error "先に 'make terraform-deploy-frontend' を実行してください"
        exit 1
    fi
    
    BUCKET_NAME=$(terraform output -raw s3_bucket_id 2>/dev/null || echo "")
    
    if [ -z "${BUCKET_NAME}" ]; then
        log_error "S3バケット名を取得できませんでした"
        log_error "Terraformデプロイが完了しているか確認してください"
        exit 1
    fi
    
    log_success "S3バケット名: ${BUCKET_NAME}"
    cd - > /dev/null
}

# ファイルサイズ計算
calculate_upload_size() {
    log_info "アップロードサイズを計算中..."
    
    TOTAL_SIZE=$(du -sh "${DIST_DIR}" | cut -f1)
    FILE_COUNT=$(find "${DIST_DIR}" -type f | wc -l)
    
    log_info "アップロード対象:"
    log_info "  ファイル数: ${FILE_COUNT}"
    log_info "  総サイズ: ${TOTAL_SIZE}"
    
    # 主要ファイルの詳細
    if [ -f "${DIST_DIR}/index.html" ]; then
        INDEX_SIZE=$(du -h "${DIST_DIR}/index.html" | cut -f1)
        log_info "  index.html: ${INDEX_SIZE}"
    fi
    
    if [ -d "${DIST_DIR}/assets" ]; then
        ASSETS_SIZE=$(du -sh "${DIST_DIR}/assets" | cut -f1)
        ASSETS_COUNT=$(find "${DIST_DIR}/assets" -type f | wc -l)
        log_info "  assets/: ${ASSETS_SIZE} (${ASSETS_COUNT}ファイル)"
    fi
}

# S3への同期アップロード
upload_to_s3() {
    log_info "S3へのアップロードを開始..."
    
    # 既存ファイルのバックアップ情報取得
    log_info "既存ファイルの情報を取得中..."
    EXISTING_COUNT=$(aws s3 ls "s3://${BUCKET_NAME}/" --recursive | wc -l || echo "0")
    
    if [ "${EXISTING_COUNT}" -gt 0 ]; then
        log_info "既存ファイル数: ${EXISTING_COUNT}"
    else
        log_info "新規アップロード（既存ファイルなし）"
    fi
    
    # 進捗表示付きアップロード
    log_info "ファイルをアップロード中..."
    
    # aws s3 syncで効率的にアップロード
    aws s3 sync "${DIST_DIR}/" "s3://${BUCKET_NAME}/" \
        --delete \
        --exact-timestamps \
        --exclude "*.map" \
        --cache-control "public, max-age=31536000" \
        --metadata-directive REPLACE \
        2>&1 | while IFS= read -r line; do
            if [[ $line == upload* ]]; then
                filename=$(echo "$line" | awk '{print $2}' | sed "s|${DIST_DIR}/||")
                log_info "  ✓ ${filename}"
            elif [[ $line == delete* ]]; then
                filename=$(echo "$line" | awk '{print $2}')
                log_warning "  ✗ 削除: ${filename}"
            fi
        done
    
    # index.htmlは短期キャッシュに設定
    log_info "index.htmlのキャッシュ設定を調整中..."
    aws s3 cp "${DIST_DIR}/index.html" "s3://${BUCKET_NAME}/index.html" \
        --cache-control "public, max-age=300" \
        --content-type "text/html" \
        --metadata-directive REPLACE > /dev/null
    
    log_success "S3アップロード完了"
}

# アップロード結果の確認
verify_upload() {
    log_info "アップロード結果を確認中..."
    
    # S3バケット内容確認
    UPLOADED_COUNT=$(aws s3 ls "s3://${BUCKET_NAME}/" --recursive | wc -l)
    UPLOADED_SIZE=$(aws s3 ls "s3://${BUCKET_NAME}/" --recursive --summarize | grep "Total Size" | awk '{print $3, $4}' || echo "不明")
    
    log_success "アップロード結果:"
    log_success "  ファイル数: ${UPLOADED_COUNT}"
    log_success "  総サイズ: ${UPLOADED_SIZE}"
    
    # 主要ファイルの存在確認
    log_info "主要ファイルの存在確認:"
    
    if aws s3 ls "s3://${BUCKET_NAME}/index.html" > /dev/null 2>&1; then
        log_success "  ✓ index.html"
    else
        log_error "  ✗ index.html が見つかりません"
        return 1
    fi
    
    if aws s3 ls "s3://${BUCKET_NAME}/assets/" > /dev/null 2>&1; then
        ASSETS_COUNT=$(aws s3 ls "s3://${BUCKET_NAME}/assets/" --recursive | wc -l)
        log_success "  ✓ assets/ (${ASSETS_COUNT}ファイル)"
    else
        log_warning "  ! assets/ ディレクトリが見つかりません"
    fi
}

# CloudFront情報の取得と表示
get_cloudfront_info() {
    log_info "CloudFront情報を取得中..."
    
    cd "${TF_DIR}"
    
    CLOUDFRONT_URL=$(terraform output -raw frontend_cloudfront_url 2>/dev/null || echo "")
    DISTRIBUTION_ID=$(terraform output -raw frontend_cloudfront_distribution_id 2>/dev/null || echo "")
    
    if [ -n "${CLOUDFRONT_URL}" ]; then
        log_success "CloudFront URL: ${CLOUDFRONT_URL}"
        log_info "ブラウザでアクセス可能になりました"
    else
        log_warning "CloudFront URLを取得できませんでした"
    fi
    
    if [ -n "${DISTRIBUTION_ID}" ]; then
        log_info "Distribution ID: ${DISTRIBUTION_ID}"
        log_info "キャッシュ無効化が必要な場合は 'make invalidate-cloudfront-cache' を実行してください"
    fi
    
    cd - > /dev/null
}

# コスト情報の表示
show_cost_info() {
    log_info "S3ストレージ使用量:"
    
    # S3ストレージ使用量
    STORAGE_SIZE=$(aws s3 ls "s3://${BUCKET_NAME}/" --recursive --summarize | grep "Total Size" | awk '{print $3}')
    if [ -n "${STORAGE_SIZE}" ]; then
        STORAGE_MB=$((STORAGE_SIZE / 1024 / 1024))
        log_info "  ストレージ使用量: ${STORAGE_MB} MB"
        
        # 概算コスト計算（東京リージョン）
        if [ "${STORAGE_MB}" -gt 0 ]; then
            MONTHLY_COST=$(echo "scale=4; ${STORAGE_MB} * 0.025 / 1024" | bc -l 2>/dev/null || echo "計算不可")
            log_info "  概算月額コスト: \$${MONTHLY_COST} (ストレージのみ)"
        fi
    fi
}

# メイン処理
main() {
    log_info "=== フロントエンドS3アップロード開始 ==="
    log_info "開始時刻: $(date)"
    log_info "環境: ${ENVIRONMENT}"
    
    # 処理実行
    load_environment
    check_prerequisites
    get_s3_bucket_name
    calculate_upload_size
    upload_to_s3
    verify_upload
    get_cloudfront_info
    show_cost_info
    
    log_success "=== フロントエンドS3アップロード完了 ==="
    log_info "完了時刻: $(date)"
    log_info "ログファイル: ${LOG_FILE}"
    
    # 次のステップの提案
    echo
    log_info "=== 次のステップ ==="
    log_info "1. CloudFrontキャッシュ無効化: make invalidate-cloudfront-cache"
    log_info "2. 動作確認: make verify-frontend-health"
    log_info "3. ブラウザでアクセス: ${CLOUDFRONT_URL:-'(URL取得失敗)'}"
}

# エラーハンドリング
trap 'log_error "スクリプト実行中にエラーが発生しました"; exit 1' ERR

# 引数チェック
if [ "$#" -gt 1 ]; then
    echo -e "${RED}使用方法: $0 [environment]${NC}"
    echo "environment: development (デフォルト), staging, production"
    exit 1
fi

# スクリプト実行
main "$@"