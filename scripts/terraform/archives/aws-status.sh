#!/bin/bash
# ===================================================================
# ファイル名: aws-status.sh
# 説明: AWS環境のリソース状態を確認するスクリプト
# 
# 用途:
#  - AWS CLIを使用してAWS環境のリソース状態を確認
#  - VPC、RDS、ECS、ロードバランサー、ターゲットグループなどの情報収集
#  - Terraformリモートステート関連リソースの確認
#  - サービスタイプ別（API、GraphQL、gRPC）の状態確認
# 
# 注意:
#  - このスクリプトはAWS環境のみに作用し、Terraformの状態には影響を与えません
#  - AWS CLIの適切な権限が必要です
#  - 既存のリソース確認のため、リソースの変更は行いません
# 
# 使用方法:
#  ./aws-status.sh [環境名]
#
# 引数:
#  環境名 - 確認する環境（development, production）、省略時はdevelopment
# ===================================================================
source $(dirname "$0")/common/aws_resource_utils.sh
set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 環境引数の解析
ENVIRONMENT=${1:-development}  # 引数がない場合はデフォルトでdevelopment

# 環境設定
case "${ENVIRONMENT}" in
  development)
    TF_DIR="deployments/terraform/environments/development"
    ENV_PREFIX="development"
    ;;
  production)
    TF_DIR="deployments/terraform/environments/production"
    ENV_PREFIX="production"
    ;;
  *)
    echo -e "${RED}エラー: サポートされていない環境です: ${ENVIRONMENT}${NC}"
    echo "サポートされている環境: development, production"
    exit 1
    ;;
esac

# サービスタイプのリスト
SERVICE_TYPES=("api" "graphql" "grpc")

STATE_BUCKET="test-management-terraform-state"
STATE_DYNAMODB="test-management-terraform-lock"
AWS_REGION="ap-northeast-1"

echo -e "${BLUE}${ENVIRONMENT}環境のAWSインフラストラクチャの状態を確認しています...${NC}"

# AWS認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  echo -e "${RED}エラー: AWS認証情報が設定されていないか、無効です${NC}"
  echo "AWS CLIの設定を確認してください: aws configure"
  exit 1
fi

# S3バケットの確認
if aws s3api head-bucket --bucket "${STATE_BUCKET}" 2>/dev/null; then
  echo -e "${GREEN}✓ リモートステートバケット (${STATE_BUCKET}) は存在します${NC}"
else
  echo -e "${YELLOW}✗ リモートステートバケット (${STATE_BUCKET}) は存在しません${NC}"
fi

# DynamoDBテーブルの確認
if aws dynamodb describe-table --table-name "${STATE_DYNAMODB}" --region "${AWS_REGION}" &>/dev/null; then
  echo -e "${GREEN}✓ ステートロックテーブル (${STATE_DYNAMODB}) は存在します${NC}"
else
  echo -e "${YELLOW}✗ ステートロックテーブル (${STATE_DYNAMODB}) は存在しません${NC}"
fi

# バックエンド設定の確認
if grep -q "backend \"s3\"" "${TF_DIR}/main.tf" && ! grep -q "# backend \"s3\"" "${TF_DIR}/main.tf"; then
  echo -e "${GREEN}✓ リモートバックエンド設定は有効です${NC}"
else
  echo -e "${YELLOW}✗ リモートバックエンド設定が無効です${NC}"
fi

# terraformの状態ファイルがS3に存在するか確認
if aws s3api head-object --bucket "${STATE_BUCKET}" --key "${ENVIRONMENT}/terraform.tfstate" &>/dev/null; then
  echo -e "${GREEN}✓ Terraform状態ファイルがS3に存在します${NC}"
else
  echo -e "${YELLOW}✗ Terraform状態ファイルがS3に存在しません${NC}"
fi

# デプロイされたリソースの確認
echo -e "${BLUE}${ENVIRONMENT}環境にデプロイされたAWSリソースを確認しています...${NC}"

# VPCの確認
VPC_COUNT=$(aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=${ENVIRONMENT}" --query 'Vpcs[].VpcId' --output text | wc -w)
if [ "${VPC_COUNT}" -gt 0 ]; then
  echo -e "${GREEN}✓ ${VPC_COUNT}個のVPCがデプロイされています${NC}"
  # VPC詳細情報の取得
  VPC_IDS=$(aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=${ENVIRONMENT}" --query 'Vpcs[].VpcId' --output text)
  for VPC_ID in $VPC_IDS; do
    VPC_NAME=$(aws ec2 describe-vpcs --vpc-ids ${VPC_ID} --query 'Vpcs[0].Tags[?Key==`Name`].Value' --output text)
    echo "  - VPC: ${VPC_ID} (${VPC_NAME})"
    
    # サブネット情報
    SUBNET_COUNT=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=${VPC_ID}" --query 'Subnets[].SubnetId' --output text | wc -w)
    if [ "${SUBNET_COUNT}" -gt 0 ]; then
      echo "    サブネット: ${SUBNET_COUNT}個"
    fi
  done
else
  echo -e "${YELLOW}✗ デプロイされたVPCが見つかりません${NC}"
fi

# RDSインスタンスの確認
RDS_COUNT=$(aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='${ENV_PREFIX}-postgres'].DBInstanceIdentifier" --output text | wc -w)
if [ "${RDS_COUNT}" -gt 0 ]; then
  echo -e "${GREEN}✓ RDSインスタンス (${ENV_PREFIX}-postgres) は存在します${NC}"
  # RDS詳細情報の取得
  RDS_STATUS=$(aws rds describe-db-instances --db-instance-identifier ${ENV_PREFIX}-postgres --query 'DBInstances[0].DBInstanceStatus' --output text)
  RDS_ENGINE=$(aws rds describe-db-instances --db-instance-identifier ${ENV_PREFIX}-postgres --query 'DBInstances[0].Engine' --output text)
  RDS_VERSION=$(aws rds describe-db-instances --db-instance-identifier ${ENV_PREFIX}-postgres --query 'DBInstances[0].EngineVersion' --output text)
  echo "  - ステータス: ${RDS_STATUS}"
  echo "  - エンジン: ${RDS_ENGINE} ${RDS_VERSION}"
else
  echo -e "${YELLOW}✗ RDSインスタンス (${ENV_PREFIX}-postgres) は存在しません${NC}"
fi

# 共有ECSクラスターの確認（共通ライブラリを使用）
SHARED_CLUSTER="${ENV_PREFIX}-shared-cluster"
if ecs_cluster_exists "$SHARED_CLUSTER"; then
  echo -e "${GREEN}✓ 共有ECSクラスター (${SHARED_CLUSTER}) は存在します${NC}"
  
  # 共有クラスター内のサービス数
  TOTAL_SERVICE_COUNT=0
  
  # 各サービスタイプごとのECSサービス確認
  for SERVICE_TYPE in "${SERVICE_TYPES[@]}"; do
    SERVICE_NAME="${ENV_PREFIX}-${SERVICE_TYPE}"
    SERVICE_EXISTS=$(aws ecs describe-services --cluster ${SHARED_CLUSTER} --services ${SERVICE_NAME} --query "services[?status!='INACTIVE'].serviceName" --output text 2>/dev/null || echo "")
    
    if [ -n "${SERVICE_EXISTS}" ]; then
      echo -e "  ${GREEN}✓ ${SERVICE_TYPE} サービス (${SERVICE_NAME}) は存在します${NC}"
      
      # サービスの詳細情報
      SERVICE_STATUS=$(aws ecs describe-services --cluster ${SHARED_CLUSTER} --services ${SERVICE_NAME} --query "services[0].status" --output text)
      RUNNING_COUNT=$(aws ecs describe-services --cluster ${SHARED_CLUSTER} --services ${SERVICE_NAME} --query "services[0].runningCount" --output text)
      DESIRED_COUNT=$(aws ecs describe-services --cluster ${SHARED_CLUSTER} --services ${SERVICE_NAME} --query "services[0].desiredCount" --output text)
      
      echo "    - ステータス: ${SERVICE_STATUS}"
      echo "    - 実行中タスク: ${RUNNING_COUNT}/${DESIRED_COUNT}"
      
      TOTAL_SERVICE_COUNT=$((TOTAL_SERVICE_COUNT + 1))
    else
      echo -e "  ${YELLOW}✗ ${SERVICE_TYPE} サービス (${SERVICE_NAME}) は存在しません${NC}"
    fi
  done
  
  echo "  合計サービス数: ${TOTAL_SERVICE_COUNT}"
else
  echo -e "${YELLOW}✗ 共有ECSクラスター (${SHARED_CLUSTER}) は存在しません${NC}"
fi

# ターゲットグループの確認（独立したセクション）
echo -e "${BLUE}ターゲットグループの状態:${NC}"
TG_ARNS=$(aws elbv2 describe-target-groups --query "TargetGroups[?contains(TargetGroupName, '${ENV_PREFIX}')].TargetGroupArn" --output text)

if [ -n "${TG_ARNS}" ]; then
  TG_COUNT=$(echo ${TG_ARNS} | wc -w)
  echo -e "${GREEN}✓ ${TG_COUNT}個のターゲットグループが見つかりました${NC}"
  
  # 各ターゲットグループの情報
  for TG_ARN in ${TG_ARNS}; do
    TG_NAME=$(aws elbv2 describe-target-groups --target-group-arns ${TG_ARN} --query "TargetGroups[0].TargetGroupName" --output text)
    TG_PORT=$(aws elbv2 describe-target-groups --target-group-arns ${TG_ARN} --query "TargetGroups[0].Port" --output text)
    TG_PROTOCOL=$(aws elbv2 describe-target-groups --target-group-arns ${TG_ARN} --query "TargetGroups[0].Protocol" --output text)
    TG_HEALTH_PATH=$(aws elbv2 describe-target-groups --target-group-arns ${TG_ARN} --query "TargetGroups[0].HealthCheckPath" --output text)
    
    echo "  - ${TG_NAME} (${TG_PROTOCOL}:${TG_PORT}, ヘルスパス: ${TG_HEALTH_PATH})"
    
    # 関連LBの確認
    TG_LBS=$(aws elbv2 describe-target-groups --target-group-arns ${TG_ARN} --query "TargetGroups[0].LoadBalancerArns" --output text)
    if [ -n "${TG_LBS}" ] && [ "${TG_LBS}" != "None" ]; then
      LB_COUNT=$(echo ${TG_LBS} | wc -w)
      echo "    関連LB: ${LB_COUNT}個"
    else
      echo "    関連LB: なし (孤立ターゲットグループ)"
    fi
  done
else
  echo -e "${YELLOW}✗ ターゲットグループが見つかりません${NC}"
fi

# 各サービスのロードバランサー確認
echo -e "${BLUE}各サービスのロードバランサー状態:${NC}"

for SERVICE_TYPE in "${SERVICE_TYPES[@]}"; do
  # サービス固有のALB名パターン
  ALB_PATTERN="${ENV_PREFIX}-${SERVICE_TYPE}"
  
  # ALBの存在確認
  ALB_EXISTS=$(aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName, '${ALB_PATTERN}')].LoadBalancerName" --output text)
  
  if [ -n "${ALB_EXISTS}" ]; then
    echo -e "${GREEN}✓ ${SERVICE_TYPE} 用ロードバランサー (${ALB_EXISTS}) は存在します${NC}"
    
    # ALBのステータス
    ALB_STATE=$(aws elbv2 describe-load-balancers --names ${ALB_EXISTS} --query "LoadBalancers[0].State.Code" --output text)
    echo "  - ステータス: ${ALB_STATE}"
    
    # ターゲットグループの確認
    TG_PATTERN="${ENV_PREFIX}-${SERVICE_TYPE}"
    TG_ARNS=$(aws elbv2 describe-target-groups --query "TargetGroups[?contains(TargetGroupName, '${TG_PATTERN}')].TargetGroupArn" --output text)
    
    if [ -n "${TG_ARNS}" ]; then
      TG_COUNT=$(echo ${TG_ARNS} | wc -w)
      echo "  - ターゲットグループ: ${TG_COUNT}個"
      
      # ターゲットの状態（オプション）
      for TG_ARN in ${TG_ARNS}; do
        TG_NAME=$(aws elbv2 describe-target-groups --target-group-arns ${TG_ARN} --query "TargetGroups[0].TargetGroupName" --output text)
        TG_HEALTH=$(aws elbv2 describe-target-health --target-group-arn ${TG_ARN} --query "TargetHealthDescriptions[0].TargetHealth.State" --output text 2>/dev/null || echo "unknown")
        echo "    * ${TG_NAME}: ${TG_HEALTH}"
      done
    else
      echo "  - ターゲットグループが見つかりません"
    fi
  else
    echo -e "${YELLOW}✗ ${SERVICE_TYPE} 用ロードバランサーは存在しません${NC}"
  fi
done

# セキュリティグループの確認
echo -e "${BLUE}セキュリティグループの状態:${NC}"

# 共有セキュリティグループ
SHARED_SG_PATTERN="${ENV_PREFIX}-shared"
SHARED_SG_COUNT=$(aws ec2 describe-security-groups --filters "Name=group-name,Values=*${SHARED_SG_PATTERN}*" --query 'SecurityGroups[].GroupId' --output text | wc -w)

if [ "${SHARED_SG_COUNT}" -gt 0 ]; then
  echo -e "${GREEN}✓ ${SHARED_SG_COUNT}個の共有セキュリティグループが見つかりました${NC}"
  
  # 詳細情報（オプション）
  SHARED_SG_IDS=$(aws ec2 describe-security-groups --filters "Name=group-name,Values=*${SHARED_SG_PATTERN}*" --query 'SecurityGroups[].GroupId' --output text)
  for SG_ID in ${SHARED_SG_IDS}; do
    SG_NAME=$(aws ec2 describe-security-groups --group-ids ${SG_ID} --query 'SecurityGroups[0].GroupName' --output text)
    echo "  - ${SG_ID} (${SG_NAME})"
  done
else
  echo -e "${YELLOW}✗ 共有セキュリティグループが見つかりません${NC}"
fi

# サービス固有のセキュリティグループ
TOTAL_SERVICE_SG_COUNT=0

for SERVICE_TYPE in "${SERVICE_TYPES[@]}"; do
  SERVICE_SG_PATTERN="${ENV_PREFIX}-${SERVICE_TYPE}"
  SERVICE_SG_COUNT=$(aws ec2 describe-security-groups --filters "Name=group-name,Values=*${SERVICE_SG_PATTERN}*" --query 'SecurityGroups[].GroupId' --output text | wc -w)
  
  if [ "${SERVICE_SG_COUNT}" -gt 0 ]; then
    echo -e "${GREEN}✓ ${SERVICE_SG_COUNT}個の ${SERVICE_TYPE} 用セキュリティグループが見つかりました${NC}"
    
    # 詳細情報（オプション）
    SERVICE_SG_IDS=$(aws ec2 describe-security-groups --filters "Name=group-name,Values=*${SERVICE_SG_PATTERN}*" --query 'SecurityGroups[].GroupId' --output text)
    for SG_ID in ${SERVICE_SG_IDS}; do
      SG_NAME=$(aws ec2 describe-security-groups --group-ids ${SG_ID} --query 'SecurityGroups[0].GroupName' --output text)
      echo "  - ${SG_ID} (${SG_NAME})"
    done
    
    TOTAL_SERVICE_SG_COUNT=$((TOTAL_SERVICE_SG_COUNT + SERVICE_SG_COUNT))
  else
    echo -e "${YELLOW}✗ ${SERVICE_TYPE} 用セキュリティグループが見つかりません${NC}"
  fi
done

echo "合計サービス固有セキュリティグループ: ${TOTAL_SERVICE_SG_COUNT}個"

# ECRリポジトリの確認
echo -e "${BLUE}ECRリポジトリの状態:${NC}"

for SERVICE_TYPE in "${SERVICE_TYPES[@]}"; do
  REPO_NAME="${ENV_PREFIX}-test-management-${SERVICE_TYPE}"
  
  if aws ecr describe-repositories --repository-names ${REPO_NAME} --region ${AWS_REGION} &>/dev/null; then
    echo -e "${GREEN}✓ ${SERVICE_TYPE} 用ECRリポジトリ (${REPO_NAME}) は存在します${NC}"
    
    # 最新イメージ情報
    LATEST_IMAGE=$(aws ecr describe-images --repository-name ${REPO_NAME} --query 'sort_by(imageDetails,& imagePushedAt)[-1].imageTags[0]' --output text --region ${AWS_REGION} 2>/dev/null || echo "タグなし")
    LATEST_DIGEST=$(aws ecr describe-images --repository-name ${REPO_NAME} --query 'sort_by(imageDetails,& imagePushedAt)[-1].imageDigest' --output text --region ${AWS_REGION} 2>/dev/null || echo "不明")
    PUSHED_DATE=$(aws ecr describe-images --repository-name ${REPO_NAME} --query 'sort_by(imageDetails,& imagePushedAt)[-1].imagePushedAt' --output text --region ${AWS_REGION} 2>/dev/null || echo "不明")
    
    if [ "${LATEST_IMAGE}" != "None" ] && [ "${LATEST_IMAGE}" != "タグなし" ]; then
      echo "  - 最新イメージ: ${LATEST_IMAGE}"
      echo "  - ダイジェスト: ${LATEST_DIGEST:0:12}..."
      echo "  - プッシュ日時: ${PUSHED_DATE}"
    else
      echo "  - イメージがありません"
    fi
  else
    echo -e "${YELLOW}✗ ${SERVICE_TYPE} 用ECRリポジトリ (${REPO_NAME}) は存在しません${NC}"
  fi
done

echo -e "${BLUE}${ENVIRONMENT}環境のインフラ状態確認が完了しました${NC}"

# 使用方法の表示
if [ "$1" = "" ]; then
  echo
  echo "別の環境を確認するには:"
  echo "  $0 production"
fi