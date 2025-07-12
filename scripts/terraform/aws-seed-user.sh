#!/bin/bash
# ===================================================================
# ファイル名: aws-seed-user.sh (完全修正版)
# 配置場所: scripts/terraform/aws-seed-user.sh
# 説明: AWS環境でのテストユーザーデータ投入スクリプト（統一設計版）
# 
# 🔧 修正ポイント:
#  - アーキテクチャ指定（--platform linux/amd64）追加
#  - 専用Dockerfileの使用（一時ファイル削除）
#  - ログストリーム名取得方法の統一
#  - エラーハンドリング強化（主要処理と補助処理の分離）
#  - aws-migrate-ci.shとの設計統一
# ===================================================================

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# デバッグ出力関数
debug_echo() {
  echo -e "${CYAN}[DEBUG] $1${NC}"
}

# 引数解析
ENVIRONMENT=${1:-development}

# 環境設定
CLUSTER_NAME="${ENVIRONMENT}-shared-cluster"
TASK_FAMILY="${ENVIRONMENT}-seed-users"
AWS_REGION=${AWS_REGION:-ap-northeast-1}
ECR_REPOSITORY="${ENVIRONMENT}-test-management-seed"

echo -e "${BLUE}========== AWS環境テストユーザー投入 (環境: ${ENVIRONMENT}) ==========${NC}"
echo -e "${BLUE}🔧 統一設計版 - アーキテクチャ指定対応${NC}"

debug_echo "初期変数設定:"
debug_echo "  ENVIRONMENT: ${ENVIRONMENT}"
debug_echo "  CLUSTER_NAME: ${CLUSTER_NAME}"
debug_echo "  TASK_FAMILY: ${TASK_FAMILY}"
debug_echo "  AWS_REGION: ${AWS_REGION}"
debug_echo "  ECR_REPOSITORY: ${ECR_REPOSITORY}"

# AWS認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  echo -e "${RED}エラー: AWS認証情報が設定されていないか、無効です${NC}"
  echo "AWS CLIの設定を確認してください: aws configure"
  exit 1
fi

# AWS認証情報表示
CALLER_IDENTITY=$(aws sts get-caller-identity)
echo -e "${GREEN}✓ AWS認証確認済み${NC}"
echo "  - Account: $(echo $CALLER_IDENTITY | jq -r '.Account')"
echo "  - User: $(echo $CALLER_IDENTITY | jq -r '.Arn' | cut -d'/' -f2)"

debug_echo "AWS認証詳細: $(echo $CALLER_IDENTITY | jq -c .)"

echo -e "${BLUE}ステップ1: AWS APIからRDS接続情報を取得しています...${NC}"

# AWS APIから直接RDS情報を取得（CI環境用）
echo -e "${BLUE}RDSインスタンス情報を検索しています...${NC}"

# RDSインスタンス検索（development環境用）
DB_INSTANCE_ID=$(aws rds describe-db-instances \
  --query 'DBInstances[?contains(DBInstanceIdentifier, `development`) && DBInstanceStatus == `available`].DBInstanceIdentifier' \
  --output text --region ${AWS_REGION} 2>/dev/null | head -1)

debug_echo "RDS検索結果: DB_INSTANCE_ID='${DB_INSTANCE_ID}'"

if [ -z "${DB_INSTANCE_ID}" ] || [ "${DB_INSTANCE_ID}" = "None" ]; then
  echo -e "${RED}エラー: development環境のRDSインスタンスが見つかりません${NC}"
  echo -e "${BLUE}利用可能なRDSインスタンス:${NC}"
  aws rds describe-db-instances \
    --query 'DBInstances[*].[DBInstanceIdentifier,DBInstanceStatus,Engine]' \
    --output table --region ${AWS_REGION} 2>/dev/null || echo "RDS情報取得失敗"
  exit 1
fi

# RDS詳細情報取得
echo -e "${BLUE}RDSインスタンス詳細を取得しています: ${DB_INSTANCE_ID}${NC}"
RDS_INFO=$(aws rds describe-db-instances \
  --db-instance-identifier ${DB_INSTANCE_ID} \
  --region ${AWS_REGION} 2>/dev/null)

DB_HOST=$(echo $RDS_INFO | jq -r '.DBInstances[0].Endpoint.Address')
DB_PORT=$(echo $RDS_INFO | jq -r '.DBInstances[0].Endpoint.Port')
VPC_ID=$(echo $RDS_INFO | jq -r '.DBInstances[0].DBSubnetGroup.VpcId')

# データベース名とユーザー名（固定値）
DB_NAME="test_management_dev"
DB_USERNAME="${TF_VAR_db_username:-testadmin}"

debug_echo "RDS接続情報:"
debug_echo "  DB_HOST: ${DB_HOST}"
debug_echo "  DB_PORT: ${DB_PORT}"  
debug_echo "  DB_NAME: ${DB_NAME}"
debug_echo "  DB_USERNAME: ${DB_USERNAME}"
debug_echo "  VPC_ID: ${VPC_ID}"

if [ -z "${DB_HOST}" ] || [ "${DB_HOST}" = "null" ]; then
  echo -e "${RED}エラー: RDSエンドポイント情報の取得に失敗しました${NC}"
  exit 1
fi

echo -e "${GREEN}✓ RDS接続情報を取得しました${NC}"
echo "  - インスタンスID: ${DB_INSTANCE_ID}"
echo "  - ホスト: ${DB_HOST}"
echo "  - ポート: ${DB_PORT}"
echo "  - データベース名: ${DB_NAME}"
echo "  - ユーザー名: ${DB_USERNAME}"
echo "  - VPC ID: ${VPC_ID}"

echo -e "${BLUE}ステップ2: テストデータ投入用Dockerイメージをビルドしています...${NC}"

# ECRリポジトリの作成（存在しない場合）
if ! aws ecr describe-repositories --repository-names ${ECR_REPOSITORY} --region ${AWS_REGION} &>/dev/null; then
  echo -e "${YELLOW}ECRリポジトリが存在しません。作成しています...${NC}"
  aws ecr create-repository --repository-name ${ECR_REPOSITORY} --region ${AWS_REGION}
  echo -e "${GREEN}✓ ECRリポジトリを作成しました${NC}"
fi

# ECRログイン
echo -e "${BLUE}ECRにログインしています...${NC}"
aws ecr get-login-password --region ${AWS_REGION} | \
  docker login --username AWS --password-stdin $(aws sts get-caller-identity --query Account --output text).dkr.ecr.${AWS_REGION}.amazonaws.com

# テストデータ投入用Dockerイメージのビルド（統一設計）
echo -e "${BLUE}テストデータ投入用イメージをビルド・プッシュしています...${NC}"
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
IMAGE_URI="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}:latest"

debug_echo "Docker情報:"
debug_echo "  ACCOUNT_ID: ${ACCOUNT_ID}"
debug_echo "  IMAGE_URI: ${IMAGE_URI}"

# Docker buildx build（アーキテクチャ指定、専用Dockerfile使用）
docker buildx build --platform linux/amd64 \
  -f deployments/docker/seed.Dockerfile \
  -t ${IMAGE_URI} . --push

echo -e "${GREEN}✓ テストデータ投入用イメージの準備が完了しました${NC}"

echo -e "${BLUE}ステップ3: ECSタスク定義を作成しています...${NC}"

# SSMからパスワード直接取得
echo -e "${BLUE}パスワードを取得しています...${NC}"
DB_PASSWORD_VALUE=$(aws ssm get-parameter \
  --name "/${ENVIRONMENT}/database/password" \
  --with-decryption \
  --query 'Parameter.Value' \
  --output text \
  --region ${AWS_REGION})

debug_echo "パスワード取得結果: '$(echo $DB_PASSWORD_VALUE | sed 's/./*/g')' (マスク表示)"

if [ -z "${DB_PASSWORD_VALUE}" ] || [ "${DB_PASSWORD_VALUE}" = "None" ]; then
  echo -e "${RED}エラー: データベースパスワードの取得に失敗しました${NC}"
  exit 1
fi

# URL エスケープ処理関数
escape_url_component() {
  local component="$1"
  # 感嘆符のエスケープ（最優先）
  component=$(echo "$component" | sed 's/!/%21/g')
  # その他の特殊文字もエスケープ
  component=$(echo "$component" | sed 's/@/%40/g')
  component=$(echo "$component" | sed 's/#/%23/g')
  component=$(echo "$component" | sed 's/%/%25/g')
  component=$(echo "$component" | sed 's/ /%20/g')
  echo "$component"
}

# パスワードエスケープ処理
DB_PASSWORD_ESCAPED=$(escape_url_component "$DB_PASSWORD_VALUE")

# エスケープ済みパスワードでDATABASE_URL構築
DATABASE_URL="postgresql://${DB_USERNAME}:${DB_PASSWORD_ESCAPED}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require"

debug_echo "DATABASE_URL構築結果: postgresql://${DB_USERNAME}:****@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require"

# JSON安全化関数
json_escape() {
  local value="$1"
  # JSON特殊文字をエスケープ
  value=$(echo "$value" | sed 's/\\/\\\\/g')  # バックスラッシュ
  value=$(echo "$value" | sed 's/"/\\"/g')    # ダブルクォート
  echo "$value"
}

# DATABASE_URLのJSON安全化
DATABASE_URL_ESCAPED_JSON=$(json_escape "$DATABASE_URL")

# ECSタスク定義JSON作成
cat > /tmp/seed-task-definition.json << EOF
{
  "family": "${TASK_FAMILY}",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::${ACCOUNT_ID}:role/${ENVIRONMENT}-shared-task-execution-role",
  "taskRoleArn": "arn:aws:iam::${ACCOUNT_ID}:role/${ENVIRONMENT}-shared-task-execution-role",
  "containerDefinitions": [
    {
      "name": "seed-container",
      "image": "${IMAGE_URI}",
      "cpu": 256,
      "memory": 512,
      "essential": true,
      "command": [
        "${DATABASE_URL_ESCAPED_JSON}",
        "-f", "/sql/aws-test-users.sql"
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/seed-users",
          "awslogs-region": "${AWS_REGION}",
          "awslogs-stream-prefix": "seed-${ENVIRONMENT}"
        }
      }
    }
  ],
  "runtimePlatform": {
    "operatingSystemFamily": "LINUX",
    "cpuArchitecture": "X86_64"
  }
}
EOF

# JSON構文チェック
if jq . /tmp/seed-task-definition.json > /dev/null 2>&1; then
  echo -e "${GREEN}✓ JSON構文は正常です${NC}"
  debug_echo "jq による JSON解析成功"
else
  echo -e "${RED}❌ JSON構文エラーが発生しています${NC}"
  echo -e "${YELLOW}jq詳細エラー:${NC}"
  jq . /tmp/seed-task-definition.json 2>&1 || true
  echo -e "${RED}JSON構文エラーのため処理を中断します${NC}"
  exit 1
fi

# CloudWatchログ グループの作成（存在しない場合）
if ! aws logs describe-log-groups --log-group-name-prefix "/ecs/seed-users" --region ${AWS_REGION} | grep -q "/ecs/seed-users"; then
  echo -e "${YELLOW}CloudWatchログ グループを作成しています...${NC}"
  aws logs create-log-group --log-group-name "/ecs/seed-users" --region ${AWS_REGION}
  echo -e "${GREEN}✓ CloudWatchログ グループを作成しました: /ecs/seed-users${NC}"
  
  # ログリテンション設定（オプション：30日）
  aws logs put-retention-policy --log-group-name "/ecs/seed-users" --retention-in-days 30 --region ${AWS_REGION}
  echo -e "${GREEN}✓ ログリテンション期間を30日に設定しました${NC}"
fi

# タスク定義の登録
echo -e "${BLUE}ECSタスク定義を登録しています...${NC}"
TASK_DEFINITION_ARN=$(aws ecs register-task-definition \
  --cli-input-json file:///tmp/seed-task-definition.json \
  --region ${AWS_REGION} \
  --query 'taskDefinition.taskDefinitionArn' --output text 2>&1)

# AWS CLI実行結果の詳細確認
if [[ $? -eq 0 && "$TASK_DEFINITION_ARN" != *"Error"* ]]; then
  echo -e "${GREEN}✓ ECSタスク定義を登録しました: ${TASK_DEFINITION_ARN}${NC}"
  debug_echo "タスク定義登録成功: ${TASK_DEFINITION_ARN}"
else
  echo -e "${RED}❌ ECSタスク定義の登録に失敗しました${NC}"
  echo -e "${YELLOW}エラー詳細:${NC}"
  echo "$TASK_DEFINITION_ARN"
  exit 1
fi

echo -e "${BLUE}ステップ4: テストデータ投入タスクを実行しています...${NC}"

# VPC設定情報の取得（AWS API使用）
echo -e "${BLUE}VPCネットワーク情報を取得しています...${NC}"

# プライベートサブネット取得
PRIVATE_SUBNET_IDS=$(aws ec2 describe-subnets \
  --filters "Name=vpc-id,Values=${VPC_ID}" "Name=tag:Name,Values=*private*" \
  --query 'Subnets[].SubnetId' --output text --region ${AWS_REGION} | tr '\t' ',' | sed 's/,$//')

if [ -z "${PRIVATE_SUBNET_IDS}" ]; then
  echo -e "${YELLOW}プライベートサブネットが見つかりません。すべてのサブネットを検索しています...${NC}"
  PRIVATE_SUBNET_IDS=$(aws ec2 describe-subnets \
    --filters "Name=vpc-id,Values=${VPC_ID}" \
    --query 'Subnets[?MapPublicIpOnLaunch==`false`].SubnetId' --output text --region ${AWS_REGION} | tr '\t' ',' | sed 's/,$//')
fi

# セキュリティグループの取得（GraphQL用セキュリティグループを使用）
echo -e "${BLUE}GraphQL用セキュリティグループを検索しています...${NC}"
SECURITY_GROUP_ID=$(aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=${ENVIRONMENT}-graphql-new-tasks-sg" "Name=vpc-id,Values=${VPC_ID}" \
  --query 'SecurityGroups[0].GroupId' --output text --region ${AWS_REGION} 2>/dev/null || echo "")

# フォールバック: 汎用的な検索
if [ "${SECURITY_GROUP_ID}" = "None" ] || [ -z "${SECURITY_GROUP_ID}" ]; then
  echo -e "${YELLOW}GraphQL固有のセキュリティグループが見つかりません。汎用検索を実行しています...${NC}"
  SECURITY_GROUP_ID=$(aws ec2 describe-security-groups \
    --filters "Name=vpc-id,Values=${VPC_ID}" \
    --query 'SecurityGroups[?contains(GroupName, `graphql`) && contains(GroupName, `tasks`)].GroupId | [0]' \
    --output text --region ${AWS_REGION} 2>/dev/null || echo "")
fi

# 最終フォールバック: development環境のタスク用セキュリティグループ
if [ "${SECURITY_GROUP_ID}" = "None" ] || [ -z "${SECURITY_GROUP_ID}" ]; then
  echo -e "${YELLOW}development環境のタスク用セキュリティグループを検索しています...${NC}"
  SECURITY_GROUP_ID=$(aws ec2 describe-security-groups \
    --filters "Name=vpc-id,Values=${VPC_ID}" "Name=tag:Environment,Values=${ENVIRONMENT}" \
    --query 'SecurityGroups[?contains(GroupName, `tasks`)].GroupId | [0]' \
    --output text --region ${AWS_REGION} 2>/dev/null || echo "")
fi

debug_echo "VPCネットワーク情報:"
debug_echo "  VPC_ID: ${VPC_ID}"
debug_echo "  PRIVATE_SUBNET_IDS: ${PRIVATE_SUBNET_IDS}"
debug_echo "  SECURITY_GROUP_ID: ${SECURITY_GROUP_ID}"

if [ "${SECURITY_GROUP_ID}" = "None" ] || [ -z "${SECURITY_GROUP_ID}" ]; then
  echo -e "${RED}エラー: 適切なセキュリティグループが見つかりません${NC}"
  echo -e "${BLUE}利用可能なセキュリティグループ:${NC}"
  aws ec2 describe-security-groups --filters "Name=vpc-id,Values=${VPC_ID}" \
    --query 'SecurityGroups[*].[GroupName,GroupId,Description]' --output table --region ${AWS_REGION}
  exit 1
fi

echo -e "${BLUE}ネットワーク設定:${NC}"
echo "  - VPC ID: ${VPC_ID}"
echo "  - プライベートサブネット: ${PRIVATE_SUBNET_IDS}"
echo "  - セキュリティグループ: ${SECURITY_GROUP_ID}"

# ECSタスク実行
echo -e "${BLUE}テストデータ投入タスクを起動しています...${NC}"

debug_echo "ECSタスク実行準備:"
debug_echo "  CLUSTER_NAME: ${CLUSTER_NAME}"  
debug_echo "  TASK_DEFINITION_ARN: ${TASK_DEFINITION_ARN}"
debug_echo "  PRIVATE_SUBNET_IDS: ${PRIVATE_SUBNET_IDS}"
debug_echo "  SECURITY_GROUP_ID: ${SECURITY_GROUP_ID}"

TASK_ARN=$(aws ecs run-task \
  --cluster ${CLUSTER_NAME} \
  --task-definition ${TASK_DEFINITION_ARN} \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[${PRIVATE_SUBNET_IDS}],securityGroups=[${SECURITY_GROUP_ID}],assignPublicIp=DISABLED}" \
  --region ${AWS_REGION} \
  --query 'tasks[0].taskArn' --output text)

debug_echo "ECSタスク起動結果: TASK_ARN='${TASK_ARN}'"

if [ "${TASK_ARN}" = "None" ] || [ -z "${TASK_ARN}" ]; then
  echo -e "${RED}エラー: タスクの起動に失敗しました${NC}"
  echo -e "${BLUE}ECSクラスター状態:${NC}"
  aws ecs describe-clusters --clusters ${CLUSTER_NAME} --region ${AWS_REGION} \
    --query 'clusters[0].{Name:clusterName,Status:status,ActiveServices:activeServicesCount,RunningTasks:runningTasksCount}' \
    --output table 2>/dev/null || echo "ECSクラスター情報取得失敗"
  exit 1
fi

echo -e "${GREEN}✓ テストデータ投入タスクを起動しました: ${TASK_ARN}${NC}"

echo -e "${BLUE}ステップ5: タスクの完了を待機しています...${NC}"

# タスク完了待機（最大5分）
WAIT_COUNT=0
MAX_WAIT=30  # 5分（10秒 × 30回）

while [ ${WAIT_COUNT} -lt ${MAX_WAIT} ]; do
  TASK_STATUS=$(aws ecs describe-tasks \
    --cluster ${CLUSTER_NAME} \
    --tasks ${TASK_ARN} \
    --region ${AWS_REGION} \
    --query 'tasks[0].lastStatus' --output text)
  
  debug_echo "タスク状態確認 (${WAIT_COUNT}/${MAX_WAIT}): ${TASK_STATUS}"
  echo -n "."
  
  if [ "${TASK_STATUS}" = "STOPPED" ]; then
    echo -e "\n${GREEN}✓ タスクが完了しました${NC}"
    break
  fi
  
  sleep 10
  WAIT_COUNT=$((WAIT_COUNT + 1))
done

if [ ${WAIT_COUNT} -ge ${MAX_WAIT} ]; then
  echo -e "\n${RED}エラー: タスクがタイムアウトしました${NC}"
  echo -e "${BLUE}現在のタスク状態: ${TASK_STATUS}${NC}"
  exit 1
fi

echo -e "${BLUE}ステップ6: 実行結果を確認しています...${NC}"

# タスクの詳細情報取得
TASK_DETAILS=$(aws ecs describe-tasks \
  --cluster ${CLUSTER_NAME} \
  --tasks ${TASK_ARN} \
  --region ${AWS_REGION})

EXIT_CODE=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].containers[0].exitCode')
STOP_REASON=$(echo "$TASK_DETAILS" | jq -r '.tasks[0].stoppedReason')

debug_echo "タスク詳細情報:"
debug_echo "  EXIT_CODE: ${EXIT_CODE}"
debug_echo "  STOP_REASON: ${STOP_REASON}"

echo "タスクの終了コード: ${EXIT_CODE}"
echo "停止理由: ${STOP_REASON}"

# CloudWatchログの表示（統一方式）
echo -e "${BLUE}テストデータ投入ログ:${NC}"
TASK_ID=$(echo ${TASK_ARN} | cut -d'/' -f3)
LOG_STREAM_NAME="seed-${ENVIRONMENT}/seed-container/${TASK_ID}"

debug_echo "CloudWatchログ確認:"
debug_echo "  TASK_ID: ${TASK_ID}"
debug_echo "  LOG_STREAM_NAME: ${LOG_STREAM_NAME}"

# 結果判定 - 先に成功・失敗判定を行い、ログ取得はオプション扱い
if [ "${EXIT_CODE}" = "0" ]; then
  echo -e "${GREEN}========== テストユーザー投入成功 ==========${NC}"
  echo -e "${GREEN}✓ test_adminユーザーが正常に作成されました${NC}"
  
  # ログ取得を試行するが、失敗してもエラーにしない
  if [ "${LOG_STREAM_NAME}" != "None" ] && [ ! -z "${LOG_STREAM_NAME}" ]; then
    echo -e "${MAGENTA}=== CloudWatchログ内容 ===${NC}"
    aws logs get-log-events \
      --log-group-name "/ecs/seed-users" \
      --log-stream-name "${LOG_STREAM_NAME}" \
      --region ${AWS_REGION} \
      --query 'events[].message' --output text || echo -e "${YELLOW}ログ取得に失敗しましたが、テストユーザー投入は成功しています${NC}"
    echo -e "${MAGENTA}=== ログ終了 ===${NC}"
  else
    echo -e "${YELLOW}警告: ログストリームが見つかりませんが、テストユーザー投入は成功しています${NC}"
  fi
  
  echo -e "${BLUE}テストユーザー情報:${NC}"
  echo -e "  - ユーザー名: test_admin"
  echo -e "  - パスワード: password"
  echo -e "  - ロール: Admin"
  echo -e "  - ユーザーID: USER001"
  echo -e "${BLUE}次のステップ: GraphQL認証テストを実行してください${NC}"
  echo -e "  make verify-graphql-health TF_ENV=${ENVIRONMENT}"
  
  # 常に成功で終了
  exit 0
else
  echo -e "${RED}========== テストユーザー投入失敗 ==========${NC}"
  echo -e "${RED}✗ テストデータ投入中にエラーが発生しました (終了コード: ${EXIT_CODE})${NC}"
  echo -e "${YELLOW}トラブルシューティング:${NC}"
  echo -e "1. CloudWatchログでエラー詳細を確認してください"
  echo -e "2. マイグレーションが正常に完了しているか確認してください"
  echo -e "3. usersテーブルが存在するか確認してください"
  echo -e "${MAGENTA}=== デバッグ情報サマリー ===${NC}"
  debug_echo "DATABASE_URL: $(echo $DATABASE_URL | sed 's/:[^:]*@/:****@/')"
  debug_echo "TASK_ARN: ${TASK_ARN}"
  debug_echo "TASK_DEFINITION_ARN: ${TASK_DEFINITION_ARN}"
  exit 1
fi

# クリーンアップ（一時ファイル削除）
rm -f /tmp/seed-task-definition.json

echo -e "${GREEN}AWS環境テストユーザー投入が完了しました${NC}"