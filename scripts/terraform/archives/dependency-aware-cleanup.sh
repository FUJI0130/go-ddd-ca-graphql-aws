#!/bin/bash
# 強化版AWSクリーンアップスクリプト - dependency-aware-cleanup.sh

export AWS_REGION=ap-northeast-1
export ENV_PREFIX=development

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}AWS環境の依存関係を考慮したクリーンアップを実行しています...${NC}"

# Phase 1: ECSタスクとサービスの完全停止
echo -e "${BLUE}Phase 1: ECSタスクとサービスを完全に停止しています...${NC}"
CLUSTER=${ENV_PREFIX}-shared-cluster

# ECSサービスに関連するタスクを確実にすべて停止
SERVICES=$(aws ecs list-services --cluster ${CLUSTER} --query "serviceArns" --output text)
if [ ! -z "${SERVICES}" ]; then
  for SERVICE in ${SERVICES}; do
    SERVICE_NAME=$(echo ${SERVICE} | cut -d '/' -f 3)
    echo -e "${YELLOW}サービスタスク数を0に設定: ${SERVICE_NAME}${NC}"
    aws ecs update-service --cluster ${CLUSTER} --service ${SERVICE_NAME} --desired-count 0
  done
  
  # タスクが完全に終了するのを待機
  echo -e "${BLUE}すべてのタスクが終了するまで待機しています...${NC}"
  sleep 30
  
  # 実行中のタスクを強制停止
  TASKS=$(aws ecs list-tasks --cluster ${CLUSTER} --query "taskArns" --output text)
  if [ ! -z "${TASKS}" ]; then
    for TASK in ${TASKS}; do
      echo -e "${YELLOW}タスクを強制停止: ${TASK}${NC}"
      aws ecs stop-task --cluster ${CLUSTER} --task ${TASK} --reason "Forced cleanup"
    done
    
    # タスクが完全に終了するのを待機
    echo -e "${BLUE}すべてのタスクが終了するのを待機しています...${NC}"
    sleep 30
  fi
  
  # ECSサービスを強制削除
  for SERVICE in ${SERVICES}; do
    SERVICE_NAME=$(echo ${SERVICE} | cut -d '/' -f 3)
    echo -e "${YELLOW}サービスを削除: ${SERVICE_NAME}${NC}"
    aws ecs delete-service --cluster ${CLUSTER} --service ${SERVICE_NAME} --force
  done
fi

# Phase 2: ロードバランサーとリソースの削除
echo -e "${BLUE}Phase 2: ロードバランサーとリスナーを削除しています...${NC}"
LBS=$(aws elbv2 describe-load-balancers --query "LoadBalancers[?contains(LoadBalancerName,'${ENV_PREFIX}')].LoadBalancerArn" --output text)

for LB in ${LBS}; do
  LB_NAME=$(aws elbv2 describe-load-balancers --load-balancer-arns ${LB} --query "LoadBalancers[0].LoadBalancerName" --output text)
  echo -e "${YELLOW}ロードバランサーを削除: ${LB_NAME}${NC}"
  
  # リスナーを先に削除
  LISTENERS=$(aws elbv2 describe-listeners --load-balancer-arn ${LB} --query "Listeners[*].ListenerArn" --output text)
  for LISTENER in ${LISTENERS}; do
    echo -e "${YELLOW}リスナーを削除: ${LISTENER}${NC}"
    aws elbv2 delete-listener --listener-arn ${LISTENER}
  done
  
  # ロードバランサーを削除
  aws elbv2 delete-load-balancer --load-balancer-arn ${LB}
done

# Phase 3: ターゲットグループの削除
echo -e "${BLUE}Phase 3: ターゲットグループを削除しています...${NC}"
TGS=$(aws elbv2 describe-target-groups --query "TargetGroups[?contains(TargetGroupName,'${ENV_PREFIX}')].TargetGroupArn" --output text)

for TG in ${TGS}; do
  TG_NAME=$(aws elbv2 describe-target-groups --target-group-arns ${TG} --query "TargetGroups[0].TargetGroupName" --output text)
  echo -e "${YELLOW}ターゲットグループを削除: ${TG_NAME}${NC}"
  
  # ターゲットを登録解除
  TARGETS=$(aws elbv2 describe-target-health --target-group-arn ${TG} --query "TargetHealthDescriptions[*].Target.Id" --output text)
  for TARGET in ${TARGETS}; do
    echo -e "${YELLOW}ターゲットを登録解除: ${TARGET}${NC}"
    aws elbv2 deregister-target --target-group-arn ${TG} --targets Id=${TARGET}
  done
  
  # ターゲットグループを削除
  aws elbv2 delete-target-group --target-group-arn ${TG}
done

# Phase 4: RDSインスタンスの削除完了を待機
echo -e "${BLUE}Phase 4: RDSインスタンスの削除完了を待機しています...${NC}"
RDS_INSTANCES=$(aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='${ENV_PREFIX}-postgres'].DBInstanceIdentifier" --output text)

if [ ! -z "${RDS_INSTANCES}" ]; then
  for RDS in ${RDS_INSTANCES}; do
    echo -e "${YELLOW}RDSインスタンスの削除状態確認: ${RDS}${NC}"
    STATUS=$(aws rds describe-db-instances --db-instance-identifier ${RDS} --query "DBInstances[0].DBInstanceStatus" --output text 2>/dev/null || echo "deleted")
    
    if [ "${STATUS}" != "deleted" ]; then
      echo -e "${YELLOW}現在のステータス: ${STATUS} - RDSがまだ削除中です${NC}"
      echo -e "${YELLOW}RDSインスタンスを再度削除: ${RDS}${NC}"
      aws rds delete-db-instance --db-instance-identifier ${RDS} --skip-final-snapshot --force
      echo -e "${BLUE}RDSインスタンスの削除が完了するまで待機しています (最大5分)...${NC}"
      timeout 300 aws rds wait db-instance-deleted --db-instance-identifier ${RDS}
    else
      echo -e "${GREEN}RDSインスタンスは既に削除されています${NC}"
    fi
  done
fi

# Phase 5: Elastic IPの解放とNAT Gatewayの削除
echo -e "${BLUE}Phase 5: Elastic IPとNAT Gatewayを削除しています...${NC}"
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=${ENV_PREFIX}" --query "Vpcs[0].VpcId" --output text)

# NAT Gatewayの削除
if [ ! -z "${VPC_ID}" ] && [ "${VPC_ID}" != "None" ]; then
  NAT_GATEWAYS=$(aws ec2 describe-nat-gateways --filter "Name=vpc-id,Values=${VPC_ID}" --query "NatGateways[?State!='deleted'].NatGatewayId" --output text)
  
  for NG in ${NAT_GATEWAYS}; do
    echo -e "${YELLOW}NAT Gatewayを削除: ${NG}${NC}"
    aws ec2 delete-nat-gateway --nat-gateway-id ${NG}
  done
  
  # NAT Gatewayの削除には時間がかかるため待機
  if [ ! -z "${NAT_GATEWAYS}" ]; then
    echo -e "${BLUE}NAT Gatewayの削除が完了するまで待機しています (60秒)...${NC}"
    sleep 60
  fi
fi

# Elastic IPの解放
EIPS=$(aws ec2 describe-addresses --query "Addresses[*].AllocationId" --output text)
for EIP in ${EIPS}; do
  echo -e "${YELLOW}Elastic IPを解放: ${EIP}${NC}"
  aws ec2 release-address --allocation-id ${EIP} || echo -e "${YELLOW}このEIPは既に解放されているか、使用中です${NC}"
done

# Phase 6: ECSクラスターの削除
echo -e "${BLUE}Phase 6: ECSクラスターを削除しています...${NC}"
CLUSTERS=$(aws ecs list-clusters --query "clusterArns[?contains(@,'${ENV_PREFIX}')]" --output text)

for CLUSTER in ${CLUSTERS}; do
  CLUSTER_NAME=$(echo ${CLUSTER} | cut -d '/' -f 2)
  echo -e "${YELLOW}クラスターを削除: ${CLUSTER_NAME}${NC}"
  aws ecs delete-cluster --cluster ${CLUSTER_NAME} || echo -e "${YELLOW}クラスターの削除に失敗しました - 依存関係をさらに確認します${NC}"
  
  # クラスターに残っているタスクを再確認して強制停止
  REMAINING_TASKS=$(aws ecs list-tasks --cluster ${CLUSTER_NAME} --query "taskArns" --output text 2>/dev/null || echo "")
  if [ ! -z "${REMAINING_TASKS}" ]; then
    echo -e "${RED}クラスターにまだタスクが残っています - 強制停止します${NC}"
    for TASK in ${REMAINING_TASKS}; do
      aws ecs stop-task --cluster ${CLUSTER_NAME} --task ${TASK} --reason "Emergency cleanup"
    done
    sleep 30
    aws ecs delete-cluster --cluster ${CLUSTER_NAME}
  fi
done

# Phase 7: セキュリティグループの依存関係を解決して削除
echo -e "${BLUE}Phase 7: セキュリティグループの依存関係を解決しています...${NC}"
# VPC内のすべてのENIを検索して削除
if [ ! -z "${VPC_ID}" ] && [ "${VPC_ID}" != "None" ]; then
  ENIS=$(aws ec2 describe-network-interfaces --filters "Name=vpc-id,Values=${VPC_ID}" --query "NetworkInterfaces[*].NetworkInterfaceId" --output text)
  
  for ENI in ${ENIS}; do
    echo -e "${YELLOW}ネットワークインターフェースの削除を試行: ${ENI}${NC}"
    # ENIの強制デタッチを試みる
    ATTACHMENT=$(aws ec2 describe-network-interfaces --network-interface-ids ${ENI} --query "NetworkInterfaces[0].Attachment.AttachmentId" --output text 2>/dev/null || echo "")
    if [ ! -z "${ATTACHMENT}" ] && [ "${ATTACHMENT}" != "None" ]; then
      echo -e "${YELLOW}アタッチメントをデタッチ: ${ATTACHMENT}${NC}"
      aws ec2 detach-network-interface --attachment-id ${ATTACHMENT} --force
      sleep 5
    fi
    
    # ENIを削除
    aws ec2 delete-network-interface --network-interface-id ${ENI} || echo -e "${YELLOW}このENIはまだ使用中か、既に削除されています${NC}"
  done
fi

# セキュリティグループを削除
echo -e "${BLUE}セキュリティグループを削除しています...${NC}"
SGs=$(aws ec2 describe-security-groups --filters "Name=vpc-id,Values=${VPC_ID}" --query "SecurityGroups[?GroupName!='default'].GroupId" --output text)

for SG in ${SGs}; do
  SG_NAME=$(aws ec2 describe-security-groups --group-ids ${SG} --query "SecurityGroups[0].GroupName" --output text 2>/dev/null || echo "unknown")
  echo -e "${YELLOW}セキュリティグループを削除: ${SG_NAME} (${SG})${NC}"
  aws ec2 delete-security-group --group-id ${SG} || echo -e "${YELLOW}このセキュリティグループはまだ依存関係があるか、既に削除されています${NC}"
done

# 60秒待機してリソース削除が完了するのを待つ
echo -e "${BLUE}リソース削除が完了するのを待機しています (60秒)...${NC}"
sleep 60

# Phase 8: VPC関連リソースの削除
echo -e "${BLUE}Phase 8: VPC関連リソースを削除しています...${NC}"
if [ ! -z "${VPC_ID}" ] && [ "${VPC_ID}" != "None" ]; then
  echo -e "${YELLOW}VPC ID: ${VPC_ID}${NC}"
  
  # インターネットゲートウェイをデタッチして削除
  IGWs=$(aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=${VPC_ID}" --query "InternetGateways[*].InternetGatewayId" --output text)
  for IGW in ${IGWs}; do
    echo -e "${YELLOW}インターネットゲートウェイをデタッチ: ${IGW}${NC}"
    aws ec2 detach-internet-gateway --internet-gateway-id ${IGW} --vpc-id ${VPC_ID} || echo -e "${YELLOW}デタッチに失敗 - 既にデタッチされているか、依存関係があります${NC}"
    
    echo -e "${YELLOW}インターネットゲートウェイを削除: ${IGW}${NC}"
    aws ec2 delete-internet-gateway --internet-gateway-id ${IGW} || echo -e "${YELLOW}削除に失敗 - 依存関係があります${NC}"
  done
  
  # ルートテーブルを削除
  RTs=$(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=${VPC_ID}" --query "RouteTables[?Associations[0].Main==\`false\` || length(Associations)==\`0\`].RouteTableId" --output text)
  for RT in ${RTs}; do
    # アソシエーションを解除
    ASSOCIATIONS=$(aws ec2 describe-route-tables --route-table-ids ${RT} --query "RouteTables[0].Associations[*].RouteTableAssociationId" --output text)
    for ASSOC in ${ASSOCIATIONS}; do
      echo -e "${YELLOW}ルートテーブルのアソシエーションを解除: ${ASSOC}${NC}"
      aws ec2 disassociate-route-table --association-id ${ASSOC} || echo -e "${YELLOW}アソシエーション解除に失敗${NC}"
    done
    
    echo -e "${YELLOW}ルートテーブルを削除: ${RT}${NC}"
    aws ec2 delete-route-table --route-table-id ${RT} || echo -e "${YELLOW}ルートテーブルの削除に失敗 - 依存関係があります${NC}"
  done
  
  # サブネットを削除
  SUBNETS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=${VPC_ID}" --query "Subnets[*].SubnetId" --output text)
  for SUBNET in ${SUBNETS}; do
    echo -e "${YELLOW}サブネットを削除: ${SUBNET}${NC}"
    aws ec2 delete-subnet --subnet-id ${SUBNET} || echo -e "${YELLOW}サブネットの削除に失敗 - 依存関係があります${NC}"
  done
  
  # VPCを削除
  echo -e "${YELLOW}VPCを削除: ${VPC_ID}${NC}"
  aws ec2 delete-vpc --vpc-id ${VPC_ID} || echo -e "${YELLOW}VPCの削除に失敗 - 依存関係があります${NC}"
fi

echo -e "${GREEN}AWS環境のクリーンアップが完了しました${NC}"
echo -e "${BLUE}リソース状態を最終確認しています...${NC}"

# 最終確認
echo -e "${BLUE}VPC:${NC}"
aws ec2 describe-vpcs --filters "Name=tag:Environme