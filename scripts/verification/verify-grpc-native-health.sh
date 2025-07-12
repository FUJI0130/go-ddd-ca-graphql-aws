#!/bin/bash
# ===================================================================
# ファイル名: verify-grpc-native-health.sh
# 説明: gRPCネイティブヘルスチェックを検証するスクリプト
# 
# 用途:
#  - gRPCヘルスチェックプロトコルを使用してヘルスチェックを検証
#  - ターゲットグループのヘルス状態を確認
# 
# 要件:
#  - grpcurlがインストールされていること
#  - AWS CLIがインストールされていること
# 
# 使用方法:
#  ./verify-grpc-native-health.sh [環境名]
#
# 引数:
#  環境名 - 検証する環境（development, production）、省略時はdevelopment
# ===================================================================

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 環境引数の解析
ENVIRONMENT=${1:-development}

# 環境設定
ENV_PREFIX="${ENVIRONMENT}"
MAX_RETRIES=5
RETRY_INTERVAL=10

echo -e "${BLUE}gRPCネイティブヘルスチェックを検証しています (環境: ${ENVIRONMENT})...${NC}"

# grpcurlコマンドの存在確認
if ! command -v grpcurl &> /dev/null; then
  echo -e "${RED}エラー: grpcurlコマンドが見つかりません${NC}"
  echo "grpcurlのインストール方法: https://github.com/fullstorydev/grpcurl#installation"
  exit 1
fi

# AWS認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  echo -e "${RED}エラー: AWS認証情報が設定されていないか、無効です${NC}"
  echo "AWS CLIの設定を確認してください: aws configure"
  exit 1
fi

# ALBのDNS名を取得
# ALB_NAME="${ENV_PREFIX}-grpc-alb"
ALB_NAME="${ENV_PREFIX}-grpc-new-alb"
ALB_DNS=$(aws elbv2 describe-load-balancers --names ${ALB_NAME} --query "LoadBalancers[0].DNSName" --output text 2>/dev/null || echo "")

if [ -z "${ALB_DNS}" ] || [ "${ALB_DNS}" == "None" ]; then
  echo -e "${RED}エラー: ALB ${ALB_NAME} が見つかりません${NC}"
  exit 1
fi

echo -e "${BLUE}ALB DNS名: ${ALB_DNS}${NC}"

# ターゲットグループのARNを取得
# TG_NAME="${ENV_PREFIX}-grpc-grpc-native-tg"
TG_NAME="${ENV_PREFIX}-grpc-native-new-tg"
TG_ARN=$(aws elbv2 describe-target-groups --names ${TG_NAME} --query "TargetGroups[0].TargetGroupArn" --output text 2>/dev/null || echo "")

if [ -z "${TG_ARN}" ] || [ "${TG_ARN}" == "None" ]; then
  echo -e "${RED}エラー: ターゲットグループ ${TG_NAME} が見つかりません${NC}"
  exit 1
fi

# gRPCヘルスチェックの確認（複数回試行）
echo -e "${BLUE}gRPCヘルスチェックを検証しています...${NC}"

for i in $(seq 1 ${MAX_RETRIES}); do
  echo -e "${YELLOW}試行 ${i}/${MAX_RETRIES}...${NC}"
  
  # grpcurlを使用してヘルスチェック（TLS設定でHTTPS接続）
  GRPC_OUTPUT=$(grpcurl -insecure -proto proto/testsuite/v1/health.proto -authority ${ALB_DNS} ${ALB_DNS}:443 grpc.health.v1.Health/Check 2>&1 || echo "FAILED")
  
  if [[ "${GRPC_OUTPUT}" == *"SERVING"* ]]; then
    echo -e "${GREEN}✓ gRPCヘルスチェックが成功しました${NC}"
    echo -e "${BLUE}レスポンス:${NC}"
    echo "${GRPC_OUTPUT}"
    break
  else
    echo -e "${YELLOW}✗ gRPCヘルスチェックが失敗しました${NC}"
    echo -e "${YELLOW}出力: ${GRPC_OUTPUT}${NC}"
    
    if [ $i -lt ${MAX_RETRIES} ]; then
      echo -e "${YELLOW}${RETRY_INTERVAL}秒後に再試行します...${NC}"
      sleep ${RETRY_INTERVAL}
    else
      echo -e "${RED}最大試行回数に達しました。ヘルスチェックに失敗しました。${NC}"
    fi
  fi
done

# ターゲットグループのヘルス状態を確認
echo -e "${BLUE}ターゲットグループのヘルス状態を確認しています...${NC}"
TARGETS_HEALTH=$(aws elbv2 describe-target-health --target-group-arn ${TG_ARN} --query "TargetHealthDescriptions[].TargetHealth.State" --output text 2>/dev/null || echo "")

if [ -z "${TARGETS_HEALTH}" ]; then
  echo -e "${YELLOW}警告: ターゲットグループにターゲットが登録されていません${NC}"
elif [[ "${TARGETS_HEALTH}" == *"healthy"* ]]; then
  echo -e "${GREEN}✓ 少なくとも1つのターゲットがhealthy状態です${NC}"
  
  # 詳細なターゲット情報
  echo -e "${BLUE}詳細なターゲット状態:${NC}"
  aws elbv2 describe-target-health --target-group-arn ${TG_ARN} --query "TargetHealthDescriptions[].[Target.Id, TargetHealth.State, TargetHealth.Reason]" --output text | while read -r line; do
    TARGET_ID=$(echo $line | cut -d' ' -f1)
    HEALTH_STATE=$(echo $line | cut -d' ' -f2)
    REASON=$(echo $line | cut -d' ' -f3-)
    
    if [ "${HEALTH_STATE}" == "healthy" ]; then
      echo -e "  ${GREEN}✓ ${TARGET_ID}: ${HEALTH_STATE}${NC}"
    else
      echo -e "  ${YELLOW}✗ ${TARGET_ID}: ${HEALTH_STATE} (理由: ${REASON})${NC}"
    fi
  done
else
  echo -e "${RED}✗ ヘルシーなターゲットが見つかりません${NC}"
  
  # 詳細なターゲット情報
  echo -e "${BLUE}詳細なターゲット状態:${NC}"
  aws elbv2 describe-target-health --target-group-arn ${TG_ARN} --query "TargetHealthDescriptions[].[Target.Id, TargetHealth.State, TargetHealth.Reason]" --output text | while read -r line; do
    TARGET_ID=$(echo $line | cut -d' ' -f1)
    HEALTH_STATE=$(echo $line | cut -d' ' -f2)
    REASON=$(echo $line | cut -d' ' -f3-)
    
    echo -e "  ${YELLOW}✗ ${TARGET_ID}: ${HEALTH_STATE} (理由: ${REASON})${NC}"
  done
  
  exit 1
fi

# ECSサービスの状態確認
echo -e "${BLUE}ECSサービスの状態を確認しています...${NC}"
CLUSTER_NAME="${ENV_PREFIX}-shared-cluster"
# SERVICE_NAME="${ENV_PREFIX}-grpc"
SERVICE_NAME="${ENV_PREFIX}-grpc-new"

SERVICE_STATUS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query "services[0].status" --output text 2>/dev/null || echo "")
RUNNING_COUNT=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query "services[0].runningCount" --output text 2>/dev/null || echo "0")
DESIRED_COUNT=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query "services[0].desiredCount" --output text 2>/dev/null || echo "0")

if [ "${SERVICE_STATUS}" == "ACTIVE" ]; then
  echo -e "${GREEN}✓ ECSサービスはアクティブです${NC}"
  echo -e "  - 実行中タスク: ${RUNNING_COUNT}/${DESIRED_COUNT}"
  
  if [ "${RUNNING_COUNT}" -lt "${DESIRED_COUNT}" ]; then
    echo -e "${YELLOW}警告: 実行中タスク数が希望数より少ないです${NC}"
  fi
else
  echo -e "${RED}✗ ECSサービスが見つからないか、アクティブではありません (状態: ${SERVICE_STATUS})${NC}"
  exit 1
fi

echo -e "${GREEN}gRPCネイティブヘルスチェック検証が完了しました${NC}"
exit 0