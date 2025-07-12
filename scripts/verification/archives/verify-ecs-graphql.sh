#!/bin/bash
# scripts/verification/verify-ecs-graphql.sh
# GraphQLサービスのデプロイ検証

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ENVIRONMENT=${1:-development}
SERVICE_NAME="${ENVIRONMENT}-graphql-service"
CLUSTER_NAME="${ENVIRONMENT}-cluster"
MAX_RETRIES=20
RETRY_INTERVAL=15

echo -e "${BLUE}GraphQLサービス(${SERVICE_NAME})のデプロイ検証を開始します...${NC}"

# 1. ECSサービスの状態確認
echo "ECSサービスのステータスを確認しています..."
SERVICE_DETAILS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME})
SERVICE_STATUS=$(echo ${SERVICE_DETAILS} | jq -r '.services[0].status')

if [ "${SERVICE_STATUS}" != "ACTIVE" ]; then
  echo -e "${RED}エラー: サービスがアクティブではありません (Status: ${SERVICE_STATUS})${NC}"
  exit 1
fi

# 2. デプロイ状態の確認
DEPLOYMENT_STATUS=$(echo ${SERVICE_DETAILS} | jq -r '.services[0].deployments[0].status')
if [ "${DEPLOYMENT_STATUS}" != "PRIMARY" ]; then
  echo -e "${YELLOW}警告: 最新のデプロイがプライマリではありません (Status: ${DEPLOYMENT_STATUS})${NC}"
fi

# 3. 実行中のタスク数の確認
RUNNING_TASKS=$(echo ${SERVICE_DETAILS} | jq -r '.services[0].runningCount')
DESIRED_TASKS=$(echo ${SERVICE_DETAILS} | jq -r '.services[0].desiredCount')

if [ ${RUNNING_TASKS} -lt ${DESIRED_TASKS} ]; then
  echo -e "${YELLOW}警告: 実行中のタスク数が期待値よりも少ないです (実行中: ${RUNNING_TASKS}, 期待値: ${DESIRED_TASKS})${NC}"
  
  # タスク起動失敗の理由を確認
  TASKS=$(aws ecs list-tasks --cluster ${CLUSTER_NAME} --service-name ${SERVICE_NAME} --query 'taskArns' --output text)
  if [ -n "${TASKS}" ]; then
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
  fi
else
  echo -e "${GREEN}✓ タスク数は期待通りです (実行中: ${RUNNING_TASKS}, 期待値: ${DESIRED_TASKS})${NC}"
fi

# 4. ALBターゲットグループのヘルスチェック状態を確認
# ALBのARNを取得
LOAD_BALANCER_NAME="${ENVIRONMENT}-alb"
ALB_ARN=$(aws elbv2 describe-load-balancers --names ${LOAD_BALANCER_NAME} --query 'LoadBalancers[0].LoadBalancerArn' --output text)

if [ -z "${ALB_ARN}" ] || [ "${ALB_ARN}" == "None" ]; then
  echo -e "${YELLOW}警告: ALB ${LOAD_BALANCER_NAME} が見つかりません。ヘルスチェック状態を確認できません。${NC}"
else
  # ターゲットグループのARNを取得
  TARGET_GROUP_ARN=$(aws elbv2 describe-target-groups --load-balancer-arn ${ALB_ARN} --query 'TargetGroups[?contains(TargetGroupName, `graphql`)].TargetGroupArn' --output text)
  
  if [ -z "${TARGET_GROUP_ARN}" ] || [ "${TARGET_GROUP_ARN}" == "None" ]; then
    echo -e "${YELLOW}警告: GraphQLターゲットグループが見つかりません。ヘルスチェック状態を確認できません。${NC}"
  else
    echo "ALBターゲットグループのヘルスチェック状態を確認しています..."
    
    # 健全/不健全なターゲットの数を確認
    HEALTHY_COUNT=0
    UNHEALTHY_COUNT=0
    RETRIES=0
    
    while [ ${HEALTHY_COUNT} -lt ${DESIRED_TASKS} ] && [ ${RETRIES} -lt ${MAX_RETRIES} ]; do
      HEALTH_STATUS=$(aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN})
      HEALTHY_COUNT=$(echo ${HEALTH_STATUS} | jq -r '.TargetHealthDescriptions[] | select(.TargetHealth.State=="healthy") | .TargetHealth.State' | wc -l)
      UNHEALTHY_COUNT=$(echo ${HEALTH_STATUS} | jq -r '.TargetHealthDescriptions[] | select(.TargetHealth.State!="healthy") | .TargetHealth.State' | wc -l)
      
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
      echo "不健全なターゲットの詳細:"
      echo ${HEALTH_STATUS} | jq -r '.TargetHealthDescriptions[] | select(.TargetHealth.State!="healthy") | {id: .Target.Id, port: .Target.Port, state: .TargetHealth.State, reason: .TargetHealth.Reason, description: .TargetHealth.Description}'
    else
      echo -e "${GREEN}✓ すべてのターゲットが健全です (${HEALTHY_COUNT}/${DESIRED_TASKS})${NC}"
    fi
  fi
fi

# 5. GraphQLエンドポイントをテスト
echo "GraphQLエンドポイントをテストしています..."

# ALBのDNS名を取得
ALB_DNS=$(aws elbv2 describe-load-balancers --names ${LOAD_BALANCER_NAME} --query 'LoadBalancers[0].DNSName' --output text)

if [ -z "${ALB_DNS}" ] || [ "${ALB_DNS}" == "None" ]; then
  echo -e "${YELLOW}警告: ALB ${LOAD_BALANCER_NAME} のDNS名が取得できません。GraphQLエンドポイントをテストできません。${NC}"
else
  # GraphQLエンドポイントを呼び出し
  GRAPHQL_URL="http://${ALB_DNS}/graphql"
  QUERY='{"query":"{ __schema { queryType { name } } }"}'

  HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Content-Type: application/json" -d "${QUERY}" ${GRAPHQL_URL})

  if [ "${HTTP_STATUS}" == "200" ]; then
    echo -e "${GREEN}✓ GraphQLエンドポイントが正常に応答しました (HTTP ${HTTP_STATUS})${NC}"
    
    # スキーマが取得できるか詳細に確認
    RESPONSE=$(curl -s -H "Content-Type: application/json" -d "${QUERY}" ${GRAPHQL_URL})
    if echo ${RESPONSE} | jq -e '.data.__schema.queryType.name' > /dev/null; then
      echo -e "${GREEN}✓ GraphQLスキーマが正常に取得できました${NC}"
      
      # テストスイート一覧を取得してみる
      LIST_QUERY='{"query":"{ testSuites { edges { node { id name } } totalCount } }"}'
      LIST_RESPONSE=$(curl -s -H "Content-Type: application/json" -d "${LIST_QUERY}" ${GRAPHQL_URL})
      
      if echo ${LIST_RESPONSE} | jq -e '.data.testSuites' > /dev/null; then
        TOTAL_COUNT=$(echo ${LIST_RESPONSE} | jq -r '.data.testSuites.totalCount')
        echo -e "${GREEN}✓ テストスイート一覧クエリが成功しました (総数: ${TOTAL_COUNT})${NC}"
      else
        echo -e "${YELLOW}警告: テストスイート一覧クエリに失敗しました${NC}"
        echo "応答内容:"
        echo ${LIST_RESPONSE} | jq '.'
      fi
    else
      echo -e "${RED}エラー: GraphQLスキーマが取得できませんでした${NC}"
      echo "応答内容:"
      echo ${RESPONSE} | jq '.'
    fi
  else
    echo -e "${RED}エラー: GraphQLエンドポイントが正常に応答しません (HTTP ${HTTP_STATUS})${NC}"
    echo "詳細な応答を確認します:"
    curl -v -H "Content-Type: application/json" -d "${QUERY}" ${GRAPHQL_URL}
  fi
fi

echo -e "${BLUE}GraphQLサービスのデプロイ検証が完了しました${NC}"