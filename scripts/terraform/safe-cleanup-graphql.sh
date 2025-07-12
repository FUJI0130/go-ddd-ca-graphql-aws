#!/bin/bash
# ===================================================================
# ファイル名: safe-cleanup-graphql.sh
# 説明: GraphQLサービスと関連する全リソースの完全クリーンアップスクリプト
# 
# 用途:
#  - GraphQLサービスのAWSリソースを安全に段階的に削除する
#  - リソース間の依存関係を考慮した削除順序
#  - ECSサービス、ロードバランサー、ターゲットグループを削除
#  - RDSやネットワークリソースなどの共有リソースも削除
#  - 完全なクリーンアップを提供
# 
# 使用方法:
#  ./safe-cleanup-graphql.sh <環境名>
#
# 引数:
#  環境名 - クリーンアップする環境（development, production）
# ===================================================================

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 引数の解析
ENVIRONMENT=${1:-development}

# ログ出力関数
log_info() {
  echo -e "${BLUE}[INFO] $1${NC}"
}

log_success() {
  echo -e "${GREEN}[SUCCESS] $1${NC}"
}

log_warning() {
  echo -e "${YELLOW}[WARNING] $1${NC}"
}

log_error() {
  echo -e "${RED}[ERROR] $1${NC}"
}

# 環境固有の設定
ENV_PREFIX="${ENVIRONMENT}"
CLUSTER_NAME="${ENV_PREFIX}-shared-cluster"
SERVICE_NAME_GRAPHQL="${ENV_PREFIX}-graphql-new"  # -newサフィックス対応
DB_INSTANCE_NAME="${ENV_PREFIX}-postgres"

# AWS CLIの存在確認
if ! command -v aws &> /dev/null; then
  log_error "AWS CLIがインストールされていません"
  exit 1
fi

# 認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  log_error "AWS認証情報が無効です"
  exit 1
fi

# 確認メッセージを表示
log_warning "このスクリプトはGraphQLサービスおよび関連する全てのリソース（RDSやネットワークなどの共有リソースを含む）を削除します。"
log_warning "これは完全なクリーンアップであり、他のサービス（REST API、gRPC）のリソースも影響を受ける可能性があります。"
read -p "続行しますか？ (y/n): " CONFIRM
if [[ ! "${CONFIRM}" =~ ^[Yy]$ ]]; then
  log_info "操作がキャンセルされました"
  exit 0
fi

# Terraformバックアップの作成
log_info "Terraformステートのバックアップを作成しています..."
cd deployments/terraform/environments/${ENVIRONMENT} || { log_error "環境ディレクトリが見つかりません"; exit 1; }
mkdir -p backup-$(date +%Y%m%d)
cp -r .terraform* terraform.tfstate* backup-$(date +%Y%m%d)/ 2>/dev/null || true
log_success "バックアップが作成されました: backup-$(date +%Y%m%d)/"

# 実行前のコスト状況を確認して記録
log_info "クリーンアップ前のリソース状況を確認しています..."
AWS_RESOURCES_BEFORE=$(aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='${DB_INSTANCE_NAME}'].DBInstanceStatus" --output text 2>/dev/null)
ECS_SERVICES_BEFORE=$(aws ecs list-services --cluster ${CLUSTER_NAME} --region ${AWS_REGION:-ap-northeast-1} --output text | grep -c "${ENV_PREFIX}" || echo "0")
log_info "確認されたリソース - RDSインスタンス: ${AWS_RESOURCES_BEFORE:-なし}, ECSサービス数: ${ECS_SERVICES_BEFORE:-0}"

# ステップ1: ECSサービスの削除準備（タスク数を0に設定）
log_info "GraphQL ECSサービスのタスク数を0に設定しています..."

# ECSサービスGraphQLの存在確認
if aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME_GRAPHQL} --region ${AWS_REGION:-ap-northeast-1} 2>/dev/null | grep -q "ACTIVE"; then
  log_info "サービス ${SERVICE_NAME_GRAPHQL} が見つかりました。タスク数を0に設定しています..."
  aws ecs update-service --cluster ${CLUSTER_NAME} --service ${SERVICE_NAME_GRAPHQL} --desired-count 0 --region ${AWS_REGION:-ap-northeast-1}
  log_success "サービス ${SERVICE_NAME_GRAPHQL} のタスク数を0に設定しました"
else
  log_warning "サービス ${SERVICE_NAME_GRAPHQL} は存在しないか、既に削除されています"
fi

# タスクの停止を待機
log_info "ECSタスクの停止を待機しています..."
sleep 30

# 実行中のタスクを確認し、強制停止
log_info "実行中のタスクを確認しています..."
RUNNING_TASKS=$(aws ecs list-tasks --cluster ${CLUSTER_NAME} --family ${SERVICE_NAME_GRAPHQL} --region ${AWS_REGION:-ap-northeast-1} --query 'taskArns[]' --output text)

if [ ! -z "${RUNNING_TASKS}" ]; then
  log_warning "実行中のタスクが見つかりました。強制停止します..."
  for TASK in ${RUNNING_TASKS}; do
    aws ecs stop-task --cluster ${CLUSTER_NAME} --task ${TASK} --region ${AWS_REGION:-ap-northeast-1}
    log_info "タスク ${TASK} を停止しました"
  done
else
  log_info "実行中のタスクはありません"
fi

# 十分なタイムアウト時間を待機
log_info "タスクの完全な停止を待機しています..."
sleep 60

# ステップ2: Terraformによるリソース削除（GraphQLサービス関連）
log_info "Terraformを使用してGraphQLリソースを削除しています..."

# GraphQL基本サービスの削除
log_info "GraphQLサービスを削除しています..."
if terraform state list | grep -q "module.service_graphql_new"; then
  terraform destroy -target=module.service_graphql_new -auto-approve || log_warning "GraphQLサービスの削除に失敗しました"
else
  log_warning "GraphQLサービスはterraformステートに存在しません"
fi

# 削除の確認と待機
log_info "GraphQLサービス削除の完了を待機しています..."
sleep 30

# GraphQLロードバランサーの削除
log_info "GraphQLロードバランサーを削除しています..."
if terraform state list | grep -q "module.loadbalancer_graphql_new"; then
  terraform destroy -target=module.loadbalancer_graphql_new -auto-approve || log_warning "GraphQLロードバランサーの削除に失敗しました"
fi

# 削除の確認と待機
log_info "GraphQLロードバランサー削除の完了を待機しています..."
sleep 30

# GraphQLターゲットグループの削除
log_info "GraphQLターゲットグループを削除しています..."
if terraform state list | grep -q "module.target_group_graphql_new"; then
  terraform destroy -target=module.target_group_graphql_new -auto-approve || log_warning "GraphQLターゲットグループの削除に失敗しました"
fi

# 削除の確認と待機
log_info "GraphQLターゲットグループ削除の完了を待機しています..."
sleep 30

# ステップ3: AWS CLIを使用して残ったGraphQL特有リソースを確認・削除
log_info "残ったGraphQL特有のAWSリソースを確認しています..."

# ECSサービスの確認
ECS_SERVICE=$(aws ecs list-services --cluster ${CLUSTER_NAME} --region ${AWS_REGION:-ap-northeast-1} --query "serviceArns[?contains(@, '${SERVICE_NAME_GRAPHQL}')]" --output text)
if [ ! -z "${ECS_SERVICE}" ]; then
  log_warning "未削除のGraphQL ECSサービスが見つかりました。手動で削除を試みます..."
  aws ecs update-service --cluster ${CLUSTER_NAME} --service ${SERVICE_NAME_GRAPHQL} --desired-count 0 --region ${AWS_REGION:-ap-northeast-1}
  aws ecs delete-service --cluster ${CLUSTER_NAME} --service ${SERVICE_NAME_GRAPHQL} --force --region ${AWS_REGION:-ap-northeast-1}
else
  log_info "未削除のGraphQL ECSサービスはありません"
fi

# ロードバランサーの確認
LB=$(aws elbv2 describe-load-balancers --region ${AWS_REGION:-ap-northeast-1} --query "LoadBalancers[?contains(LoadBalancerName, '${ENV_PREFIX}-graphql-new')].LoadBalancerArn" --output text)
if [ ! -z "${LB}" ]; then
  log_warning "未削除のGraphQLロードバランサーが見つかりました。手動で削除を試みます..."
  aws elbv2 delete-load-balancer --load-balancer-arn ${LB} --region ${AWS_REGION:-ap-northeast-1}
else
  log_info "未削除のGraphQLロードバランサーはありません"
fi

# ターゲットグループの確認
TG=$(aws elbv2 describe-target-groups --region ${AWS_REGION:-ap-northeast-1} --query "TargetGroups[?contains(TargetGroupName, '${ENV_PREFIX}-graphql-new')].TargetGroupArn" --output text)
if [ ! -z "${TG}" ]; then
  log_warning "未削除のGraphQLターゲットグループが見つかりました。手動で削除を試みます..."
  aws elbv2 delete-target-group --target-group-arn ${TG} --region ${AWS_REGION:-ap-northeast-1}
else
  log_info "未削除のGraphQLターゲットグループはありません"
fi

# ステップ4: 他のECSサービスの削除（REST API, gRPC）
log_info "他のECSサービスの削除を確認します..."

# REST API ECSサービスの削除
SERVICE_NAME_API="${ENV_PREFIX}-api-new"
if terraform state list | grep -q "module.service_api_new"; then
  log_info "REST API ECSサービスを削除しています..."
  terraform destroy -target=module.service_api_new -auto-approve || log_warning "REST API ECSサービスの削除に失敗しました"
else
  log_info "REST API ECSサービスはterraformステートに存在しないか、既に削除されています"
fi

# gRPC ECSサービスの削除
SERVICE_NAME_GRPC="${ENV_PREFIX}-grpc-new"
if terraform state list | grep -q "module.service_grpc_new"; then
  log_info "gRPC ECSサービスを削除しています..."
  terraform destroy -target=module.service_grpc_new -auto-approve || log_warning "gRPC ECSサービスの削除に失敗しました"
else
  log_info "gRPC ECSサービスはterraformステートに存在しないか、既に削除されています"
fi

# 削除の確認と待機
log_info "ECSサービス削除の完了を待機しています..."
sleep 30

# ステップ5: 共有リソースの削除開始
log_info "共有リソースの削除を開始します..."

# ECSクラスターの削除
log_info "ECSクラスターを削除しています..."
if terraform state list | grep -q "module.shared"; then
  log_info "共有ECSクラスターリソースを削除しています..."
  terraform destroy -target=module.shared -auto-approve || log_warning "共有ECSクラスターの削除に失敗しました"
else
  log_info "共有ECSクラスターはterraformステートに存在しないか、既に削除されています"
fi

# セキュリティグループ削除の確認と待機
log_info "共有リソース削除の完了を待機しています..."
sleep 30

# RDSインスタンスの削除
log_info "RDSインスタンスを削除しています..."
if terraform state list | grep -q "module.database"; then
  terraform destroy -target=module.database -auto-approve || log_warning "RDSインスタンスの削除に失敗しました"
else
  log_info "RDSインスタンスはterraformステートに存在しないか、既に削除されています"
fi

# RDS削除の確認と待機
log_info "RDS削除の完了を待機しています..."
sleep 30

# ネットワークリソースの削除（最後に実行）
log_info "ネットワークリソースを削除しています..."
if terraform state list | grep -q "module.networking"; then
  terraform destroy -target=module.networking -auto-approve || log_warning "ネットワークリソースの削除に失敗しました"
else
  log_info "ネットワークリソースはterraformステートに存在しないか、既に削除されています"
fi

# ステップ6: 最終確認と報告
log_info "リソース削除の最終確認を行っています..."

# RDSインスタンスの確認
RDS_STATUS=$(aws rds describe-db-instances --query "DBInstances[?DBInstanceIdentifier=='${DB_INSTANCE_NAME}'].DBInstanceStatus" --output text 2>/dev/null || echo "削除済み")
if [ "${RDS_STATUS}" == "削除済み" ]; then
  log_success "RDSインスタンス ${DB_INSTANCE_NAME} は正常に削除されました"
else
  log_warning "RDSインスタンス ${DB_INSTANCE_NAME} が残存しています。状態: ${RDS_STATUS}"
  log_info "RDSインスタンスの削除には時間がかかる場合があります。後ほど確認してください。"
fi

# ECSサービスの確認
ECS_SERVICES=$(aws ecs list-services --cluster ${CLUSTER_NAME} --region ${AWS_REGION:-ap-northeast-1} --output text)
if [ -z "${ECS_SERVICES}" ] || ! echo "${ECS_SERVICES}" | grep -q "${ENV_PREFIX}"; then
  log_success "ECSサービスは正常に削除されました"
else
  log_warning "未削除のECSサービスが見つかりました:"
  echo "${ECS_SERVICES}"
fi

# ロードバランサーの確認
LBS=$(aws elbv2 describe-load-balancers --region ${AWS_REGION:-ap-northeast-1} --query "LoadBalancers[?contains(LoadBalancerName, '${ENV_PREFIX}')].LoadBalancerName" --output text)
if [ -z "${LBS}" ]; then
  log_success "ロードバランサーは正常に削除されました"
else
  log_warning "未削除のロードバランサーが見つかりました:"
  echo "${LBS}"
fi

# ターゲットグループの確認
TGS=$(aws elbv2 describe-target-groups --region ${AWS_REGION:-ap-northeast-1} --query "TargetGroups[?contains(TargetGroupName, '${ENV_PREFIX}')].TargetGroupName" --output text)
if [ -z "${TGS}" ]; then
  log_success "ターゲットグループは正常に削除されました"
else
  log_warning "未削除のターゲットグループが見つかりました:"
  echo "${TGS}"
fi

# 最終メッセージ
log_success "GraphQLサービスおよび関連する全てのリソースのクリーンアップ処理が完了しました"
log_info "残ったリソースがあれば手動で削除するか、AWSコンソールで確認してください"
log_info "コスト推定を実行するには: make cost-estimate TF_ENV=${ENVIRONMENT}"

# コスト状況を再確認
log_info "コスト状況を確認しています..."
# make -s cost-estimate TF_ENV=${ENVIRONMENT}