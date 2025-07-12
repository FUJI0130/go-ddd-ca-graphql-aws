#!/bin/bash
# ===================================================================
# ファイル名: safe-cleanup-grpc.sh
# 説明: 安全なAWSリソースクリーンアップスクリプト
# 
# 用途:
#  - AWSリソースを安全に段階的に削除する
#  - リソース間の依存関係を考慮した削除順序
#  - ECSサービスのDRAINING状態からの強制削除サポート
# 
# 使用方法:
#  ./safe-cleanup-grpc.sh <環境名>
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
# SERVICE_NAME_GRPC="${ENV_PREFIX}-grpc"
SERVICE_NAME_GRPC="${ENV_PREFIX}-grpc-new"  # -newサフィックス追加

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

# Terraformバックアップの作成
log_info "Terraformステートのバックアップを作成しています..."
cd deployments/terraform/environments/${ENVIRONMENT} || { log_error "環境ディレクトリが見つかりません"; exit 1; }
mkdir -p backup-$(date +%Y%m%d)
cp -r .terraform* terraform.tfstate* backup-$(date +%Y%m%d)/ 2>/dev/null || true
log_success "バックアップが作成されました: backup-$(date +%Y%m%d)/"

# ステップ1: ECSサービスの削除準備（タスク数を0に設定）
log_info "ECSサービスのタスク数を0に設定しています..."

# ECSサービスgRPCの存在確認
if aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME_GRPC} --region ${AWS_REGION:-ap-northeast-1} 2>/dev/null | grep -q "ACTIVE"; then
  log_info "サービス ${SERVICE_NAME_GRPC} が見つかりました。タスク数を0に設定しています..."
  aws ecs update-service --cluster ${CLUSTER_NAME} --service ${SERVICE_NAME_GRPC} --desired-count 0 --region ${AWS_REGION:-ap-northeast-1}
  log_success "サービス ${SERVICE_NAME_GRPC} のタスク数を0に設定しました"
else
  log_warning "サービス ${SERVICE_NAME_GRPC} は存在しないか、既に削除されています"
fi


# タスクの停止を待機
log_info "ECSタスクの停止を待機しています..."
sleep 30

# 実行中のタスクを確認し、強制停止
log_info "実行中のタスクを確認しています..."
RUNNING_TASKS=$(aws ecs list-tasks --cluster ${CLUSTER_NAME} --region ${AWS_REGION:-ap-northeast-1} --query 'taskArns[]' --output text)

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

# ステップ2: Terraformによるリソース削除
log_info "Terraformを使用してリソースを削除しています..."


# HTTPSリスナーの削除
log_info "HTTPSリスナーを削除しています..."
# if terraform state list | grep -q "aws_lb_listener.grpc_https"; then
  # terraform destroy -target=aws_lb_listener.grpc_https -auto-approve || log_warning "HTTPSリスナーの削除に失敗しました"

if terraform state list | grep -q "aws_lb_listener.grpc_https_new"; then
  terraform destroy -target=aws_lb_listener.grpc_https_new -auto-approve || log_warning "HTTPSリスナーの削除に失敗しました"
else
  log_warning "HTTPSリスナーはterraformステートに存在しません"
fi

# 基本サービスの削除
log_info "基本gRPCサービスを削除しています..."
# if terraform state list | grep -q "module.service_grpc"; then
  # terraform destroy -target=module.service_grpc -auto-approve || log_warning "基本gRPCサービスの削除に失敗しました"

if terraform state list | grep -q "module.service_grpc_new"; then
  terraform destroy -target=module.service_grpc_new -auto-approve || log_warning "基本gRPCサービスの削除に失敗しました"
else
  log_warning "基本gRPCサービスはterraformステートに存在しません"
fi

# ターゲットグループの削除
log_info "ターゲットグループを削除しています..."
# if terraform state list | grep -q "module.target_group_grpc_native"; then
#   terraform destroy -target=module.target_group_grpc_native -auto-approve || log_warning "gRPC Nativeターゲットグループの削除に失敗しました"
# fi

# if terraform state list | grep -q "module.target_group_grpc"; then
#   terraform destroy -target=module.target_group_grpc -auto-approve || log_warning "gRPCターゲットグループの削除に失敗しました"

if terraform state list | grep -q "module.target_group_grpc_native_new"; then
  terraform destroy -target=module.target_group_grpc_native_new -auto-approve || log_warning "gRPC Nativeターゲットグループの削除に失敗しました"
fi

if terraform state list | grep -q "module.target_group_grpc_new"; then
  terraform destroy -target=module.target_group_grpc_new -auto-approve || log_warning "gRPCターゲットグループの削除に失敗しました"
fi

# ロードバランサーの削除
log_info "ロードバランサーを削除しています..."
# if terraform state list | grep -q "module.loadbalancer_grpc"; then
#   terraform destroy -target=module.loadbalancer_grpc -auto-approve || log_warning "gRPCロードバランサーの削除に失敗しました"

if terraform state list | grep -q "module.loadbalancer_grpc_new"; then
  terraform destroy -target=module.loadbalancer_grpc_new -auto-approve || log_warning "gRPCロードバランサーの削除に失敗しました"
fi

# 残りのリソースの削除
log_info "残りのリソースを削除しています..."
terraform destroy -auto-approve || log_warning "一部のリソース削除に失敗しました"

# ステップ3: AWS CLIを使用して残ったリソースを確認・削除
log_info "残ったAWSリソースを確認しています..."

# ECSサービスの確認
ECS_SERVICES=$(aws ecs list-services --cluster ${CLUSTER_NAME} --region ${AWS_REGION:-ap-northeast-1} --query 'serviceArns[]' --output text)
if [ ! -z "${ECS_SERVICES}" ]; then
  log_warning "未削除のECSサービスが見つかりました。手動で削除してください:"
  for SERVICE in ${ECS_SERVICES}; do
    echo "- ${SERVICE}"
  done
else
  log_info "未削除のECSサービスはありません"
fi

# ロードバランサーの確認
LBS=$(aws elbv2 describe-load-balancers --region ${AWS_REGION:-ap-northeast-1} --query "LoadBalancers[?contains(LoadBalancerName, '${ENV_PREFIX}')].LoadBalancerArn" --output text)
if [ ! -z "${LBS}" ]; then
  log_warning "未削除のロードバランサーが見つかりました。手動で削除してください:"
  for LB in ${LBS}; do
    echo "- ${LB}"
  done
else
  log_info "未削除のロードバランサーはありません"
fi

# 終了メッセージ
log_success "クリーンアップ処理が完了しました"
log_info "残ったリソースがあれば手動で削除するか、AWSコンソールで確認してください"
log_info "コスト推定を実行するには: make cost-estimate TF_ENV=${ENVIRONMENT}"