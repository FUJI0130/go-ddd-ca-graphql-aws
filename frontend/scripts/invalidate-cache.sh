#!/bin/bash
# frontend/scripts/invalidate-cache.sh
# CloudFrontキャッシュ無効化スクリプト（バックエンドパターン踏襲）

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
LOG_FILE="${PROJECT_ROOT}/logs/invalidate-cache-$(date +%Y%m%d_%H%M%S).log"
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
    
    log_success "前提条件確認完了"
}

# CloudFront Distribution ID取得
get_distribution_id() {
    log_info "CloudFront Distribution IDを取得中..."
    
    cd "${TF_DIR}"
    
    if [ ! -f ".terraform/terraform.tfstate" ] && [ ! -f "terraform.tfstate" ]; then
        log_error "Terraform状態が見つかりません"
        log_error "先に 'make terraform-deploy-frontend' を実行してください"
        exit 1
    fi
    
    DISTRIBUTION_ID=$(terraform output -raw cloudfront_distribution_id 2>/dev/null || echo "")
    
    if [ -z "${DISTRIBUTION_ID}" ]; then
        log_error "CloudFront Distribution IDを取得できませんでした"
        log_error "Terraformデプロイが完了しているか確認してください"
        exit 1
    fi
    
    log_success "Distribution ID: ${DISTRIBUTION_ID}"
    cd - > /dev/null
}

# CloudFront情報の表示
show_cloudfront_info() {
    log_info "CloudFront分析情報を取得中..."
    
    cd "${TF_DIR}"
    
    # 追加情報取得
    CLOUDFRONT_URL=$(terraform output -raw frontend_cloudfront_url 2>/dev/null || echo "")
    CLOUDFRONT_DOMAIN=$(terraform output -raw frontend_cloudfront_domain_name 2>/dev/null || echo "")
    
    if [ -n "${CLOUDFRONT_URL}" ]; then
        log_info "CloudFront URL: ${CLOUDFRONT_URL}"
    fi
    
    if [ -n "${CLOUDFRONT_DOMAIN}" ]; then
        log_info "CloudFront ドメイン: ${CLOUDFRONT_DOMAIN}"
    fi
    
    # Distribution状態確認
    log_info "Distribution状態を確認中..."
    DISTRIBUTION_STATUS=$(aws cloudfront get-distribution --id "${DISTRIBUTION_ID}" \
        --query 'Distribution.Status' --output text 2>/dev/null || echo "不明")
    
    log_info "Distribution状態: ${DISTRIBUTION_STATUS}"
    
    if [ "${DISTRIBUTION_STATUS}" != "Deployed" ]; then
        log_warning "Distributionがまだデプロイ中の可能性があります"
        log_warning "無効化は実行可能ですが、完了まで時間がかかる場合があります"
    fi
    
    cd - > /dev/null
}

# 無効化パスの決定
determine_invalidation_paths() {
    log_info "無効化パスを決定中..."
    
    # デフォルトの無効化パス
    INVALIDATION_PATHS='["/*"]'
    
    # カスタム無効化パスの確認（オプション）
    if [ -n "${INVALIDATION_CUSTOM_PATHS:-}" ]; then
        log_info "カスタム無効化パスが指定されています: ${INVALIDATION_CUSTOM_PATHS}"
        INVALIDATION_PATHS="${INVALIDATION_CUSTOM_PATHS}"
    else
        log_info "デフォルト無効化パス: /* (全ファイル)"
    fi
    
    # 無効化パスの表示
    echo "${INVALIDATION_PATHS}" | python3 -m json.tool 2>/dev/null | while IFS= read -r line; do
        if [[ $line == *"\""* ]]; then
            path=$(echo "$line" | sed 's/.*"\(.*\)".*/\1/')
            log_info "  対象パス: ${path}"
        fi
    done
}

# 既存の無効化チェック
check_existing_invalidations() {
    log_info "既存の無効化状況を確認中..."
    
    # 進行中の無効化を取得
    EXISTING_INVALIDATIONS=$(aws cloudfront list-invalidations \
        --distribution-id "${DISTRIBUTION_ID}" \
        --query 'InvalidationList.Items[?Status==`InProgress`].{Id:Id,Status:Status,CreateTime:CreateTime}' \
        --output table 2>/dev/null || echo "")
    
    if [ -n "${EXISTING_INVALIDATIONS}" ] && [ "${EXISTING_INVALIDATIONS}" != "[]" ]; then
        log_warning "進行中の無効化があります:"
        echo "${EXISTING_INVALIDATIONS}"
        log_warning "新しい無効化を追加で実行します"
    else
        log_info "進行中の無効化はありません"
    fi
}

# キャッシュ無効化の実行
execute_invalidation() {
    log_info "CloudFrontキャッシュ無効化を実行中..."
    
    # 無効化実行
    INVALIDATION_RESULT=$(aws cloudfront create-invalidation \
        --distribution-id "${DISTRIBUTION_ID}" \
        --paths "${INVALIDATION_PATHS}" \
        --output json 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        INVALIDATION_ID=$(echo "${INVALIDATION_RESULT}" | python3 -c "import sys, json; print(json.load(sys.stdin)['Invalidation']['Id'])" 2>/dev/null || echo "不明")
        INVALIDATION_STATUS=$(echo "${INVALIDATION_RESULT}" | python3 -c "import sys, json; print(json.load(sys.stdin)['Invalidation']['Status'])" 2>/dev/null || echo "不明")
        CREATE_TIME=$(echo "${INVALIDATION_RESULT}" | python3 -c "import sys, json; print(json.load(sys.stdin)['Invalidation']['CreateTime'])" 2>/dev/null || echo "不明")
        
        log_success "無効化リクエスト送信完了"
        log_success "  無効化ID: ${INVALIDATION_ID}"
        log_success "  状態: ${INVALIDATION_STATUS}"
        log_success "  開始時刻: ${CREATE_TIME}"
    else
        log_error "無効化リクエストの送信に失敗しました"
        exit 1
    fi
}

# 無効化進捗の監視
monitor_invalidation_progress() {
    log_info "無効化進捗を監視中..."
    log_info "（通常15-20分程度かかります）"
    
    # 進捗監視のオプション
    MONITOR_PROGRESS=${MONITOR_PROGRESS:-"false"}
    
    if [ "${MONITOR_PROGRESS}" = "true" ]; then
        log_info "リアルタイム進捗監視を開始..."
        
        local check_count=0
        local max_checks=60  # 最大30分（30秒間隔×60回）
        
        while [ $check_count -lt $max_checks ]; do
            sleep 30
            check_count=$((check_count + 1))
            
            CURRENT_STATUS=$(aws cloudfront get-invalidation \
                --distribution-id "${DISTRIBUTION_ID}" \
                --id "${INVALIDATION_ID}" \
                --query 'Invalidation.Status' \
                --output text 2>/dev/null || echo "不明")
            
            case "${CURRENT_STATUS}" in
                "Completed")
                    log_success "無効化完了しました！ (${check_count}/2分経過)"
                    return 0
                    ;;
                "InProgress")
                    log_info "無効化進行中... (${check_count}/2分経過)"
                    ;;
                *)
                    log_warning "無効化状態: ${CURRENT_STATUS} (${check_count}/2分経過)"
                    ;;
            esac
        done
        
        log_warning "監視タイムアウト。手動で状態を確認してください"
    else
        log_info "進捗監視はスキップされました"
        log_info "状態確認: aws cloudfront get-invalidation --distribution-id ${DISTRIBUTION_ID} --id ${INVALIDATION_ID}"
    fi
}

# 無効化状況の最終確認
verify_invalidation_status() {
    log_info "最新の無効化状況を確認中..."
    
    # 最新の無効化状況取得
    LATEST_INVALIDATIONS=$(aws cloudfront list-invalidations \
        --distribution-id "${DISTRIBUTION_ID}" \
        --max-items 3 \
        --query 'InvalidationList.Items[].{Id:Id,Status:Status,CreateTime:CreateTime}' \
        --output table 2>/dev/null || echo "取得失敗")
    
    if [ "${LATEST_INVALIDATIONS}" != "取得失敗" ]; then
        log_info "最近の無効化履歴:"
        echo "${LATEST_INVALIDATIONS}"
    else
        log_warning "無効化履歴の取得に失敗しました"
    fi
}

# コスト情報の表示
show_cost_info() {
    log_info "CloudFront無効化コスト情報:"
    log_info "  無効化リクエスト: 1回 = \$0.005 (月1000回まで無料)"
    log_info "  ※ 無料枠内であれば追加料金は発生しません"
    
    # 今月の無効化回数を推定（概算）
    CURRENT_MONTH=$(date +%Y-%m)
    log_info "  参考: 頻繁な無効化はコストに影響する可能性があります"
}

# 最適化提案
suggest_optimizations() {
    log_info "=== 最適化提案 ==="
    log_info "1. 頻繁な無効化を避けるため、バージョニングを活用"
    log_info "2. 静的ファイルには長期キャッシュを設定"
    log_info "3. index.htmlのみ短期キャッシュに設定済み"
    log_info "4. 大きな変更時のみ全体無効化（/*）を実行"
    log_info "5. 部分的な変更時は特定パスのみ無効化を検討"
}

# メイン処理
main() {
    log_info "=== CloudFrontキャッシュ無効化開始 ==="
    log_info "開始時刻: $(date)"
    log_info "環境: ${ENVIRONMENT}"
    
    # 処理実行
    load_environment
    check_prerequisites
    get_distribution_id
    show_cloudfront_info
    determine_invalidation_paths
    check_existing_invalidations
    execute_invalidation
    monitor_invalidation_progress
    verify_invalidation_status
    show_cost_info
    suggest_optimizations
    
    log_success "=== CloudFrontキャッシュ無効化処理完了 ==="
    log_info "完了時刻: $(date)"
    log_info "ログファイル: ${LOG_FILE}"
    
    # 次のステップの提案
    echo
    log_info "=== 次のステップ ==="
    log_info "1. 動作確認: make verify-frontend-health"
    log_info "2. ブラウザでアクセスして確認"
    log_info "3. 無効化完了確認: aws cloudfront get-invalidation --distribution-id ${DISTRIBUTION_ID} --id ${INVALIDATION_ID:-'(ID不明)'}"
}

# エラーハンドリング
trap 'log_error "スクリプト実行中にエラーが発生しました"; exit 1' ERR

# 引数チェック
if [ "$#" -gt 1 ]; then
    echo -e "${RED}使用方法: $0 [environment]${NC}"
    echo "environment: development (デフォルト), staging, production"
    echo
    echo "オプション環境変数:"
    echo "  MONITOR_PROGRESS=true     - リアルタイム進捗監視を有効化"
    echo "  INVALIDATION_CUSTOM_PATHS - カスタム無効化パス (JSON配列形式)"
    echo
    echo "例:"
    echo "  $0 development"
    echo "  MONITOR_PROGRESS=true $0 development"
    echo "  INVALIDATION_CUSTOM_PATHS=\'["index.html", "/assets/*"]\' $0 development"
    exit 1
fi

# スクリプト実行
main "$@"