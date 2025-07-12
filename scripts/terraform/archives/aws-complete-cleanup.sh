#!/bin/bash
# aws-complete-cleanup.sh - すべてのAWSリソースを削除

export AWS_REGION=ap-northeast-1
export ENV_PREFIX=development

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}AWS環境のすべてのリソースを削除しています...${NC}"

# 1. すべてのECSサービスを停止
echo -e "${BLUE}ECSサービスを停止しています...${NC}"
CLUSTER=${ENV_PREFIX}-shared-cluster
SERVICES=$(aws ecs list-services --cluster ${CLUSTER} --query "serviceArns" --output text)

for SERVICE in ${SERVICES}; do
  SERVICE_NAME=$(echo ${SERVICE} | cut -d '/' -f 3)
  echo -e "${YELLOW}サービスを停止: ${SERVICE_NAME}${NC}"
  aws ecs update-service --cluster ${CLUSTER} --service ${SERVICE_NAME} --desired-count 0
  aws ecs delete-service --cluster ${CLUSTER} --service ${SERVICE_NAME} --force
done

# 2. すべてのロードバランサーとリスナーを削除
echo -e "${BLUE}ロードバランサーとリスナーを削除しています...${NC}"
LBS=$(aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName,'${ENV_PREFIX}')].LoadBalancerArn" --output text)

for LB in ${LBS}; do
  LB_NAME=$(aws elbv2 describe-load-balancers --load-balancer-arns ${LB} --query "LoadBalancers[0].LoadBalancerName" --output text)
  echo -e "${YELLOW}ロードバランサーを削除: ${LB_NAME}${NC}"
  
  # リスナーを先に削除
  LISTENERS=$(aws elbv2 describe-listeners --load-balancer-arn ${LB} --query "Listeners[*].ListenerArn" --output text)
  for LISTENER in ${LISTENERS}; do
    aws elbv2 delete-listener --listener-arn ${LISTENER}
  done
  
  # ロードバランサーを削除
  aws elbv2 delete-load-balancer --load-balancer-arn ${LB}
done

# 3. すべてのターゲットグループを削除
echo -e "${BLUE}ターゲットグループを削除しています...${NC}"
TGS=$(aws elbv2 describe-target-groups --query "TargetGroups[?contains(TargetGroupName,'${ENV_PREFIX}')].TargetGroupArn" --output text)

for TG in ${TGS}; do
  TG_NAME=$(aws elbv2 describe-target-groups --target-group-arns ${TG} --query "TargetGroups[0].TargetGroupName" --output text)
  echo -e "${YELLOW}ターゲットグループを削除: ${TG_NAME}${NC}"
  aws elbv2 delete-target-group --target-group-arn ${TG}
done

# 4. すべてのセキュリティグループを削除
echo -e "${BLUE}セキュリティグループを削除しています...${NC}"
SGs=$(aws ec2 describe-security-groups --filters "Name=tag:Environment,Values=${ENV_PREFIX}" --query "SecurityGroups[*].GroupId" --output text)

for SG in ${SGs}; do
  SG_NAME=$(aws ec2 describe-security-groups --group-ids ${SG} --query "SecurityGroups[0].GroupName" --output text)
  echo -e "${YELLOW}セキュリティグループを削除: ${SG_NAME}${NC}"
  aws ec2 delete-security-group --group-id ${SG} || echo -e "${RED}セキュリティグループの削除に失敗: ${SG_NAME} - 依存関係があるかもしれません${NC}"
done

# 5. RDSインスタンスを削除
echo -e "${BLUE}RDSインスタンスを削除しています...${NC}"
RDS_INSTANCES=$(aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='${ENV_PREFIX}-postgres'].DBInstanceIdentifier" --output text)

for RDS in ${RDS_INSTANCES}; do
  echo -e "${YELLOW}RDSインスタンスを削除: ${RDS}${NC}"
  aws rds delete-db-instance --db-instance-identifier ${RDS} --skip-final-snapshot
  
  echo -e "${BLUE}RDSインスタンスの削除が完了するまで待機しています...${NC}"
  aws rds wait db-instance-deleted --db-instance-identifier ${RDS}
done

# 6. すべてのECSクラスターを削除
echo -e "${BLUE}ECSクラスターを削除しています...${NC}"
CLUSTERS=$(aws ecs list-clusters --query "clusterArns[?contains(@,'${ENV_PREFIX}')]" --output text)

for CLUSTER in ${CLUSTERS}; do
  CLUSTER_NAME=$(echo ${CLUSTER} | cut -d '/' -f 2)
  echo -e "${YELLOW}クラスターを削除: ${CLUSTER_NAME}${NC}"
  aws ecs delete-cluster --cluster ${CLUSTER_NAME}
done

# 7. VPCを削除（依存リソースがすべて削除された後）
echo -e "${BLUE}VPC関連リソースを削除しています...${NC}"
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=${ENV_PREFIX}" --query "Vpcs[0].VpcId" --output text)

if [ ! -z "${VPC_ID}" ] && [ "${VPC_ID}" != "None" ]; then
  echo -e "${YELLOW}VPC ID: ${VPC_ID}${NC}"
  
  # サブネットを削除
  SUBNETS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=${VPC_ID}" --query "Subnets[*].SubnetId" --output text)
  for SUBNET in ${SUBNETS}; do
    echo -e "${YELLOW}サブネットを削除: ${SUBNET}${NC}"
    aws ec2 delete-subnet --subnet-id ${SUBNET}
  done
  
  # ルートテーブルを削除
  RTS=$(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=${VPC_ID}" --query "RouteTables[?Associations[0].Main==\`false\`].RouteTableId" --output text)
  for RT in ${RTS}; do
    echo -e "${YELLOW}ルートテーブルを削除: ${RT}${NC}"
    aws ec2 delete-route-table --route-table-id ${RT}
  done
  
  # インターネットゲートウェイをデタッチして削除
  IGWs=$(aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=${VPC_ID}" --query "InternetGateways[*].InternetGatewayId" --output text)
  for IGW in ${IGWs}; do
    echo -e "${YELLOW}インターネットゲートウェイをデタッチ: ${IGW}${NC}"
    aws ec2 detach-internet-gateway --internet-gateway-id ${IGW} --vpc-id ${VPC_ID}
    echo -e "${YELLOW}インターネットゲートウェイを削除: ${IGW}${NC}"
    aws ec2 delete-internet-gateway --internet-gateway-id ${IGW}
  done
  
  # VPCを削除
  echo -e "${YELLOW}VPCを削除: ${VPC_ID}${NC}"
  aws ec2 delete-vpc --vpc-id ${VPC_ID}
else
  echo -e "${YELLOW}VPCが見つかりません${NC}"
fi

echo -e "${GREEN}AWS環境のクリーンアップが完了しました${NC}"
echo -e "${BLUE}リソース状態を確認しています...${NC}"
aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=${ENV_PREFIX}" --query "Vpcs[*].[VpcId,CidrBlock]" --output text
aws ecs list-clusters --query "clusterArns[?contains(@,'${ENV_PREFIX}')]" --output text
aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName,'${ENV_PREFIX}')].[LoadBalancerName]" --output text
aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='${ENV_PREFIX}-postgres'].DBInstanceIdentifier" --output text