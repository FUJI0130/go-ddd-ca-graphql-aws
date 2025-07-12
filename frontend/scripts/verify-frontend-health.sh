#!/bin/bash
# frontend/scripts/verify-frontend-health.sh
# フロントエンド動作検証スクリプト（バックエンドパターン踏襲）

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
TF_DIR="${PROJECT_ROOT}/deployments/terraform/environments/${ENVIRONMENT}"

# ログファイル設定
LOG_FILE="${PROJECT_ROOT}/logs/verify-frontend-$(date +%Y%m%d_%H%M%S).log"
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

# テスト結果記録用変数
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNINGS=0

# テスト結果記録関数
record_test_result() {
    local test_name="$1"
    local result="$2"
    local message="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    case "$result" in
        "PASS")
            PASSED_TESTS=$((PASSED_TESTS + 1))
            log_success "✓ ${test_name}: ${message}"
            ;;
        "FAIL")
            FAILED_TESTS=$((FAILED_TESTS + 1))
            log_error "✗ ${test_name}: ${message}"
            ;;
        "WARN")
            WARNINGS=$((WARNINGS + 1))
            log_warning "⚠ ${test_name}: ${message}"
            ;;
    esac
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
    
    # 必要コマンドの確認
    for cmd in curl aws terraform jq; do
        if command -v "$cmd" &> /dev/null; then
            record_test_result "前提条件" "PASS" "$cmd コマンドが利用可能"
        else
            record_test_result "前提条件" "FAIL" "$cmd コマンドが見つかりません"
        fi
    done
    
    # AWS認証確認
    if aws sts get-caller-identity &> /dev/null; then
        record_test_result "AWS認証" "PASS" "AWS認証情報が有効"
    else
        record_test_result "AWS認証" "FAIL" "AWS認証情報が無効"
    fi
}

# CloudFront/S3情報の取得
get_deployment_info() {
    log_info "デプロイ情報を取得中..."
    
    cd "${TF_DIR}"
    
    if [ ! -f ".terraform/terraform.tfstate" ] && [ ! -f "terraform.tfstate" ]; then
        record_test_result "Terraform状態" "FAIL" "Terraform状態ファイルが見つかりません"
        return 1
    fi
    
    # CloudFront情報取得
    CLOUDFRONT_URL=$(terraform output -raw frontend_cloudfront_url 2>/dev/null || echo "")
    CLOUDFRONT_DOMAIN=$(terraform output -raw frontend_cloudfront_domain_name 2>/dev/null || echo "")
    DISTRIBUTION_ID=$(terraform output -raw frontend_cloudfront_distribution_id 2>/dev/null || echo "")
    
    # S3情報取得
    S3_BUCKET_NAME=$(terraform output -raw frontend_s3_bucket_name 2>/dev/null || echo "")
    
    # GraphQL API情報取得（バックエンド連携用）
    GRAPHQL_ALB_DNS=$(terraform output -raw graphql_alb_dns_name 2>/dev/null || echo "")
    
    cd - > /dev/null
    
    # 取得結果の確認
    if [ -n "${CLOUDFRONT_URL}" ]; then
        record_test_result "CloudFront URL" "PASS" "${CLOUDFRONT_URL}"
    else
        record_test_result "CloudFront URL" "FAIL" "CloudFront URLが取得できません"
    fi
    
    if [ -n "${S3_BUCKET_NAME}" ]; then
        record_test_result "S3バケット" "PASS" "${S3_BUCKET_NAME}"
    else
        record_test_result "S3バケット" "FAIL" "S3バケット名が取得できません"
    fi
    
    if [ -n "${GRAPHQL_ALB_DNS}" ]; then
        record_test_result "GraphQL API" "PASS" "https://${GRAPHQL_ALB_DNS}/query"
    else
        record_test_result "GraphQL API" "WARN" "GraphQL API情報が取得できません（バックエンド未デプロイの可能性）"
    fi
}

# S3バケットの健全性確認
verify_s3_health() {
    log_info "S3バケットの健全性を確認中..."
    
    if [ -z "${S3_BUCKET_NAME}" ]; then
        record_test_result "S3存在確認" "FAIL" "S3バケット名が不明"
        return 1
    fi
    
    # バケット存在確認
    if aws s3 ls "s3://${S3_BUCKET_NAME}/" > /dev/null 2>&1; then
        record_test_result "S3存在確認" "PASS" "S3バケットにアクセス可能"
    else
        record_test_result "S3存在確認" "FAIL" "S3バケットにアクセスできません"
        return 1
    fi
    
    # 主要ファイル存在確認
    if aws s3 ls "s3://${S3_BUCKET_NAME}/index.html" > /dev/null 2>&1; then
        record_test_result "index.html" "PASS" "index.htmlが存在"
    else
        record_test_result "index.html" "FAIL" "index.htmlが見つかりません"
    fi
    
    # assetsディレクトリ確認
    if aws s3 ls "s3://${S3_BUCKET_NAME}/assets/" > /dev/null 2>&1; then
        ASSETS_COUNT=$(aws s3 ls "s3://${S3_BUCKET_NAME}/assets/" --recursive | wc -l)
        record_test_result "Assetsファイル" "PASS" "${ASSETS_COUNT}個のアセットファイル"
    else
        record_test_result "Assetsファイル" "WARN" "assetsディレクトリが見つかりません"
    fi
    
    # バケット使用量確認
    STORAGE_SIZE=$(aws s3 ls "s3://${S3_BUCKET_NAME}/" --recursive --summarize | grep "Total Size" | awk '{print $3}' || echo "0")
    if [ "${STORAGE_SIZE}" -gt 0 ]; then
        STORAGE_MB=$((STORAGE_SIZE / 1024 / 1024))
        record_test_result "S3使用量" "PASS" "${STORAGE_MB} MB"
    else
        record_test_result "S3使用量" "WARN" "ストレージ使用量が0 MB"
    fi
}

# CloudFrontの健全性確認
verify_cloudfront_health() {
    log_info "CloudFrontの健全性を確認中..."
    
    if [ -z "${DISTRIBUTION_ID}" ]; then
        record_test_result "CloudFront情報" "FAIL" "Distribution IDが不明"
        return 1
    fi
    
    # Distribution状態確認
    DISTRIBUTION_STATUS=$(aws cloudfront get-distribution --id "${DISTRIBUTION_ID}" \
        --query 'Distribution.Status' --output text 2>/dev/null || echo "不明")
    
    case "${DISTRIBUTION_STATUS}" in
        "Deployed")
            record_test_result "Distribution状態" "PASS" "デプロイ済み"
            ;;
        "InProgress")
            record_test_result "Distribution状態" "WARN" "デプロイ進行中"
            ;;
        *)
            record_test_result "Distribution状態" "FAIL" "状態: ${DISTRIBUTION_STATUS}"
            ;;
    esac
    
    # Distribution設定確認
    DISTRIBUTION_INFO=$(aws cloudfront get-distribution --id "${DISTRIBUTION_ID}" \
        --query 'Distribution.DistributionConfig' --output json 2>/dev/null || echo "{}")
    
    # オリジン設定確認
    ORIGIN_COUNT=$(echo "${DISTRIBUTION_INFO}" | jq '.Origins.Quantity' 2>/dev/null || echo "0")
    if [ "${ORIGIN_COUNT}" -gt 0 ]; then
        record_test_result "オリジン設定" "PASS" "${ORIGIN_COUNT}個のオリジン設定"
    else
        record_test_result "オリジン設定" "FAIL" "オリジン設定が見つかりません"
    fi
    
    # キャッシュビヘイビア確認
    BEHAVIOR_COUNT=$(echo "${DISTRIBUTION_INFO}" | jq '.CacheBehaviors.Quantity' 2>/dev/null || echo "0")
    DEFAULT_BEHAVIOR=$(echo "${DISTRIBUTION_INFO}" | jq -r '.DefaultCacheBehavior.TargetOriginId' 2>/dev/null || echo "不明")
    record_test_result "キャッシュビヘイビア" "PASS" "デフォルト+${BEHAVIOR_COUNT}個のカスタム設定"
}

# HTTP/HTTPSアクセステスト
verify_http_access() {
    log_info "HTTP/HTTPSアクセステストを実行中..."
    
    if [ -z "${CLOUDFRONT_URL}" ]; then
        record_test_result "HTTP接続テスト" "FAIL" "CloudFront URLが不明"
        return 1
    fi
    
    # HTTPSアクセステスト
    log_info "HTTPSアクセステスト実行中..."
    HTTPS_RESPONSE=$(curl -s -w "%{http_code}|%{time_total}|%{size_download}" \
        -o /dev/null "${CLOUDFRONT_URL}" 2>/dev/null || echo "000|0|0")
    
    HTTP_CODE=$(echo "${HTTPS_RESPONSE}" | cut -d'|' -f1)
    RESPONSE_TIME=$(echo "${HTTPS_RESPONSE}" | cut -d'|' -f2)
    CONTENT_SIZE=$(echo "${HTTPS_RESPONSE}" | cut -d'|' -f3)
    
    case "${HTTP_CODE}" in
        "200")
            record_test_result "HTTPS接続" "PASS" "応答時間: ${RESPONSE_TIME}s, サイズ: ${CONTENT_SIZE}bytes"
            ;;
        "403"|"404")
            record_test_result "HTTPS接続" "FAIL" "HTTPエラー: ${HTTP_CODE}"
            ;;
        "000")
            record_test_result "HTTPS接続" "FAIL" "接続失敗（タイムアウトまたはDNSエラー）"
            ;;
        *)
            record_test_result "HTTPS接続" "WARN" "予期しない応答: ${HTTP_CODE}"
            ;;
    esac
    
    # レスポンスヘッダー確認
    log_info "レスポンスヘッダー確認中..."
    HEADERS=$(curl -s -I "${CLOUDFRONT_URL}" 2>/dev/null || echo "")
    
    # セキュリティヘッダー確認
    if echo "${HEADERS}" | grep -i "x-content-type-options" > /dev/null; then
        record_test_result "セキュリティヘッダー" "PASS" "X-Content-Type-Options設定済み"
    else
        record_test_result "セキュリティヘッダー" "WARN" "X-Content-Type-Optionsが設定されていません"
    fi
    
    # CloudFrontヘッダー確認
    if echo "${HEADERS}" | grep -i "x-cache" > /dev/null; then
        CACHE_STATUS=$(echo "${HEADERS}" | grep -i "x-cache" | cut -d':' -f2 | tr -d ' \r\n')
        record_test_result "CloudFrontキャッシュ" "PASS" "ヘッダー: ${CACHE_STATUS}"
    else
        record_test_result "CloudFrontキャッシュ" "WARN" "CloudFrontヘッダーが見つかりません"
    fi
    
    # Content-Type確認
    if echo "${HEADERS}" | grep -i "content-type.*text/html" > /dev/null; then
        record_test_result "Content-Type" "PASS" "HTMLとして配信"
    else
        record_test_result "Content-Type" "WARN" "Content-Typeが正しく設定されていない可能性"
    fi
}

# フロントエンドコンテンツ確認
verify_frontend_content() {
    log_info "フロントエンドコンテンツ確認中..."
    
    if [ -z "${CLOUDFRONT_URL}" ]; then
        record_test_result "コンテンツ取得" "FAIL" "CloudFront URLが不明"
        return 1
    fi
    
    # HTMLコンテンツ取得
    HTML_CONTENT=$(curl -s "${CLOUDFRONT_URL}" 2>/dev/null || echo "")
    
    if [ -z "${HTML_CONTENT}" ]; then
        record_test_result "HTMLコンテンツ" "FAIL" "HTMLコンテンツを取得できません"
        return 1
    fi
    
    # 基本的なHTML構造確認
    if echo "${HTML_CONTENT}" | grep -i "<html" > /dev/null; then
        record_test_result "HTML構造" "PASS" "有効なHTML構造"
    else
        record_test_result "HTML構造" "FAIL" "HTMLタグが見つかりません"
    fi
    
    # React関連確認
    if echo "${HTML_CONTENT}" | grep -i "react" > /dev/null; then
        record_test_result "React検出" "PASS" "Reactアプリケーション"
    else
        record_test_result "React検出" "WARN" "Reactの痕跡が見つかりません"
    fi
    
    # Vite関連確認
    if echo "${HTML_CONTENT}" | grep -i "vite" > /dev/null; then
        record_test_result "Vite検出" "PASS" "Viteビルド成果物"
    else
        record_test_result "Vite検出" "WARN" "Viteの痕跡が見つかりません"
    fi
    
    # CSS/JS読み込み確認
    CSS_COUNT=$(echo "${HTML_CONTENT}" | grep -o '<link.*\.css' | wc -l)
    JS_COUNT=$(echo "${HTML_CONTENT}" | grep -o '<script.*\.js' | wc -l)
    
    if [ "${CSS_COUNT}" -gt 0 ]; then
        record_test_result "CSSファイル" "PASS" "${CSS_COUNT}個のCSSファイル"
    else
        record_test_result "CSSファイル" "WARN" "CSSファイルが見つかりません"
    fi
    
    if [ "${JS_COUNT}" -gt 0 ]; then
        record_test_result "JavaScriptファイル" "PASS" "${JS_COUNT}個のJSファイル"
    else
        record_test_result "JavaScriptファイル" "WARN" "JavaScriptファイルが見つかりません"
    fi
    
    # アプリケーション名確認
    if echo "${HTML_CONTENT}" | grep -i "テスト管理システム" > /dev/null; then
        record_test_result "アプリケーション名" "PASS" "正しいアプリケーション名"
    else
        record_test_result "アプリケーション名" "WARN" "アプリケーション名が見つかりません"
    fi
}

# バックエンドAPI連携テスト
verify_backend_connectivity() {
    log_info "バックエンドAPI連携テスト実行中..."
    
    if [ -z "${GRAPHQL_ALB_DNS}" ]; then
        record_test_result "バックエンド情報" "WARN" "GraphQL API情報が不明（バックエンド未デプロイの可能性）"
        return 0
    fi
    
    GRAPHQL_URL="https://${GRAPHQL_ALB_DNS}/query"
    
    # GraphQL エンドポイント疎通確認
    log_info "GraphQLエンドポイント疎通確認中..."
    GRAPHQL_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null \
        -X POST "${GRAPHQL_URL}" \
        -H "Content-Type: application/json" \
        -d '{"query":"query{__typename}"}' 2>/dev/null || echo "000")
    
    case "${GRAPHQL_RESPONSE}" in
        "200")
            record_test_result "GraphQL疎通" "PASS" "GraphQLエンドポイントが応答"
            ;;
        "400"|"401"|"403")
            record_test_result "GraphQL疎通" "PASS" "GraphQLエンドポイントが存在（認証エラーは正常）"
            ;;
        "404")
            record_test_result "GraphQL疎通" "FAIL" "GraphQLエンドポイントが見つかりません"
            ;;
        "000")
            record_test_result "GraphQL疎通" "FAIL" "GraphQLエンドポイントに接続できません"
            ;;
        *)
            record_test_result "GraphQL疎通" "WARN" "予期しない応答: ${GRAPHQL_RESPONSE}"
            ;;
    esac
    
    # CORS設定確認
    log_info "CORS設定確認中..."
    CORS_RESPONSE=$(curl -s -I \
        -H "Origin: ${CLOUDFRONT_URL}" \
        -H "Access-Control-Request-Method: POST" \
        -H "Access-Control-Request-Headers: Content-Type" \
        -X OPTIONS "${GRAPHQL_URL}" 2>/dev/null || echo "")
    
    if echo "${CORS_RESPONSE}" | grep -i "access-control-allow-origin" > /dev/null; then
        record_test_result "CORS設定" "PASS" "CORSヘッダーが設定済み"
    else
        record_test_result "CORS設定" "WARN" "CORSヘッダーが見つかりません"
    fi
}

# パフォーマンステスト
verify_performance() {
    log_info "パフォーマンステスト実行中..."
    
    if [ -z "${CLOUDFRONT_URL}" ]; then
        record_test_result "パフォーマンステスト" "FAIL" "CloudFront URLが不明"
        return 1
    fi
    
    # 複数回アクセスして平均応答時間を測定
    log_info "応答時間測定中（5回実行）..."
    
    total_time=0
    successful_requests=0
    
    for i in {1..5}; do
        response_time=$(curl -s -w "%{time_total}" -o /dev/null "${CLOUDFRONT_URL}" 2>/dev/null || echo "0")
        if [ "$(echo "$response_time > 0" | bc -l 2>/dev/null || echo "0")" = "1" ]; then
            total_time=$(echo "$total_time + $response_time" | bc -l 2>/dev/null || echo "$total_time")
            successful_requests=$((successful_requests + 1))
        fi
        sleep 1
    done
    
    if [ $successful_requests -gt 0 ]; then
        avg_time=$(echo "scale=3; $total_time / $successful_requests" | bc -l 2>/dev/null || echo "0")
        
        if [ "$(echo "$avg_time < 2.0" | bc -l 2>/dev/null || echo "0")" = "1" ]; then
            record_test_result "平均応答時間" "PASS" "${avg_time}秒（良好）"
        elif [ "$(echo "$avg_time < 5.0" | bc -l 2>/dev/null || echo "0")" = "1" ]; then
            record_test_result "平均応答時間" "WARN" "${avg_time}秒（改善の余地あり）"
        else
            record_test_result "平均応答時間" "FAIL" "${avg_time}秒（遅すぎます）"
        fi
    else
        record_test_result "平均応答時間" "FAIL" "応答時間測定失敗"
    fi
    
    # 圧縮効果確認
    log_info "圧縮効果確認中..."
    
    # 非圧縮サイズ（Accept-Encodingなし）
    uncompressed_size=$(curl -s -w "%{size_download}" -o /dev/null \
        -H "Accept-Encoding:" "${CLOUDFRONT_URL}" 2>/dev/null || echo "0")
    
    # 圧縮サイズ（gzip）
    compressed_size=$(curl -s -w "%{size_download}" -o /dev/null \
        -H "Accept-Encoding: gzip" "${CLOUDFRONT_URL}" 2>/dev/null || echo "0")
    
    if [ "$uncompressed_size" -gt 0 ] && [ "$compressed_size" -gt 0 ] && [ "$compressed_size" -lt "$uncompressed_size" ]; then
        compression_ratio=$(echo "scale=1; (($uncompressed_size - $compressed_size) * 100) / $uncompressed_size" | bc -l 2>/dev/null || echo "0")
        record_test_result "圧縮効果" "PASS" "${compression_ratio}%削減（${uncompressed_size}→${compressed_size}bytes）"
    else
        record_test_result "圧縮効果" "WARN" "圧縮効果を測定できませんでした"
    fi
}

# セキュリティチェック
verify_security() {
    log_info "セキュリティチェック実行中..."
    
    if [ -z "${CLOUDFRONT_URL}" ]; then
        record_test_result "セキュリティチェック" "FAIL" "CloudFront URLが不明"
        return 1
    fi
    
    # HTTPSリダイレクト確認
    HTTP_URL=$(echo "${CLOUDFRONT_URL}" | sed 's/https:/http:/')
    HTTP_REDIRECT=$(curl -s -w "%{http_code}" -o /dev/null "${HTTP_URL}" 2>/dev/null || echo "000")
    
    case "${HTTP_REDIRECT}" in
        "301"|"302"|"308")
            record_test_result "HTTPS リダイレクト" "PASS" "HTTPからHTTPSにリダイレクト"
            ;;
        "200")
            record_test_result "HTTPS リダイレクト" "WARN" "HTTPでもアクセス可能（要確認）"
            ;;
        *)
            record_test_result "HTTPS リダイレクト" "WARN" "HTTPリダイレクト確認不可"
            ;;
    esac
    
    # セキュリティヘッダー詳細確認
    SECURITY_HEADERS=$(curl -s -I "${CLOUDFRONT_URL}" 2>/dev/null || echo "")
    
    # 各セキュリティヘッダーをチェック
    for header in "X-Content-Type-Options" "X-Frame-Options" "Strict-Transport-Security" "Content-Security-Policy"; do
        if echo "${SECURITY_HEADERS}" | grep -i "$header" > /dev/null; then
            record_test_result "セキュリティヘッダー" "PASS" "$header 設定済み"
        else
            record_test_result "セキュリティヘッダー" "WARN" "$header 未設定"
        fi
    done
    
    # 不要なサーバー情報の露出確認
    if echo "${SECURITY_HEADERS}" | grep -i "server.*apache\|nginx\|iis" > /dev/null; then
        record_test_result "情報露出" "WARN" "サーバー情報が露出している可能性"
    else
        record_test_result "情報露出" "PASS" "サーバー情報の露出なし"
    fi
}

# 総合結果レポート
generate_summary_report() {
    log_info "=== 検証結果サマリー ==="
    
    # 結果統計
    log_info "実行統計:"
    log_info "  総テスト数: ${TOTAL_TESTS}"
    log_success "  成功: ${PASSED_TESTS}"
    log_error "  失敗: ${FAILED_TESTS}"
    log_warning "  警告: ${WARNINGS}"
    
    # 成功率計算
    if [ $TOTAL_TESTS -gt 0 ]; then
        SUCCESS_RATE=$(echo "scale=1; ($PASSED_TESTS * 100) / $TOTAL_TESTS" | bc -l 2>/dev/null || echo "0")
        log_info "  成功率: ${SUCCESS_RATE}%"
    fi
    
    # 総合判定
    echo
    if [ $FAILED_TESTS -eq 0 ]; then
        if [ $WARNINGS -eq 0 ]; then
            log_success "🎉 総合判定: 完全成功 - フロントエンドは正常に動作しています"
        else
            log_warning "⚠️  総合判定: 成功（警告あり） - 基本機能は動作していますが改善点があります"
        fi
    else
        log_error "❌ 総合判定: 失敗 - 重要な問題が発見されました"
    fi
    
    # アクセス情報
    echo
    log_info "=== アクセス情報 ==="
    if [ -n "${CLOUDFRONT_URL}" ]; then
        log_info "フロントエンドURL: ${CLOUDFRONT_URL}"
    fi
    if [ -n "${GRAPHQL_ALB_DNS}" ]; then
        log_info "GraphQL API: https://${GRAPHQL_ALB_DNS}/query"
    fi
    
    # 次のステップ提案
    echo
    log_info "=== 推奨アクション ==="
    
    if [ $FAILED_TESTS -gt 0 ]; then
        log_error "重要: 失敗したテストを確認し、問題を解決してください"
        log_info "1. AWS環境のデプロイ状態を確認"
        log_info "2. Terraformエラーがないか確認"
        log_info "3. ビルドが正しく完了しているか確認"
    fi
    
    if [ $WARNINGS -gt 0 ]; then
        log_warning "推奨: 警告項目を確認し、可能であれば改善してください"
        log_info "1. セキュリティヘッダーの追加設定"
        log_info "2. パフォーマンス最適化"
        log_info "3. CORS設定の確認"
    fi
    
    if [ $FAILED_TESTS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
        log_success "フロントエンドは本番レディです！"
        log_info "1. ブラウザでアクセスして最終確認"
        log_info "2. 各種機能の手動テスト"
        log_info "3. 必要に応じてDNS設定の追加"
    fi
}

# メイン処理
main() {
    log_info "=== フロントエンド動作検証開始 ==="
    log_info "開始時刻: $(date)"
    log_info "環境: ${ENVIRONMENT}"
    
    # 各検証項目の実行
    load_environment
    check_prerequisites
    get_deployment_info
    verify_s3_health
    verify_cloudfront_health
    verify_http_access
    verify_frontend_content
    verify_backend_connectivity
    verify_performance
    verify_security
    
    # 総合レポート生成
    generate_summary_report
    
    log_info "完了時刻: $(date)"
    log_info "ログファイル: ${LOG_FILE}"
    
    # 終了コード設定
    if [ $FAILED_TESTS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# エラーハンドリング
trap 'log_error "スクリプト実行中にエラーが発生しました"; exit 1' ERR

# 引数チェック
if [ "$#" -gt 1 ]; then
    echo -e "${RED}使用方法: $0 [environment]${NC}"
    echo "environment: development (デフォルト), staging, production"
    echo
    echo "このスクリプトは以下の項目を検証します:"
    echo "  ✓ AWS環境とTerraform状態"
    echo "  ✓ S3バケットとファイル"
    echo "  ✓ CloudFront配信"
    echo "  ✓ HTTP/HTTPS接続"
    echo "  ✓ フロントエンドコンテンツ"
    echo "  ✓ バックエンドAPI連携"
    echo "  ✓ パフォーマンス測定"
    echo "  ✓ セキュリティチェック"
    exit 1
fi

# スクリプト実行
main "$@"