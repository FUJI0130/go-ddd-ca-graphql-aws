#!/bin/bash
# ===================================================================
# ファイル名: wait-for-certificate.sh
# 説明: ACM証明書の検証完了を待機するスクリプト
# 
# 用途:
#  - ACM証明書の検証状態をポーリングで確認
#  - 検証状態が「ISSUED」になるまで待機
#  - タイムアウト時間を超えた場合はエラー終了
# 
# 使用方法:
#  ./wait-for-certificate.sh [環境名] [最大待機時間(分)]
#
# 引数:
#  環境名 - 検証する環境（development, production）、省略時はdevelopment
#  最大待機時間 - 待機する最大時間（分）、省略時は30分
# ===================================================================

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 引数の解析
ENVIRONMENT=${1:-development}
MAX_WAIT_TIME=${2:-30}  # デフォルト30分
POLL_INTERVAL=60        # ポーリング間隔（秒）

echo -e "${BLUE}ACM証明書の検証完了を待機しています (環境: ${ENVIRONMENT})...${NC}"
echo -e "${BLUE}最大待機時間: ${MAX_WAIT_TIME}分${NC}"

# AWS認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  echo -e "${RED}エラー: AWS認証情報が設定されていないか、無効です${NC}"
  echo "AWS CLIの設定を確認してください: aws configure"
  exit 1
fi

# 証明書ARNの取得（ドメイン名は環境に応じて調整）
CERT_ARN=$(aws acm list-certificates --query "CertificateSummaryList[?contains(DomainName, 'grpc')].CertificateArn" --output text)

if [ -z "${CERT_ARN}" ] || [ "${CERT_ARN}" == "None" ]; then
  echo -e "${RED}エラー: 証明書が見つかりません${NC}"
  exit 1
fi

echo -e "${BLUE}監視する証明書ARN: ${CERT_ARN}${NC}"

# 待機開始時間の記録
START_TIME=$(date +%s)
END_TIME=$((START_TIME + MAX_WAIT_TIME * 60))

# 証明書の検証状態を監視
while true; do
  CURRENT_TIME=$(date +%s)
  
  # タイムアウトチェック
  if [ ${CURRENT_TIME} -gt ${END_TIME} ]; then
    echo -e "${RED}エラー: 最大待機時間(${MAX_WAIT_TIME}分)を超えました${NC}"
    echo -e "${RED}証明書の検証が完了しませんでした${NC}"
    exit 1
  fi
  
  # 経過時間を表示
  ELAPSED_TIME=$(( (CURRENT_TIME - START_TIME) / 60 ))
  echo -e "${BLUE}経過時間: ${ELAPSED_TIME}/${MAX_WAIT_TIME}分${NC}"
  
  # 証明書の状態を確認
  STATUS=$(aws acm describe-certificate --certificate-arn ${CERT_ARN} --query 'Certificate.Status' --output text)
  
  echo -e "${BLUE}現在の証明書状態: ${STATUS}${NC}"
  
  # 状態に応じた処理
  case "${STATUS}" in
    ISSUED)
      echo -e "${GREEN}✓ 証明書の検証が完了しました${NC}"
      exit 0
      ;;
    PENDING_VALIDATION)
      echo -e "${YELLOW}証明書はまだ検証中です. ${POLL_INTERVAL}秒後に再確認します...${NC}"
      sleep ${POLL_INTERVAL}
      ;;
    FAILED|INACTIVE|EXPIRED|VALIDATION_TIMED_OUT)
      echo -e "${RED}エラー: 証明書の検証に失敗しました (状態: ${STATUS})${NC}"
      exit 1
      ;;
    *)
      echo -e "${YELLOW}未知の状態: ${STATUS}. ${POLL_INTERVAL}秒後に再確認します...${NC}"
      sleep ${POLL_INTERVAL}
      ;;
  esac
done