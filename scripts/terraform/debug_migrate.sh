#!/bin/bash
# デバッグ用migrate実行スクリプト（JSON構文修正版）
# アーキテクチャ問題とmigrate CLI問題を段階的に特定

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ENVIRONMENT=${1:-development}
TERRAFORM_DIR="deployments/terraform/environments/${ENVIRONMENT}"
CLUSTER_NAME="${ENVIRONMENT}-shared-cluster"
TASK_FAMILY="${ENVIRONMENT}-migration-debug"
AWS_REGION=${AWS_REGION:-ap-northeast-1}
ECR_REPOSITORY="${ENVIRONMENT}-test-management-migration"

echo -e "${BLUE}========== アーキテクチャ問題デバッグ：migrate CLI特定 ==========${NC}"

# RDS接続情報取得
cd ${TERRAFORM_DIR}
DB_HOST=$(terraform output -raw db_instance_address)
DB_NAME=$(terraform output -raw db_name)
DB_USERNAME="admin"
cd - > /dev/null

# パスワード取得・URL encoding
DB_PASSWORD_VALUE=$(aws ssm get-parameter \
  --name "/${ENVIRONMENT}/database/password" \
  --with-decryption \
  --query 'Parameter.Value' \
  --output text \
  --region ${AWS_REGION})

DB_PASSWORD_ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('${DB_PASSWORD_VALUE}', safe=''))")

# 完全なDATABASE_URL構築
DATABASE_URL="postgresql://${DB_USERNAME}:${DB_PASSWORD_ENCODED}@${DB_HOST}:5432/${DB_NAME}?sslmode=require"

echo -e "${GREEN}DATABASE_URL: postgresql://${DB_USERNAME}:****@${DB_HOST}:5432/${DB_NAME}?sslmode=require${NC}"

# アカウントID取得
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
IMAGE_URI="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}:latest"

# VPC設定取得
cd ${TERRAFORM_DIR}
PRIVATE_SUBNET_IDS=$(terraform output -json private_subnet_ids | jq -r '.[]' | tr '\n' ',' | sed 's/,$//')
VPC_ID=$(terraform output -raw vpc_id)
cd - > /dev/null

# セキュリティグループ取得
SECURITY_GROUP_ID=$(aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=${ENVIRONMENT}-graphql-new-tasks-sg" "Name=vpc-id,Values=${VPC_ID}" \
  --query 'SecurityGroups[0].GroupId' --output text --region ${AWS_REGION})

echo -e "${BLUE}Phase 1: アーキテクチャとmigrate CLI基本動作確認${NC}"

# CloudWatchログ グループ確認・作成
if ! aws logs describe-log-groups --log-group-name-prefix "/ecs/migration" --region ${AWS_REGION} | grep -q "/ecs/migration"; then
  echo -e "${YELLOW}CloudWatchログ グループを作成しています...${NC}"
  aws logs create-log-group --log-group-name "/ecs/migration" --region ${AWS_REGION}
fi

# Phase 1: アーキテクチャ確認用タスク定義（JSON構文修正版）
cat > /tmp/debug-task-definition-phase1.json << 'EOF'
{
  "family": "development-migration-debug-phase1",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::504956919683:role/development-shared-task-execution-role",
  "taskRoleArn": "arn:aws:iam::504956919683:role/development-shared-task-execution-role",
  "containerDefinitions": [
    {
      "name": "debug-container",
      "image": "504956919683.dkr.ecr.ap-northeast-1.amazonaws.com/development-test-management-migration:latest",
      "essential": true,
      "command": [
        "sh", 
        "-c", 
        "echo '=== ARCHITECTURE CHECK ===' && uname -m && echo '=== MIGRATE VERSION ===' && migrate --version && echo '=== MIGRATE HELP (first 10 lines) ===' && migrate --help | head -10 && echo '=== END PHASE 1 ==="
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/migration",
          "awslogs-region": "ap-northeast-1",
          "awslogs-stream-prefix": "debug-development"
        }
      }
    }
  ]
}
EOF

# Phase 1タスク定義登録・実行
echo -e "${BLUE}Phase 1タスク定義を登録しています...${NC}"
TASK_DEFINITION_ARN_PHASE1=$(aws ecs register-task-definition \
  --cli-input-json file:///tmp/debug-task-definition-phase1.json \
  --region ${AWS_REGION} \
  --query 'taskDefinition.taskDefinitionArn' --output text)

echo -e "${GREEN}Phase 1タスク定義登録完了: ${TASK_DEFINITION_ARN_PHASE1}${NC}"

echo -e "${BLUE}Phase 1タスクを実行しています...${NC}"
TASK_ARN_PHASE1=$(aws ecs run-task \
  --cluster ${CLUSTER_NAME} \
  --task-definition ${TASK_DEFINITION_ARN_PHASE1} \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[${PRIVATE_SUBNET_IDS}],securityGroups=[${SECURITY_GROUP_ID}],assignPublicIp=DISABLED}" \
  --region ${AWS_REGION} \
  --query 'tasks[0].taskArn' --output text)

echo -e "${GREEN}Phase 1タスク起動: ${TASK_ARN_PHASE1}${NC}"

# Phase 1完了待機
echo -e "${BLUE}Phase 1完了を待機中...${NC}"
aws ecs wait tasks-stopped --cluster ${CLUSTER_NAME} --tasks ${TASK_ARN_PHASE1} --region ${AWS_REGION}

# Phase 1結果確認
EXIT_CODE_PHASE1=$(aws ecs describe-tasks \
  --cluster ${CLUSTER_NAME} \
  --tasks ${TASK_ARN_PHASE1} \
  --region ${AWS_REGION} \
  --query 'tasks[0].containers[0].exitCode' --output text)

STOP_REASON_PHASE1=$(aws ecs describe-tasks \
  --cluster ${CLUSTER_NAME} \
  --tasks ${TASK_ARN_PHASE1} \
  --region ${AWS_REGION} \
  --query 'tasks[0].stoppedReason' --output text)

echo -e "${GREEN}Phase 1終了コード: ${EXIT_CODE_PHASE1}${NC}"
echo -e "${GREEN}Phase 1停止理由: ${STOP_REASON_PHASE1}${NC}"

# Phase 1ログ表示
echo -e "${BLUE}Phase 1実行ログ:${NC}"
sleep 10  # ログ書き込み待機
LOG_STREAM_NAME_PHASE1=$(aws logs describe-log-streams \
  --log-group-name "/ecs/migration" \
  --order-by LastEventTime \
  --descending \
  --max-items 1 \
  --region ${AWS_REGION} \
  --query 'logStreams[0].logStreamName' --output text)

if [ "${LOG_STREAM_NAME_PHASE1}" != "None" ] && [ ! -z "${LOG_STREAM_NAME_PHASE1}" ]; then
  aws logs get-log-events \
    --log-group-name "/ecs/migration" \
    --log-stream-name "${LOG_STREAM_NAME_PHASE1}" \
    --region ${AWS_REGION} \
    --query 'events[].message' --output text
else
  echo -e "${YELLOW}警告: Phase 1ログストリームが見つかりません${NC}"
fi

echo -e "${BLUE}========== Phase 1 結果分析 ==========${NC}"
if [ "${EXIT_CODE_PHASE1}" = "0" ]; then
  echo -e "${GREEN}✓ Phase 1成功: コンテナとmigrate CLIは正常に動作${NC}"
  echo -e "${BLUE}Phase 2: データベース接続テストを実行しますか？ (y/n)${NC}"
  read -r CONTINUE_PHASE2
  if [[ "$CONTINUE_PHASE2" =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Phase 2を実行中...${NC}"
    # Phase 2のコードをここに追加可能
  else
    echo -e "${YELLOW}Phase 2をスキップしました${NC}"
  fi
else
  echo -e "${RED}✗ Phase 1失敗: アーキテクチャまたはmigrate CLI問題${NC}"
  echo -e "${YELLOW}対処法:${NC}"
  echo -e "1. Docker imageを--platform=linux/amd64で再ビルド"
  echo -e "2. migrate CLIの問題を詳細調査"
  echo -e "3. コンテナ内環境変数を確認"
fi

# クリーンアップ
rm -f /tmp/debug-task-definition-*.json

echo -e "${GREEN}Phase 1デバッグ完了${NC}"