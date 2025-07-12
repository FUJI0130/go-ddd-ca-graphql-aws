#!/bin/bash

# AWS環境のコスト見積もりスクリプト
TF_ENV=${1:-development}
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." >/dev/null 2>&1 && pwd )"

# カラーコード
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========== AWS環境のコスト見積もり (${TF_ENV}) ==========${NC}"

# RDSインスタンスの状態確認
echo -e "\n${YELLOW}RDSインスタンスの状態:${NC}"
aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='${TF_ENV}-postgres'].{ID:DBInstanceIdentifier,Status:DBInstanceStatus,Class:DBInstanceClass}" --output table

# ECSサービスの状態確認
echo -e "\n${YELLOW}ECSサービスの状態:${NC}"
aws ecs list-services --cluster ${TF_ENV}-shared-cluster --query "serviceArns" --output text 2>/dev/null | xargs -r -I{} aws ecs describe-services --cluster ${TF_ENV}-shared-cluster --services {} --query "services[].{Name:serviceName,Status:status,RunningTasks:runningCount,DesiredTasks:desiredCount}" --output table

# ロードバランサーの状態確認
echo -e "\n${YELLOW}ロードバランサーの状態:${NC}"
aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName,'${TF_ENV}')].{Name:LoadBalancerName,Type:Type,State:State.Code}" --output table

# 概算コスト計算
echo -e "\n${YELLOW}概算コスト（日額）:${NC}"
echo "注: これは非常に概算の見積もりです。実際の請求額は異なる場合があります。"

# RDS見積もり
RDS_COUNT=$(aws rds describe-db-instances --query "length(DBInstances[?DBInstanceIdentifier=='${TF_ENV}-postgres'])" --output text)
if [ "$RDS_COUNT" -gt 0 ]; then
  RDS_CLASS=$(aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='${TF_ENV}-postgres'].DBInstanceClass" --output text)
  if [[ "$RDS_CLASS" == "db.t3.small" ]]; then
    RDS_COST=0.034 # USD/時間
  elif [[ "$RDS_CLASS" == "db.t3.micro" ]]; then
    RDS_COST=0.017 # USD/時間
  else
    RDS_COST=0.05 # デフォルト見積もり
  fi
  RDS_DAILY=$(echo "$RDS_COST * 24" | bc -l | xargs printf "%.2f")
  echo -e "RDS ($RDS_CLASS): 約 $RDS_DAILY USD/日"
fi

# ALB見積もり
ALB_COUNT=$(aws elbv2 describe-load-balancers --query "length(LoadBalancers[?contains(LoadBalancerName,'${TF_ENV}')])" --output text)
if [ "$ALB_COUNT" -gt 0 ]; then
  ALB_COST=0.0225 # USD/時間
  ALB_DAILY=$(echo "$ALB_COST * 24 * $ALB_COUNT" | bc -l | xargs printf "%.2f")
  echo -e "ALB ($ALB_COUNT 個): 約 $ALB_DAILY USD/日"
fi

# Fargate見積もり
TASK_COUNT=$(aws ecs list-tasks --cluster ${TF_ENV}-shared-cluster --query "length(taskArns)" --output text 2>/dev/null || echo 0)
if [ "$TASK_COUNT" -gt 0 ]; then
  FARGATE_COST=$(echo "0.02 * $TASK_COUNT" | bc -l) # 1タスクあたり約0.02 USD/時間と仮定
  FARGATE_DAILY=$(echo "$FARGATE_COST * 24" | bc -l | xargs printf "%.2f")
  echo -e "ECS Fargate ($TASK_COUNT タスク): 約 $FARGATE_DAILY USD/日"
fi

# NAT Gateway
NAT_COUNT=$(aws ec2 describe-nat-gateways --filter "Name=state,Values=available" --query "length(NatGateways[?contains(Tags[?Key=='Environment'].Value,'${TF_ENV}')])" --output text 2>/dev/null || echo 0)
if [ "$NAT_COUNT" -gt 0 ]; then
  NAT_COST=0.045 # USD/時間
  NAT_DAILY=$(echo "$NAT_COST * 24 * $NAT_COUNT" | bc -l | xargs printf "%.2f")
  echo -e "NAT Gateway ($NAT_COUNT 個): 約 $NAT_DAILY USD/日"
fi

# 合計見積もり
TOTAL_DAILY=$(echo "${RDS_DAILY:-0} + ${ALB_DAILY:-0} + ${FARGATE_DAILY:-0} + ${NAT_DAILY:-0}" | bc -l | xargs printf "%.2f")
TOTAL_MONTHLY=$(echo "$TOTAL_DAILY * 30" | bc -l | xargs printf "%.2f")
echo -e "\n${GREEN}見積もり合計: 約 $TOTAL_DAILY USD/日${NC}"
echo -e "${GREEN}見積もり合計: 約 $TOTAL_MONTHLY USD/月${NC}"

echo -e "\n${BLUE}========== コスト削減のためのアクション ==========${NC}"
echo -e "1. 使用していないECSサービスを削除: make pause-api-dev TF_ENV=${TF_ENV}"
echo -e "2. すべてのリソースを削除: make stop-api-dev TF_ENV=${TF_ENV}"
echo -e "3. 状態を検証してからリソース削除: make terraform-cleanup-safe TF_ENV=${TF_ENV}"
echo -e "4. 状態不整合時のタグベース削除: make tag-cleanup TF_ENV=${TF_ENV}"
echo -e "\n${YELLOW}注意: 上記のコマンドはTerraform状態も更新します。環境の整合性を保つために推奨されます。${NC}"