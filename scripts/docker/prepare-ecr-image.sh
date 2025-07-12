#!/bin/bash
# scripts/docker/prepare-ecr-image.sh - ECRイメージ準備スクリプト（改善版）
# 使用方法: prepare-ecr-image.sh [service_type] [environment]
# 例: prepare-ecr-image.sh api development
#     prepare-ecr-image.sh graphql production

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 引数の解析
SERVICE_TYPE=${1:-api}  # デフォルトはapi
ENVIRONMENT=${2:-development}  # デフォルトはdevelopment
AWS_REGION=${AWS_REGION:-"ap-northeast-1"}

# 有効なサービスタイプの配列
VALID_SERVICE_TYPES=("api" "graphql" "grpc")
# 有効な環境の配列
VALID_ENVIRONMENTS=("development" "production")

# 使用方法を表示
usage() {
  echo -e "${BLUE}使用方法:${NC} $0 [service_type] [environment]"
  echo "  service_type: サービスタイプ (デフォルト: api)"
  echo "  environment: 環境名 (デフォルト: development)"
  echo -e "${BLUE}有効なサービスタイプ:${NC} ${VALID_SERVICE_TYPES[*]}"
  echo -e "${BLUE}有効な環境:${NC} ${VALID_ENVIRONMENTS[*]}"
  echo -e "${BLUE}例:${NC}"
  echo "  $0 api development    # 開発環境のAPIサービス用イメージ準備"
  echo "  $0 graphql production # 本番環境のGraphQLサービス用イメージ準備"
  exit 1
}

# エラーメッセージを表示して終了
error_exit() {
  echo -e "${RED}エラー:${NC} $1" >&2
  exit 1
}

# サービスタイプの検証
if [[ ! " ${VALID_SERVICE_TYPES[*]} " =~ " ${SERVICE_TYPE} " ]]; then
  error_exit "無効なサービスタイプです: ${SERVICE_TYPE}。有効なタイプ: ${VALID_SERVICE_TYPES[*]}"
fi

# 環境の検証
if [[ ! " ${VALID_ENVIRONMENTS[*]} " =~ " ${ENVIRONMENT} " ]]; then
  error_exit "無効な環境です: ${ENVIRONMENT}。有効な環境: ${VALID_ENVIRONMENTS[*]}"
fi

# ECRリポジトリ名の設定
ECR_REPO_NAME="${ENVIRONMENT}-test-management-${SERVICE_TYPE}"

echo -e "${BLUE}AWS ECRへのDockerイメージ準備を開始します${NC}"
echo "サービス: ${SERVICE_TYPE}"
echo "環境: ${ENVIRONMENT}"
echo "リポジトリ: ${ECR_REPO_NAME}"

# AWS認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  error_exit "AWS認証情報が無効または設定されていません。AWS認証情報を確認してください。"
fi

# ECRリポジトリの存在確認と作成
if ! aws ecr describe-repositories --repository-names ${ECR_REPO_NAME} --region ${AWS_REGION} &>/dev/null; then
  echo -e "${BLUE}ECRリポジトリを作成しています:${NC} ${ECR_REPO_NAME}..."
  aws ecr create-repository --repository-name ${ECR_REPO_NAME} --region ${AWS_REGION} || error_exit "リポジトリの作成に失敗しました"
  echo -e "${GREEN}✓ リポジトリが作成されました${NC}"
else
  echo -e "${GREEN}✓ リポジトリは既に存在します${NC}"
fi

# ECRログイン
echo -e "${BLUE}ECRにログインしています...${NC}"
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_ENDPOINT="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

if ! aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${ECR_ENDPOINT}; then
  error_exit "ECRログインに失敗しました"
fi
echo -e "${GREEN}✓ ECRログインに成功しました${NC}"

# Dockerfileの存在確認
if [ ! -f "Dockerfile" ]; then
  error_exit "Dockerfileが見つかりません。ルートディレクトリで実行しているか確認してください。"
fi

# Dockerイメージビルド
echo -e "${BLUE}Dockerイメージをビルドしています...${NC}"
echo "ビルド引数: SERVICE_TYPE=${SERVICE_TYPE}"

if ! docker build -t ${ECR_REPO_NAME}:latest --build-arg SERVICE_TYPE=${SERVICE_TYPE} .; then
  error_exit "Dockerイメージのビルドに失敗しました"
fi
echo -e "${GREEN}✓ イメージのビルドに成功しました${NC}"

# ECRリポジトリへのタグ付けとプッシュ
ECR_REPO_URI=$(aws ecr describe-repositories --repository-names ${ECR_REPO_NAME} --query "repositories[0].repositoryUri" --output text --region ${AWS_REGION})

if [ -z "${ECR_REPO_URI}" ]; then
  error_exit "ECRリポジトリURIの取得に失敗しました"
fi

echo -e "${BLUE}イメージにタグを付けています:${NC} ${ECR_REPO_URI}:latest"
docker tag ${ECR_REPO_NAME}:latest ${ECR_REPO_URI}:latest || error_exit "タグ付けに失敗しました"

echo -e "${BLUE}イメージをプッシュしています:${NC} ${ECR_REPO_URI}:latest"
if ! docker push ${ECR_REPO_URI}:latest; then
  error_exit "イメージのプッシュに失敗しました"
fi

echo -e "${GREEN}✓ イメージのプッシュに成功しました:${NC} ${ECR_REPO_URI}:latest"
echo
echo -e "${BLUE}terraform.tfvarsの更新のヒント:${NC}"
echo "このイメージURIをterraform.tfvarsで使用するには："
echo "  ${SERVICE_TYPE}_image = \"${ECR_REPO_URI}:latest\""
echo
echo "または、更新スクリプトを実行してください:"
echo "  scripts/terraform/update-tfvars.sh ${ENVIRONMENT} ${SERVICE_TYPE}"