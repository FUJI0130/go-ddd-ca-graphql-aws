#!/bin/bash
# scripts/verification/verify-ecs-api.sh
# REST APIサービスのデプロイ検証

set -e

# 共通ユーティリティのインポート
source "$(dirname "$0")/../common/json_utils.sh"

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ENVIRONMENT=${1:-development}
SERVICE_NAME="${ENVIRONMENT}-api"
CLUSTER_NAME="${ENVIRONMENT}-shared-cluster"
MAX_RETRIES=20
RETRY_INTERVAL=15

echo -e "${BLUE}REST APIサービス(${SERVICE_NAME})のデプロイ検証を開始します...${NC}"

# 1. ECSサービスの存在確認
echo "ECSサービスの存在を確認しています..."
# デバッグ情報の追加例
echo "DEBUG: Using cluster '${CLUSTER_NAME}' and service '${SERVICE_NAME}'"
SERVICE_COUNT=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'length(services)' --output text)

if [ "${SERVICE_COUNT}" = "0" ]; then
  echo -e "${RED}エラー: サービス ${SERVICE_NAME} が存在しません${NC}"
  exit 1
fi

# 2. ECSサービスの状態確認
echo "ECSサービスのステータスを確認しています..."
SERVICE_STATUS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].status' --output text)

if [ "${SERVICE_STATUS}" != "ACTIVE" ]; then
  echo -e "${RED}エラー: サービスがアクティブではありません (Status: ${SERVICE_STATUS})${NC}"
  exit 1
fi

# 3. デプロイ状態の確認
DEPLOYMENT_STATUS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].deployments[0].status' --output text)
if [ "${DEPLOYMENT_STATUS}" != "PRIMARY" ]; then
  echo -e "${YELLOW}警告: 最新のデプロイがプライマリではありません (Status: ${DEPLOYMENT_STATUS})${NC}"
fi

# 4. 実行中のタスク数の確認
RUNNING_TASKS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].runningCount' --output text)
DESIRED_TASKS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].desiredCount' --output text)

if [ ${RUNNING_TASKS} -lt ${DESIRED_TASKS} ]; then
  echo -e "${YELLOW}警告: 実行中のタスク数が期待値よりも少ないです (実行中: ${RUNNING_TASKS}, 期待値: ${DESIRED_TASKS})${NC}"
  
  # タスク起動失敗の理由を確認
  TASKS=$(aws ecs list-tasks --cluster ${CLUSTER_NAME} --service-name ${SERVICE_NAME} --query 'taskArns' --output text)
  if [ -n "${TASKS}" ]; then
    # タスク詳細を取得
    if has_jq; then
      # jqが利用可能な場合
      TASK_DETAILS=$(aws ecs describe-tasks --cluster ${CLUSTER_NAME} --tasks ${TASKS})
      FAILURES=$(echo ${TASK_DETAILS} | jq -r '.failures')
      if [ "${FAILURES}" != "[]" ]; then
        echo -e "${RED}タスク起動に失敗しています:${NC}"
        echo ${FAILURES} | jq '.'
      fi
      
      # 停止したタスクの理由も確認
      STOPPED_REASON=$(echo ${TASK_DETAILS} | jq -r '.tasks[].stoppedReason')
      if [ -n "${STOPPED_REASON}" ] && [ "${STOPPED_REASON}" != "null" ]; then
        echo -e "${RED}タスクが停止した理由:${NC}"
        echo "${STOPPED_REASON}"
      fi
    else
      # jqが利用できない場合
      # 単純な存在確認とエラーメッセージの表示のみ
      FAILURES=$(aws ecs describe-tasks --cluster ${CLUSTER_NAME} --tasks ${TASKS} --query 'failures' --output text)
      if [ -n "${FAILURES}" ] && [ "${FAILURES}" != "None" ]; then
        echo -e "${RED}タスク起動に失敗しています:${NC}"
        echo "${FAILURES}"
      fi
      
      # 停止理由の確認
      STOPPED_REASON=$(aws ecs describe-tasks --cluster ${CLUSTER_NAME} --tasks ${TASKS} --query 'tasks[].stoppedReason' --output text)
      if [ -n "${STOPPED_REASON}" ] && [ "${STOPPED_REASON}" != "None" ]; then
        echo -e "${RED}タスクが停止した理由:${NC}"
        echo "${STOPPED_REASON}"
      fi
    fi
  fi
else
  echo -e "${GREEN}✓ タスク数は期待通りです (実行中: ${RUNNING_TASKS}, 期待値: ${DESIRED_TASKS})${NC}"
fi

# 5. ALBターゲットグループのヘルスチェック状態を確認
# ALBのARNを取得
LOAD_BALANCER_NAME="${ENVIRONMENT}-api-alb"
ALB_ARN=$(aws elbv2 describe-load-balancers --names ${LOAD_BALANCER_NAME} --query 'LoadBalancers[0].LoadBalancerArn' --output text)

if [ -z "${ALB_ARN}" ] || [ "${ALB_ARN}" == "None" ]; then
  echo -e "${YELLOW}警告: ALB ${LOAD_BALANCER_NAME} が見つかりません。ヘルスチェック状態を確認できません。${NC}"
else
  # ターゲットグループのARNを取得
  TARGET_GROUP_ARN=$(aws elbv2 describe-target-groups --load-balancer-arn ${ALB_ARN} --query 'TargetGroups[?contains(TargetGroupName, `api`)].TargetGroupArn' --output text)
  
  if [ -z "${TARGET_GROUP_ARN}" ] || [ "${TARGET_GROUP_ARN}" == "None" ]; then
    echo -e "${YELLOW}警告: APIターゲットグループが見つかりません。ヘルスチェック状態を確認できません。${NC}"
  else
    echo "ALBターゲットグループのヘルスチェック状態を確認しています..."
    
    # 健全/不健全なターゲットの数を確認
    HEALTHY_COUNT=0
    UNHEALTHY_COUNT=0
    RETRIES=0
    
    while [ ${HEALTHY_COUNT} -lt ${DESIRED_TASKS} ] && [ ${RETRIES} -lt ${MAX_RETRIES} ]; do
      # ヘルスチェック状態を取得
      if has_jq; then
        # jqが利用可能な場合
        HEALTH_STATUS=$(aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN})
        HEALTHY_COUNT=$(echo ${HEALTH_STATUS} | jq -r '.TargetHealthDescriptions[] | select(.TargetHealth.State=="healthy") | .TargetHealth.State' | wc -l)
        UNHEALTHY_COUNT=$(echo ${HEALTH_STATUS} | jq -r '.TargetHealthDescriptions[] | select(.TargetHealth.State!="healthy") | .TargetHealth.State' | wc -l)
      else
        # jqが利用できない場合 - AWS CLIのクエリを使用して健全なターゲット数を取得
        # 注: この方法は簡易的なもので、完全なjqの代替にはなりません
        HEALTH_STATUS=$(aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN})
        
        # 健全なターゲット数をカウント
        HEALTHY_TARGETS=$(aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN} \
                         --query 'length(TargetHealthDescriptions[?TargetHealth.State==`healthy`])' --output text)
        HEALTHY_COUNT=${HEALTHY_TARGETS:-0}
        
        # 不健全なターゲット数をカウント
        UNHEALTHY_TARGETS=$(aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN} \
                          --query 'length(TargetHealthDescriptions[?TargetHealth.State!=`healthy`])' --output text)
        UNHEALTHY_COUNT=${UNHEALTHY_TARGETS:-0}
      fi
      
      if [ ${HEALTHY_COUNT} -lt ${DESIRED_TASKS} ]; then
        RETRIES=$((RETRIES+1))
        echo "健全なターゲット: ${HEALTHY_COUNT}/${DESIRED_TASKS} (リトライ: ${RETRIES}/${MAX_RETRIES})"
        
        if [ ${RETRIES} -lt ${MAX_RETRIES} ]; then
          echo "ヘルスチェックが完了するまで ${RETRY_INTERVAL} 秒待機します..."
          sleep ${RETRY_INTERVAL}
        fi
      else
        break
      fi
    done
    
    if [ ${HEALTHY_COUNT} -lt ${DESIRED_TASKS} ]; then
      echo -e "${RED}警告: すべてのターゲットが健全ではありません (健全: ${HEALTHY_COUNT}, 不健全: ${UNHEALTHY_COUNT})${NC}"
      
      # 不健全なターゲットの詳細を表示
      if has_jq; then
        echo "不健全なターゲットの詳細:"
        echo ${HEALTH_STATUS} | jq -r '.TargetHealthDescriptions[] | select(.TargetHealth.State!="healthy") | {id: .Target.Id, port: .Target.Port, state: .TargetHealth.State, reason: .TargetHealth.Reason, description: .TargetHealth.Description}'
      else
        echo "不健全なターゲットの詳細を取得するにはjqが必要です。"
        echo "代わりに基本情報を表示します:"
        
        # 基本的な不健全ターゲット情報を表示
        aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN} \
          --query 'TargetHealthDescriptions[?TargetHealth.State!=`healthy`].[Target.Id, Target.Port, TargetHealth.State, TargetHealth.Reason]' \
          --output text
      fi
    else
      echo -e "${GREEN}✓ すべてのターゲットが健全です (${HEALTHY_COUNT}/${DESIRED_TASKS})${NC}"
    fi
  fi
fi

# 6. アプリケーションのヘルスチェックエンドポイントをテスト
echo "アプリケーションのヘルスチェックエンドポイントをテストしています..."

# ALBのDNS名を取得
ALB_DNS=$(aws elbv2 describe-load-balancers --names ${LOAD_BALANCER_NAME} --query 'LoadBalancers[0].DNSName' --output text)

if [ -z "${ALB_DNS}" ] || [ "${ALB_DNS}" == "None" ]; then
  echo -e "${YELLOW}警告: ALB ${LOAD_BALANCER_NAME} のDNS名が取得できません。APIエンドポイントをテストできません。${NC}"
else
  # ヘルスチェックエンドポイントを呼び出し
  API_HEALTH_URL="http://${ALB_DNS}/health"
  HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" ${API_HEALTH_URL})
  
  if [ "${HTTP_STATUS}" == "200" ]; then
    echo -e "${GREEN}✓ APIのヘルスチェックエンドポイントが正常に応答しました (HTTP ${HTTP_STATUS})${NC}"
  else
    echo -e "${RED}エラー: APIのヘルスチェックエンドポイントが正常に応答しません (HTTP ${HTTP_STATUS})${NC}"
    echo "詳細な応答を確認します:"
    curl -v ${API_HEALTH_URL}
  fi
fi

echo -e "${BLUE}REST APIサービスのデプロイ検証が完了しました${NC}"