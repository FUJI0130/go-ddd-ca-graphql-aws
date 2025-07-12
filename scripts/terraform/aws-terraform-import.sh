#!/bin/bash
# aws-terraform-import.sh - AWS環境のリソースをTerraform状態にインポートするスクリプト

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 引数の解析
ENV=${1:-development}
FORCE=${2:-no}  # yes/no - 強制モード

echo -e "${BLUE}AWS環境のリソースをTerraform状態にインポートします（環境: $ENV）...${NC}"

# 現在のディレクトリを保存
ORIGINAL_DIR=$(pwd)

# 環境ディレクトリに移動
cd deployments/terraform/environments/$ENV

# VPCの検出とインポート
echo -e "${BLUE}VPCを検出してインポートしています...${NC}"
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=$ENV" --query "Vpcs[0].VpcId" --output text)

if [ "$VPC_ID" != "None" ] && [ ! -z "$VPC_ID" ]; then
  echo -e "${GREEN}VPC検出: $VPC_ID${NC}"
  
  # すでにモジュールに存在するか確認
  if terraform state list module.networking.aws_vpc.main &>/dev/null; then
    echo -e "${YELLOW}VPCはすでにTerraform状態に存在します${NC}"
  else
    echo -e "${BLUE}VPCをインポートしています: $VPC_ID${NC}"
    terraform import module.networking.aws_vpc.main $VPC_ID || echo -e "${YELLOW}インポートに失敗しました${NC}"
  fi
  
  # サブネットのインポート
  echo -e "${BLUE}サブネットを検出してインポートしています...${NC}"
  SUBNET_IDS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[*].SubnetId" --output text)
  
  if [ ! -z "$SUBNET_IDS" ]; then
    # サブネットのタイプを判別してインポート
    for SUBNET_ID in $SUBNET_IDS; do
      SUBNET_NAME=$(aws ec2 describe-subnets --subnet-ids $SUBNET_ID --query "Subnets[0].Tags[?Key=='Name'].Value" --output text)
      
      if [[ "$SUBNET_NAME" == *"public"* ]]; then
        # パブリックサブネット
        ZONE=$(aws ec2 describe-subnets --subnet-ids $SUBNET_ID --query "Subnets[0].AvailabilityZone" --output text)
        ZONE_SUFFIX=${ZONE: -1}
        
        if ! terraform state list module.networking.aws_subnet.public[$ZONE_SUFFIX] &>/dev/null; then
          echo -e "${BLUE}パブリックサブネットをインポートしています: $SUBNET_ID (Zone: $ZONE)${NC}"
          terraform import module.networking.aws_subnet.public[$ZONE_SUFFIX] $SUBNET_ID || echo -e "${YELLOW}インポートに失敗しました${NC}"
        fi
      elif [[ "$SUBNET_NAME" == *"private"* ]]; then
        # プライベートサブネット
        ZONE=$(aws ec2 describe-subnets --subnet-ids $SUBNET_ID --query "Subnets[0].AvailabilityZone" --output text)
        ZONE_SUFFIX=${ZONE: -1}
        
        if ! terraform state list module.networking.aws_subnet.private[$ZONE_SUFFIX] &>/dev/null; then
          echo -e "${BLUE}プライベートサブネットをインポートしています: $SUBNET_ID (Zone: $ZONE)${NC}"
          terraform import module.networking.aws_subnet.private[$ZONE_SUFFIX] $SUBNET_ID || echo -e "${YELLOW}インポートに失敗しました${NC}"
        fi
      else
        echo -e "${YELLOW}タイプが不明なサブネット: $SUBNET_ID ($SUBNET_NAME)${NC}"
      fi
    done
  fi
else
  echo -e "${YELLOW}VPCが見つかりません${NC}"
fi

# RDSインスタンスのインポート
echo -e "${BLUE}RDSインスタンスを検出してインポートしています...${NC}"
RDS_ID="${ENV}-postgres"
if aws rds describe-db-instances --db-instance-identifier $RDS_ID &>/dev/null; then
  echo -e "${GREEN}RDS検出: $RDS_ID${NC}"
  
  if ! terraform state list module.database.aws_db_instance.postgres &>/dev/null; then
    echo -e "${BLUE}RDSインスタンスをインポートしています: $RDS_ID${NC}"
    terraform import module.database.aws_db_instance.postgres $RDS_ID || echo -e "${YELLOW}インポートに失敗しました${NC}"
  else
    echo -e "${YELLOW}RDSはすでにTerraform状態に存在します${NC}"
  fi
else
  echo -e "${YELLOW}RDSインスタンスが見つかりません${NC}"
fi

# ECSクラスターのインポート
echo -e "${BLUE}ECSクラスターを検出してインポートしています...${NC}"
CLUSTER_NAME="${ENV}-shared-cluster"
if aws ecs describe-clusters --clusters $CLUSTER_NAME --query "clusters[?clusterName=='$CLUSTER_NAME']" --output text &>/dev/null; then
  echo -e "${GREEN}ECSクラスター検出: $CLUSTER_NAME${NC}"
  
  if ! terraform state list module.shared_ecs_cluster.aws_ecs_cluster.main &>/dev/null; then
    echo -e "${BLUE}ECSクラスターをインポートしています: $CLUSTER_NAME${NC}"
    terraform import module.shared_ecs_cluster.aws_ecs_cluster.main $CLUSTER_NAME || echo -e "${YELLOW}インポートに失敗しました${NC}"
  else
    echo -e "${YELLOW}ECSクラスターはすでにTerraform状態に存在します${NC}"
  fi
  
  # ECSサービスのインポート
  echo -e "${BLUE}ECSサービスを検出してインポートしています...${NC}"
  for SERVICE_TYPE in api graphql grpc; do
    SERVICE_NAME="${ENV}-${SERVICE_TYPE}"
    
    if aws ecs describe-services --cluster $CLUSTER_NAME --services $SERVICE_NAME --query "services[?serviceName=='$SERVICE_NAME']" --output text &>/dev/null; then
      echo -e "${GREEN}ECSサービス検出: $SERVICE_NAME${NC}"
      
      if ! terraform state list module.service_${SERVICE_TYPE}.aws_ecs_service.app &>/dev/null; then
        echo -e "${BLUE}ECSサービスをインポートしています: $SERVICE_NAME${NC}"
        terraform import module.service_${SERVICE_TYPE}.aws_ecs_service.app $CLUSTER_NAME/$SERVICE_NAME || echo -e "${YELLOW}インポートに失敗しました${NC}"
      else
        echo -e "${YELLOW}ECSサービス($SERVICE_TYPE)はすでにTerraform状態に存在します${NC}"
      fi
    else
      echo -e "${YELLOW}ECSサービス($SERVICE_TYPE)が見つかりません${NC}"
    fi
  done
else
  echo -e "${YELLOW}ECSクラスターが見つかりません${NC}"
fi

# ALBのインポート
echo -e "${BLUE}ロードバランサーを検出してインポートしています...${NC}"
for SERVICE_TYPE in api graphql grpc; do
  ALB_NAME="${ENV}-${SERVICE_TYPE}-alb"
  
  ALB_ARN=$(aws elbv2 describe-load-balancers --query "LoadBalancers[?LoadBalancerName=='$ALB_NAME'].LoadBalancerArn" --output text)
  
  if [ ! -z "$ALB_ARN" ] && [ "$ALB_ARN" != "None" ]; then
    echo -e "${GREEN}ALB検出: $ALB_NAME${NC}"
    
    if ! terraform state list module.loadbalancer_${SERVICE_TYPE}.aws_lb.main &>/dev/null; then
      echo -e "${BLUE}ALBをインポートしています: $ALB_NAME${NC}"
      terraform import module.loadbalancer_${SERVICE_TYPE}.aws_lb.main $ALB_ARN || echo -e "${YELLOW}インポートに失敗しました${NC}"
    else
      echo -e "${YELLOW}ALB($SERVICE_TYPE)はすでにTerraform状態に存在します${NC}"
    fi
    
    # ターゲットグループのインポート
    TG_NAME="${ENV}-${SERVICE_TYPE}-tg"
    TG_ARN=$(aws elbv2 describe-target-groups --query "TargetGroups[?TargetGroupName=='$TG_NAME'].TargetGroupArn" --output text)
    
    if [ ! -z "$TG_ARN" ] && [ "$TG_ARN" != "None" ]; then
      echo -e "${GREEN}ターゲットグループ検出: $TG_NAME${NC}"
      
      if ! terraform state list module.target_group_${SERVICE_TYPE}.aws_lb_target_group.app &>/dev/null; then
        echo -e "${BLUE}ターゲットグループをインポートしています: $TG_NAME${NC}"
        terraform import module.target_group_${SERVICE_TYPE}.aws_lb_target_group.app $TG_ARN || echo -e "${YELLOW}インポートに失敗しました${NC}"
      else
        echo -e "${YELLOW}ターゲットグループ($SERVICE_TYPE)はすでにTerraform状態に存在します${NC}"
      fi
    else
      echo -e "${YELLOW}ターゲットグループ($SERVICE_TYPE)が見つかりません${NC}"
    fi
  else
    echo -e "${YELLOW}ALB($SERVICE_TYPE)が見つかりません${NC}"
  fi
done

# 元のディレクトリに戻る
cd $ORIGINAL_DIR

echo -e "${GREEN}AWS環境リソースのインポートが完了しました${NC}"
echo -e "${BLUE}Terraform状態の整合性を確認するには次のコマンドを実行してください:${NC}"
echo -e "make terraform-verify TF_ENV=$ENV"