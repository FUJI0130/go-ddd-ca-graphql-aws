#!/bin/bash
# tag-cleanup.sh - タグベースでAWSリソースを削除するスクリプト
# 共通ユーティリティのインポート
source $(dirname "$0")/common/aws_resource_utils.sh

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 引数の解析
ENV=${1:-development}
AUTO_MODE=${2:-""}
ENV_PREFIX=${ENV}  # プレフィックスの設定

# 自動モードでない場合は確認を求める
if [ "$AUTO_MODE" != "auto" ]; then
  echo -e "${RED}警告: このスクリプトはタグベースでAWSリソースを削除します。この操作は元に戻せません！${NC}"
  read -p "続行しますか？(y/n) " CONTINUE

  if [ "$CONTINUE" != "y" ]; then
    echo "中止します"
    exit 1
  fi
fi

echo -e "${BLUE}タグベースのリソース削除を開始します（環境: $ENV）...${NC}"

# Elastic IPの解放
echo -e "\n${BLUE}Elastic IPを確認しています...${NC}"
ELASTIC_IPS=$(aws ec2 describe-addresses --query "Addresses[?contains(Tags[?Key=='Environment'].Value, '$ENV')].AllocationId" --output text)
if [ ! -z "$ELASTIC_IPS" ]; then
  for EIP_ID in $ELASTIC_IPS; do
    echo -e "${YELLOW}Elastic IPを解放: $EIP_ID${NC}"
    aws ec2 release-address --allocation-id $EIP_ID
  done
else
  echo -e "${GREEN}解放対象のElastic IPはありません${NC}"
fi

# ECSサービスの削除
echo -e "\n${BLUE}ECSサービスを確認しています...${NC}"
CLUSTER_NAME="$ENV-shared-cluster"

# 環境変数の確認とデバッグ出力
if [ "$AWS_RESOURCE_DEBUG" = "true" ]; then
  echo -e "${BLUE}[DEBUG] ENV=$ENV, CLUSTER_NAME=$CLUSTER_NAME${NC}"
fi

# 共通ライブラリの関数を使用
if ecs_delete_cluster "$CLUSTER_NAME" true; then
  echo -e "${GREEN}ECSクラスターの削除が完了しました${NC}"
else
  echo -e "${RED}ECSクラスターの削除中に問題が発生しました。手動での確認をお勧めします${NC}"
  echo -e "${YELLOW}手動削除コマンド: aws ecs delete-cluster --cluster \"$CLUSTER_NAME\"${NC}"
fi

# ロードバランサーの削除
echo -e "\n${BLUE}ロードバランサーを確認しています...${NC}"
LB_ARNS=$(aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName, '$ENV')].LoadBalancerArn" --output text)
if [ ! -z "$LB_ARNS" ]; then
  for LB_ARN in $LB_ARNS; do
    # リスナーの削除（依存関係のため先に削除）
    LISTENERS=$(aws elbv2 describe-listeners --load-balancer-arn $LB_ARN --query "Listeners[*].ListenerArn" --output text)
    if [ ! -z "$LISTENERS" ]; then
      for LISTENER in $LISTENERS; do
        echo -e "${YELLOW}リスナーを削除: $LISTENER${NC}"
        aws elbv2 delete-listener --listener-arn $LISTENER
      done
    fi
    
    LB_NAME=$(aws elbv2 describe-load-balancers --load-balancer-arns $LB_ARN --query "LoadBalancers[0].LoadBalancerName" --output text)
    echo -e "${YELLOW}ロードバランサーを削除: $LB_NAME${NC}"
    aws elbv2 delete-load-balancer --load-balancer-arn $LB_ARN
  done
else
  echo -e "${GREEN}削除対象のロードバランサーはありません${NC}"
fi

# ターゲットグループの削除（ロードバランサーの有無に関わらず実行）
echo -e "\n${BLUE}ターゲットグループを確認しています...${NC}"
TG_ARNS=$(aws elbv2 describe-target-groups --query "TargetGroups[?contains(TargetGroupName, '$ENV')].TargetGroupArn" --output text)
if [ ! -z "$TG_ARNS" ]; then
  for TG_ARN in $TG_ARNS; do
    TG_NAME=$(aws elbv2 describe-target-groups --target-group-arns $TG_ARN --query "TargetGroups[0].TargetGroupName" --output text)
    echo -e "${YELLOW}ターゲットグループを削除: $TG_NAME${NC}"
    aws elbv2 delete-target-group --target-group-arn $TG_ARN
  done
else
  echo -e "${GREEN}削除対象のターゲットグループはありません${NC}"
fi

# セキュリティグループの削除
echo -e "\n${BLUE}セキュリティグループを確認しています...${NC}"
# エラーハンドリングを追加
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=$ENV" --query "Vpcs[0].VpcId" --output text 2>/dev/null || echo "")
SG_IDS=$(aws ec2 describe-security-groups --filters "Name=vpc-id,Values=$VPC_ID" --query "SecurityGroups[*].GroupId" --output text 2>/dev/null || echo "")

if [ -z "$SG_IDS" ] || [ "$SG_IDS" = "None" ]; then
  echo -e "${GREEN}削除対象のセキュリティグループはありません${NC}"
else
  for SG_ID in $SG_IDS; do
    # デフォルトSGはスキップ
    SG_NAME=$(aws ec2 describe-security-groups --group-ids $SG_ID --query "SecurityGroups[0].GroupName" --output text 2>/dev/null || echo "unknown")
    if [[ "$SG_NAME" == "default" ]]; then
      echo -e "${YELLOW}デフォルトセキュリティグループはスキップします: $SG_ID${NC}"
      continue
    fi
    
    echo -e "${YELLOW}セキュリティグループを削除: $SG_ID ($SG_NAME)${NC}"
    aws ec2 delete-security-group --group-id $SG_ID 2>/dev/null || echo -e "${RED}削除できませんでした: $SG_ID - 依存リソースが存在する可能性があります${NC}"
  done
fi

# NAT Gatewayの削除
echo -e "\n${BLUE}NAT Gatewayを確認しています...${NC}"
NAT_IDS=$(aws ec2 describe-nat-gateways --filter "Name=state,Values=available" --query "NatGateways[?contains(Tags[?Key=='Name'].Value, '$ENV')].NatGatewayId" --output text)
if [ ! -z "$NAT_IDS" ]; then
 for NAT_ID in $NAT_IDS; do
   echo -e "${YELLOW}NAT Gatewayを削除: $NAT_ID${NC}"
   aws ec2 delete-nat-gateway --nat-gateway-id $NAT_ID
 done
 echo -e "${YELLOW}NAT Gatewayの削除を待機しています（60秒）...${NC}"
 sleep 60
else
 echo -e "${GREEN}削除対象のNAT Gatewayはありません${NC}"
fi

# インターネットゲートウェイのデタッチと削除
echo -e "\n${BLUE}インターネットゲートウェイを確認しています...${NC}"
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=$ENV" --query "Vpcs[0].VpcId" --output text)
if [ "$VPC_ID" != "None" ] && [ ! -z "$VPC_ID" ]; then
 IGW_ID=$(aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=$VPC_ID" --query "InternetGateways[0].InternetGatewayId" --output text)
 if [ "$IGW_ID" != "None" ] && [ ! -z "$IGW_ID" ]; then
   echo -e "${YELLOW}インターネットゲートウェイをデタッチ: $IGW_ID${NC}"
   aws ec2 detach-internet-gateway --internet-gateway-id $IGW_ID --vpc-id $VPC_ID
   echo -e "${YELLOW}インターネットゲートウェイを削除: $IGW_ID${NC}"
   aws ec2 delete-internet-gateway --internet-gateway-id $IGW_ID
 else
   echo -e "${GREEN}削除対象のインターネットゲートウェイはありません${NC}"
 fi
else
 echo -e "${GREEN}VPCが見つかりません${NC}"
fi

# サブネットの削除
echo -e "\n${BLUE}サブネットを確認しています...${NC}"
if [ "$VPC_ID" != "None" ] && [ ! -z "$VPC_ID" ]; then
 SUBNET_IDS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[*].SubnetId" --output text)
 if [ ! -z "$SUBNET_IDS" ]; then
   for SUBNET_ID in $SUBNET_IDS; do
     echo -e "${YELLOW}サブネットを削除: $SUBNET_ID${NC}"
     aws ec2 delete-subnet --subnet-id $SUBNET_ID || echo -e "${RED}削除できませんでした - 依存リソースが存在する可能性があります${NC}"
   done
 else
   echo -e "${GREEN}削除対象のサブネットはありません${NC}"
 fi
fi

# ルートテーブルの削除
echo -e "\n${BLUE}ルートテーブルを確認しています...${NC}"
if [ "$VPC_ID" != "None" ] && [ ! -z "$VPC_ID" ]; then
 # メインでないルートテーブルを取得
 RT_IDS=$(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=$VPC_ID" --query "RouteTables[?Associations[0].Main!=\`true\`].RouteTableId" --output text)
 if [ ! -z "$RT_IDS" ]; then
   for RT_ID in $RT_IDS; do
     # 関連付けを解除
     ASSOC_IDS=$(aws ec2 describe-route-tables --route-table-ids $RT_ID --query "RouteTables[0].Associations[?SubnetId!=null].RouteTableAssociationId" --output text)
     for ASSOC_ID in $ASSOC_IDS; do
       echo -e "${YELLOW}ルートテーブル関連付けを解除: $ASSOC_ID${NC}"
       aws ec2 disassociate-route-table --association-id $ASSOC_ID
     done
     
     echo -e "${YELLOW}ルートテーブルを削除: $RT_ID${NC}"
     aws ec2 delete-route-table --route-table-id $RT_ID || echo -e "${RED}削除できませんでした - 依存リソースが存在する可能性があります${NC}"
   done
 else
   echo -e "${GREEN}削除対象のルートテーブルはありません${NC}"
 fi
fi

# VPCの削除
echo -e "\n${BLUE}VPCを確認しています...${NC}"
if [ "$VPC_ID" != "None" ] && [ ! -z "$VPC_ID" ]; then
 echo -e "${YELLOW}VPCを削除: $VPC_ID${NC}"
 aws ec2 delete-vpc --vpc-id $VPC_ID || echo -e "${RED}VPC削除に失敗しました - 依存リソースが存在する可能性があります${NC}"
else
 echo -e "${GREEN}削除対象のVPCはありません${NC}"
fi

# 確認と報告
echo -e "\n${BLUE}クリーンアップ後の状態を確認しています...${NC}"
make -C $(pwd) check-resources TF_ENV=$ENV
make -C $(pwd) cost-estimate TF_ENV=$ENV

echo -e "\n${GREEN}タグベースのクリーンアップが完了しました${NC}"
echo -e "${YELLOW}注意: Terraform状態は更新されていません。terraform state listで確認し、必要に応じてterraform state rmを実行してください。${NC}"