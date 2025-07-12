#!/bin/bash
# scripts/terraform/update-tfvars.sh - 新しいモジュール構造に対応したterraform.tfvars更新スクリプト
# 使用方法: update-tfvars.sh [environment] [service_type]
# 例: update-tfvars.sh development
#     update-tfvars.sh production api

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 環境引数の解析
TF_ENV=${1:-development}
SERVICE_TYPE=${2:-""}  # 指定がない場合は全サービスを更新

# 設定
TFVARS_FILE="deployments/terraform/environments/${TF_ENV}/terraform.tfvars"
AWS_REGION=${AWS_REGION:-"ap-northeast-1"}

# 全サービスタイプ
ALL_SERVICE_TYPES=("api" "graphql" "grpc")

# スクリプトの使用法を表示
usage() {
  echo -e "${BLUE}使用法:${NC} $0 [environment] [service_type]"
  echo "  environment: 環境名 (デフォルト: development)"
  echo "  service_type: サービスタイプ (省略時は全サービス更新)"
  echo -e "${BLUE}例:${NC}"
  echo "  $0 production              # 本番環境の全サービス更新"
  echo "  $0 development api         # 開発環境のAPIサービスのみ更新"
  echo "  SERVICE_TYPE=graphql $0    # 環境変数でサービスタイプ指定"
  exit 1
}

# エラーメッセージを表示して終了
error_exit() {
  echo -e "${RED}エラー:${NC} $1" >&2
  exit 1
}

# ファイルの存在チェック
if [ ! -f "${TFVARS_FILE}" ]; then
  error_exit "terraform.tfvarsファイルが見つかりません: ${TFVARS_FILE}"
fi

# AWSコマンドの存在チェック
if ! command -v aws &> /dev/null; then
  error_exit "AWS CLIがインストールされていません。AWS CLIをインストールしてください。"
fi

# AWS認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  error_exit "AWS認証情報が無効または設定されていません。AWS認証情報を確認してください。"
fi

# バックアップの作成
cp "${TFVARS_FILE}" "${TFVARS_FILE}.bak"
echo -e "${BLUE}terraform.tfvarsのバックアップを作成しました:${NC} ${TFVARS_FILE}.bak"

# 処理するサービスを決定
if [ -n "${SERVICE_TYPE}" ]; then
  # 単一サービスの処理
  if [[ ! " ${ALL_SERVICE_TYPES[*]} " =~ " ${SERVICE_TYPE} " ]]; then
    error_exit "無効なサービスタイプです: ${SERVICE_TYPE}。有効なタイプ: ${ALL_SERVICE_TYPES[*]}"
  fi
  SERVICES_TO_UPDATE=("${SERVICE_TYPE}")
  echo -e "${BLUE}${TF_ENV}環境の${SERVICE_TYPE}サービスのイメージURIを更新します${NC}"
else
  # 全サービスの処理
  SERVICES_TO_UPDATE=("${ALL_SERVICE_TYPES[@]}")
  echo -e "${BLUE}${TF_ENV}環境の全サービスのイメージURIを更新します${NC}"
fi

# サービスタイプごとにECRイメージURIを取得して更新
update_service_image_uri() {
  local service_type=$1
  local var_name="${service_type}_image"
  
  echo -e "${BLUE}サービス ${service_type} のイメージURIを取得しています...${NC}"
  
  # ECRリポジトリ名を構築
  local ECR_REPO_NAME="${TF_ENV}-test-management-${service_type}"
  
  # リポジトリの存在確認
  if ! aws ecr describe-repositories --repository-names ${ECR_REPO_NAME} --region ${AWS_REGION} &>/dev/null; then
    echo -e "${YELLOW}警告:${NC} ECRリポジトリ ${ECR_REPO_NAME} が見つかりません。スキップします。"
    return
  fi
  
  # リポジトリURIを取得
  local ECR_REPO_URI=$(aws ecr describe-repositories --repository-names ${ECR_REPO_NAME} --query "repositories[0].repositoryUri" --output text --region ${AWS_REGION})
  
  if [ -z "$ECR_REPO_URI" ]; then
    echo -e "${YELLOW}警告:${NC} ECRリポジトリ ${ECR_REPO_NAME} のURIを取得できませんでした。スキップします。"
    return
  fi
  
  local LATEST_IMAGE_URI="${ECR_REPO_URI}:latest"
  
  # terraform.tfvarsを更新
  if grep -q "${var_name}" ${TFVARS_FILE}; then
    # 変数が存在する場合は更新
    sed -i -e "s|${var_name}.*=.*|${var_name} = \"${LATEST_IMAGE_URI}\"|g" ${TFVARS_FILE}
    echo -e "${GREEN}✓${NC} ${var_name} を更新しました: ${LATEST_IMAGE_URI}"
  else
    # 変数が存在しない場合は追加
    echo -e "\n# ${service_type^} service image" >> ${TFVARS_FILE}
    echo "${var_name} = \"${LATEST_IMAGE_URI}\"" >> ${TFVARS_FILE}
    echo -e "${GREEN}✓${NC} ${var_name} を追加しました: ${LATEST_IMAGE_URI}"
  fi
}

# 各サービスを更新
for service_type in "${SERVICES_TO_UPDATE[@]}"; do
  update_service_image_uri "${service_type}"
done

echo -e "${GREEN}terraform.tfvarsの更新が完了しました。${NC}"
echo "更新前のファイルはバックアップとして保存されています: ${TFVARS_FILE}.bak"