#!/bin/bash
# ===================================================================
# ファイル名: aws-migrate.sh
# 配置場所: scripts/terraform/aws-migrate.sh
# 説明: AWS環境でのデータベースマイグレーション実行スクリプト（DATABASE_URL事前構築版）
# 
# 修正内容:
#  - DATABASE_URL事前構築方式への根本的変更
#  - URL encoding対応による特殊文字問題解決
#  - Here Document方式の完全廃止
# 
# 用途:
#  - ECSタスクを使用してVPC内でマイグレーションを実行
#  - RDS PostgreSQLに対してマイグレーションファイルを適用
#  - migrate CLIツールをコンテナ内で実行
# 
# 実行フロー:
#  1. terraform outputからRDS接続情報を取得
#  2. SSMからパスワード直接取得・URL encoding実行
#  3. 完全なDATABASE_URL構築
#  4. マイグレーション専用Dockerイメージをビルド・ECRにプッシュ
#  5. ECSでワンタイムタスクを起動してマイグレーション実行
#  6. 実行結果を確認・検証
# 
# 使用方法:
#  ./aws-migrate.sh <環境名>
#
# 引数:
#  環境名 - マイグレーション対象環境（development, production）
# ===================================================================

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 引数解析
ENVIRONMENT=${1:-development}

# 環境設定
TERRAFORM_DIR="deployments/terraform/environments/${ENVIRONMENT}"
CLUSTER_NAME="${ENVIRONMENT}-shared-cluster"
TASK_FAMILY="${ENVIRONMENT}-migration"
AWS_REGION=${AWS_REGION:-ap-northeast-1}
ECR_REPOSITORY="${ENVIRONMENT}-test-management-migration"

echo -e "${BLUE}========== AWS環境マイグレーション実行 (環境: ${ENVIRONMENT}) ==========${NC}"

# AWS認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  echo -e "${RED}エラー: AWS認証情報が設定されていないか、無効です${NC}"
  echo "AWS CLIの設定を確認してください: aws configure"
  exit 1
fi

# Terraformディレクトリの確認
if [ ! -d "${TERRAFORM_DIR}" ]; then
  echo -e "${RED}エラー: Terraform環境ディレクトリが見つかりません: ${TERRAFORM_DIR}${NC}"
  exit 1
fi

echo -e "${BLUE}ステップ1: RDS接続情報を取得しています...${NC}"

# Terraformからデータベース接続情報を取得
cd ${TERRAFORM_DIR}

DB_HOST=$(terraform output -raw db_instance_address 2>/dev/null || echo "")
DB_NAME=$(terraform output -raw db_name 2>/dev/null || echo "")
DB_USERNAME="admin"  # 固定値として使用

if [ -z "${DB_HOST}" ] || [ "${DB_HOST}" = "null" ]; then
  echo -e "${RED}エラー: RDSインスタンスが見つかりません${NC}"
  echo "まずGraphQLサービスをデプロイしてください: make deploy-graphql-new-dev"
  exit 1
fi

echo -e "${GREEN}✓ RDS接続情報を取得しました${NC}"
echo "  - ホスト: ${DB_HOST}"
echo "  - データベース名: ${DB_NAME}"
echo "  - ユーザー名: ${DB_USERNAME}"

# プロジェクトルートディレクトリに戻る
cd - > /dev/null

echo -e "${BLUE}ステップ1.5: データベース接続URL構築しています...${NC}"

# SSMからパスワード直接取得
echo -e "${BLUE}パスワードを取得しています...${NC}"
DB_PASSWORD_VALUE=$(aws ssm get-parameter \
  --name "/${ENVIRONMENT}/database/password" \
  --with-decryption \
  --query 'Parameter.Value' \
  --output text \
  --region ${AWS_REGION})

if [ -z "${DB_PASSWORD_VALUE}" ] || [ "${DB_PASSWORD_VALUE}" = "None" ]; then
  echo -e "${RED}エラー: データベースパスワードの取得に失敗しました${NC}"
  exit 1
fi

# URL encoding処理
echo -e "${BLUE}パスワードをURL encodingしています...${NC}"
DB_PASSWORD_ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('${DB_PASSWORD_VALUE}', safe=''))")
echo -e "${GREEN}✓ パスワードをURL encodingしました${NC}"

# 完全なDATABASE_URL構築
DATABASE_URL="postgresql://${DB_USERNAME}:${DB_PASSWORD_ENCODED}@${DB_HOST}:5432/${DB_NAME}?sslmode=require"
echo -e "${GREEN}✓ データベース接続URLを構築しました${NC}"
echo "  - DATABASE_URL: postgresql://${DB_USERNAME}:****@${DB_HOST}:5432/${DB_NAME}?sslmode=require"

echo -e "${BLUE}ステップ2: マイグレーション用Dockerイメージをビルドしています...${NC}"

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

# Dockerイメージのビルド
echo -e "${BLUE}マイグレーション用Dockerイメージをビルドしています...${NC}"
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
IMAGE_URI="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}:latest"

# docker buildx build --platform linux/amd64 \
#   -f deployments/docker/migrate.Dockerfile \
#   -t ${ECR_REPOSITORY}:latest . --load

docker buildx build --platform linux/amd64 \
  -f deployments/docker/migrate.Dockerfile \
  -t ${IMAGE_URI} . --push

# docker tagコマンド（ECRプッシュ用）
docker tag ${ECR_REPOSITORY}:latest ${IMAGE_URI}

# ECRにプッシュ
echo -e "${BLUE}ECRにイメージをプッシュしています...${NC}"
docker push ${IMAGE_URI}
echo -e "${GREEN}✓ マイグレーション用イメージの準備が完了しました${NC}"

echo -e "${BLUE}ステップ3: ECSタスク定義を作成しています...${NC}"

# ECSタスク定義JSON作成（DATABASE_URL方式）
cat > /tmp/migration-task-definition.json << EOF
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
      "name": "migration-container",
      "image": "${IMAGE_URI}",
      "essential": true,
      "environment": [
        {
          "name": "DATABASE_URL",
          "value": "${DATABASE_URL}"
        }
      ],
      "command": [
        "migrate", 
        "-path", "/migrations", 
        "-database", "postgresql://admin:SecurePassword123%21@development-postgres.cngycmkc0hhn.ap-northeast-1.rds.amazonaws.com:5432/test_management_dev?sslmode=require",
        "up"
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/migration",
          "awslogs-region": "${AWS_REGION}",
          "awslogs-stream-prefix": "migration-${ENVIRONMENT}"
        }
      }
    }
  ]
}
EOF

# CloudWatchログ グループの作成（存在しない場合）
if ! aws logs describe-log-groups --log-group-name-prefix "/ecs/migration" --region ${AWS_REGION} | grep -q "/ecs/migration"; then
  echo -e "${YELLOW}CloudWatchログ グループを作成しています...${NC}"
  aws logs create-log-group --log-group-name "/ecs/migration" --region ${AWS_REGION}
fi

# タスク定義の登録
echo -e "${BLUE}ECSタスク定義を登録しています...${NC}"
TASK_DEFINITION_ARN=$(aws ecs register-task-definition \
  --cli-input-json file:///tmp/migration-task-definition.json \
  --region ${AWS_REGION} \
  --query 'taskDefinition.taskDefinitionArn' --output text)

echo -e "${GREEN}✓ ECSタスク定義を登録しました: ${TASK_DEFINITION_ARN}${NC}"

echo -e "${BLUE}ステップ4: マイグレーションタスクを実行しています...${NC}"

# VPC設定情報の取得
cd ${TERRAFORM_DIR}
PRIVATE_SUBNET_IDS=$(terraform output -json private_subnet_ids | jq -r '.[]' | tr '\n' ',' | sed 's/,$//')
VPC_ID=$(terraform output -raw vpc_id)
cd - > /dev/null

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

# 最終フォールバック: 利用可能なセキュリティグループの表示
if [ "${SECURITY_GROUP_ID}" = "None" ] || [ -z "${SECURITY_GROUP_ID}" ]; then
  echo -e "${RED}エラー: GraphQL用セキュリティグループが見つかりません${NC}"
  echo -e "${BLUE}利用可能なセキュリティグループ:${NC}"
  aws ec2 describe-security-groups --filters "Name=vpc-id,Values=${VPC_ID}" \
    --query 'SecurityGroups[*].[GroupName,GroupId]' --output table --region ${AWS_REGION}
  echo -e "${YELLOW}対処方法:${NC}"
  echo "1. GraphQLサービスがデプロイされているか確認してください"
  echo "2. 上記の一覧からタスク用セキュリティグループを確認してください"
  echo "3. 必要に応じて手動でセキュリティグループIDを指定してください"
  exit 1
fi

echo -e "${BLUE}ネットワーク設定:${NC}"
echo "  - VPC ID: ${VPC_ID}"
echo "  - プライベートサブネット: ${PRIVATE_SUBNET_IDS}"
echo "  - セキュリティグループ: ${SECURITY_GROUP_ID}"

# ECSタスク実行
echo -e "${BLUE}マイグレーションタスクを起動しています...${NC}"
TASK_ARN=$(aws ecs run-task \
  --cluster ${CLUSTER_NAME} \
  --task-definition ${TASK_DEFINITION_ARN} \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[${PRIVATE_SUBNET_IDS}],securityGroups=[${SECURITY_GROUP_ID}],assignPublicIp=DISABLED}" \
  --region ${AWS_REGION} \
  --query 'tasks[0].taskArn' --output text)

if [ "${TASK_ARN}" = "None" ] || [ -z "${TASK_ARN}" ]; then
  echo -e "${RED}エラー: タスクの起動に失敗しました${NC}"
  exit 1
fi

echo -e "${GREEN}✓ マイグレーションタスクを起動しました: ${TASK_ARN}${NC}"

echo -e "${BLUE}ステップ5: タスクの完了を待機しています...${NC}"

# タスク完了待機（最大10分）
WAIT_COUNT=0
MAX_WAIT=60  # 10分（10秒 × 60回）

while [ ${WAIT_COUNT} -lt ${MAX_WAIT} ]; do
  TASK_STATUS=$(aws ecs describe-tasks \
    --cluster ${CLUSTER_NAME} \
    --tasks ${TASK_ARN} \
    --region ${AWS_REGION} \
    --query 'tasks[0].lastStatus' --output text)
  
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
  exit 1
fi

echo -e "${BLUE}ステップ6: 実行結果を確認しています...${NC}"

# タスクの終了コードを確認
EXIT_CODE=$(aws ecs describe-tasks \
  --cluster ${CLUSTER_NAME} \
  --tasks ${TASK_ARN} \
  --region ${AWS_REGION} \
  --query 'tasks[0].containers[0].exitCode' --output text)

STOP_REASON=$(aws ecs describe-tasks \
  --cluster ${CLUSTER_NAME} \
  --tasks ${TASK_ARN} \
  --region ${AWS_REGION} \
  --query 'tasks[0].stoppedReason' --output text)

echo "タスクの終了コード: ${EXIT_CODE}"
echo "停止理由: ${STOP_REASON}"

# CloudWatchログの表示
echo -e "${BLUE}マイグレーション実行ログ:${NC}"
LOG_STREAM_NAME=$(aws logs describe-log-streams \
  --log-group-name "/ecs/migration" \
  --order-by LastEventTime \
  --descending \
  --max-items 1 \
  --region ${AWS_REGION} \
  --query 'logStreams[0].logStreamName' --output text)

if [ "${LOG_STREAM_NAME}" != "None" ] && [ ! -z "${LOG_STREAM_NAME}" ]; then
  aws logs get-log-events \
    --log-group-name "/ecs/migration" \
    --log-stream-name "${LOG_STREAM_NAME}" \
    --region ${AWS_REGION} \
    --query 'events[].message' --output text
else
  echo -e "${YELLOW}警告: ログストリームが見つかりません${NC}"
fi

# 結果判定
if [ "${EXIT_CODE}" = "0" ]; then
  echo -e "${GREEN}========== マイグレーション成功 ==========${NC}"
  echo -e "${GREEN}✓ 全てのマイグレーションファイルが正常に適用されました${NC}"
  echo -e "${BLUE}次のステップ: テストユーザーを投入してください${NC}"
  echo -e "  make seed-test-users-dev TF_ENV=${ENVIRONMENT}"
else
  echo -e "${RED}========== マイグレーション失敗 ==========${NC}"
  echo -e "${RED}✗ マイグレーション実行中にエラーが発生しました (終了コード: ${EXIT_CODE})${NC}"
  echo -e "${YELLOW}トラブルシューティング:${NC}"
  echo -e "1. CloudWatchログでエラー詳細を確認してください"
  echo -e "2. RDS接続情報を確認してください"
  echo -e "3. セキュリティグループ設定を確認してください"
  exit 1
fi

# クリーンアップ（一時ファイル削除）
rm -f /tmp/migration-task-definition.json

echo -e "${GREEN}AWS環境マイグレーション実行が完了しました${NC}"