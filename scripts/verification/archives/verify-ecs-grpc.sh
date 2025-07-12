#!/bin/bash
# scripts/verification/verify-ecs-grpc.sh
# gRPCサービスのデプロイ検証

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ENVIRONMENT=${1:-development}
SERVICE_NAME="${ENVIRONMENT}-grpc-service"
CLUSTER_NAME="${ENVIRONMENT}-cluster"
MAX_RETRIES=20
RETRY_INTERVAL=15

echo -e "${BLUE}gRPCサービス(${SERVICE_NAME})のデプロイ検証を開始します...${NC}"

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
  TARGET_GROUP_ARN=$(aws elbv2 describe-target-groups --load-balancer-arn ${ALB_ARN} --query 'TargetGroups[?contains(TargetGroupName, `grpc`)].TargetGroupArn' --output text)
  
  if [ -z "${TARGET_GROUP_ARN}" ] || [ "${TARGET_GROUP_ARN}" == "None" ]; then
    echo -e "${YELLOW}警告: gRPCターゲットグループが見つかりません。ヘルスチェック状態を確認できません。${NC}"
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

# 5. gRPCサービスをテスト
echo "gRPCサービスをテストしています..."

# ALBのDNS名を取得
ALB_DNS=$(aws elbv2 describe-load-balancers --names ${LOAD_BALANCER_NAME} --query 'LoadBalancers[0].DNSName' --output text)

if [ -z "${ALB_DNS}" ] || [ "${ALB_DNS}" == "None" ]; then
  echo -e "${YELLOW}警告: ALB ${LOAD_BALANCER_NAME} のDNS名が取得できません。gRPCサービスをテストできません。${NC}"
else
  # gRPCサービスをテスト
  if command -v grpcurl &> /dev/null; then
    echo "grpcurlを使用してgRPCサービスをテストしています..."
    GRPC_ENDPOINT="${ALB_DNS}:50051"
    
    # サービスの一覧を取得
    echo "利用可能なサービスを確認しています..."
    if SERVICES=$(grpcurl -plaintext ${GRPC_ENDPOINT} list 2>/dev/null); then
      if [ -z "${SERVICES}" ]; then
        echo -e "${RED}エラー: gRPCサービスが応答しましたが、サービスが見つかりません${NC}"
      else
        echo -e "${GREEN}✓ gRPCサービスが応答しました。利用可能なサービス:${NC}"
        echo "${SERVICES}"
        
        # TestSuiteServiceが存在するか確認
        if echo "${SERVICES}" | grep -q "testsuite.v1.TestSuiteService"; then
          echo -e "${GREEN}✓ TestSuiteServiceが見つかりました${NC}"
          
          # メソッド一覧を取得
          METHODS=$(grpcurl -plaintext ${GRPC_ENDPOINT} list testsuite.v1.TestSuiteService)
          echo "利用可能なメソッド:"
          echo "${METHODS}"
          
          # ListTestSuitesメソッドをテスト
          if echo "${METHODS}" | grep -q "ListTestSuites"; then
            echo "ListTestSuitesメソッドをテストしています..."
            if LIST_RESULT=$(grpcurl -plaintext -d '{"limit": 10}' ${GRPC_ENDPOINT} testsuite.v1.TestSuiteService/ListTestSuites 2>/dev/null); then
              echo -e "${GREEN}✓ ListTestSuitesメソッドが正常に応答しました${NC}"
              echo "応答内容の一部:"
              echo "${LIST_RESULT}" | jq -r '. | {totalCount: .totalCount}' 2>/dev/null || echo "${LIST_RESULT:0:100}..."
            else
              echo -e "${RED}エラー: ListTestSuitesメソッドの呼び出しに失敗しました${NC}"
            fi
          else
            echo -e "${YELLOW}警告: ListTestSuitesメソッドが見つかりません${NC}"
          fi
        else
          echo -e "${RED}エラー: TestSuiteServiceが見つかりません${NC}"
        fi
      fi
    else
      echo -e "${RED}エラー: gRPCサービスへの接続に失敗しました${NC}"
    fi
  else
    echo -e "${YELLOW}警告: grpcurlコマンドが見つかりません。gRPCサービスをテストできません${NC}"
    echo "代替手段でgRPCサービスのポートが開いているか確認します..."
    
    # ポートが開いているか確認
    if command -v nc &> /dev/null; then
      if nc -z -w 5 ${ALB_DNS} 50051; then
        echo -e "${GREEN}✓ gRPCポート(50051)が開いています${NC}"
      else
        echo -e "${RED}エラー: gRPCポート(50051)が開いていないか、接続できません${NC}"
      fi
    else
      echo -e "${YELLOW}警告: ncコマンドが見つかりません。ポート接続を確認できません${NC}"
      echo "トラブルシューティングのため、curlでHTTP/2接続を試みます..."
      
      # curlによるHTTP/2接続確認
      if curl -I --http2 -s http://${ALB_DNS}:50051 &>/dev/null; then
        echo -e "${GREEN}✓ HTTP/2接続が可能です${NC}"
      else
        echo -e "${RED}エラー: HTTP/2接続に失敗しました${NC}"
      fi
    fi
  fi
fi

echo -e "${BLUE}gRPCサービスのデプロイ検証が完了しました${NC}"