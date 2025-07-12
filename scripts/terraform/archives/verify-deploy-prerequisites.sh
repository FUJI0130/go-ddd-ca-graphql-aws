#!/bin/bash
# ===================================================================
# ファイル名: verify-deploy-prerequisites.sh
# 説明: デプロイ前の前提条件を検証するスクリプト
#
# 用途:
#  - AWS環境の残存リソースを検出
#  - デプロイを妨げる可能性のある問題を早期に発見
#  - 必要に応じてクリーンアップを推奨
#
# 引数:
#  $1: 環境名（development, productionなど）
# ===================================================================

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 共通ライブラリのインポート
SCRIPT_DIR="$(dirname "$0")"
source "$SCRIPT_DIR/common/aws_resource_utils.sh"

# デバッグフラグ
DEBUG=${DEBUG:-false}

# 引数の解析
ENV=${1:-development}
STRICT_MODE=${STRICT_MODE:-true}
SKIP_STATE_VERIFY=${SKIP_STATE_VERIFY:-false}

echo -e "${BLUE}デプロイ前の前提条件を検証しています（環境: $ENV）...${NC}"

# デバッグ情報
if [ "$DEBUG" = "true" ] || [ "$AWS_RESOURCE_DEBUG" = "true" ]; then
  aws_resource_debug "ENV=$ENV"
  aws_resource_debug "STRICT_MODE=$STRICT_MODE"
  aws_resource_debug "SKIP_STATE_VERIFY=$SKIP_STATE_VERIFY"
  aws_resource_debug "TF_VAR_db_username=${TF_VAR_db_username:-(未設定)}"
  aws_resource_debug "TF_VAR_db_password=${TF_VAR_db_password:-(未設定)}"
  
  # 環境情報の詳細出力
  aws_resource_dump_env
fi

# 1. 環境変数の検証
echo -e "${BLUE}ステップ 1/3: 環境変数を検証${NC}"
ENV_OK=true

# 環境変数ファイルの自動読み込み
ENV_FILE="$HOME/.env.terraform"
if [ -f "$ENV_FILE" ]; then
  echo -e "${BLUE}環境変数ファイルを読み込みます: $ENV_FILE${NC}"
  source "$ENV_FILE"
  
  if [ "$DEBUG" = "true" ] || [ "$AWS_RESOURCE_DEBUG" = "true" ]; then
    aws_resource_debug "環境変数ファイル読み込み後"
    aws_resource_debug "TF_VAR_db_username=${TF_VAR_db_username:-(未設定)}"
    aws_resource_debug "TF_VAR_db_password=${TF_VAR_db_password:-(未設定)}"
  fi
fi

# 必須環境変数のチェック
if [ -z "${TF_VAR_db_username}" ]; then
  aws_resource_warning "TF_VAR_db_username が設定されていません"
  ENV_OK=false
fi

if [ -z "${TF_VAR_db_password}" ]; then
  aws_resource_warning "TF_VAR_db_password が設定されていません"
  ENV_OK=false
fi

if [ "$ENV_OK" = "false" ]; then
  if [ "$STRICT_MODE" = "true" ]; then
    aws_resource_error "環境変数の検証に失敗しました。デプロイを中止します。"
    echo -e "${YELLOW}以下の環境変数を設定してください:${NC}"
    echo -e "export TF_VAR_db_username=\"適切なユーザー名\""
    echo -e "export TF_VAR_db_password=\"適切なパスワード\""
    echo -e "${YELLOW}または、~/.env.terraform ファイルを作成して認証情報を保存してください${NC}"
    exit 1
  else
    aws_resource_warning "環境変数が不足していますが、緩和モードで続行します"
  fi
else
  aws_resource_success "環境変数の検証に成功しました"
fi

# 2. AWS環境のクリーン状態確認
echo -e "${BLUE}ステップ 2/3: AWS環境のクリーン状態を検証${NC}"

# 既存のリソースをチェック
EXISTING_RESOURCES=false

# VPCの確認
VPC_COUNT=$(aws_cli_exec ec2 describe-vpcs --filters "Name=tag:Environment,Values=$ENV" --query "length(Vpcs)" --output text)
if [ "$VPC_COUNT" -gt 0 ]; then
  aws_resource_warning "$ENV 環境のVPCが $VPC_COUNT 個存在します"
  EXISTING_RESOURCES=true
fi

# RDSの確認
RDS_COUNT=$(aws_cli_exec rds describe-db-instances --query "length(DBInstances[?DBInstanceIdentifier=='$ENV-postgres'])" --output text)
if [ "$RDS_COUNT" -gt 0 ]; then
  aws_resource_warning "$ENV-postgres RDSインスタンスが存在します"
  EXISTING_RESOURCES=true
fi

# ALBの確認
ALB_COUNT=$(aws_cli_exec elbv2 describe-load-balancers --query "length(LoadBalancers[?contains(LoadBalancerName,'$ENV')])" --output text)
if [ "$ALB_COUNT" -gt 0 ]; then
  aws_resource_warning "$ENV 環境のALBが $ALB_COUNT 個存在します"
  EXISTING_RESOURCES=true
fi

# ECSクラスターの確認（共通ライブラリ関数を使用）
CLUSTER_NAME="$ENV-shared-cluster"
if ecs_cluster_exists "$CLUSTER_NAME"; then
  aws_resource_warning "$ENV 環境のECSクラスターが存在します"
  
  # サービスの確認（共通ライブラリ関数を使用）
  services=$(ecs_list_services "$CLUSTER_NAME")
  if [ ! -z "$services" ]; then
    service_count=$(echo "$services" | wc -w)
    aws_resource_warning "クラスター内に $service_count 個のサービスが存在します"
  fi
  
  EXISTING_RESOURCES=true
fi

# ターゲットグループの確認
TG_COUNT=$(aws_cli_exec elbv2 describe-target-groups --query "length(TargetGroups[?contains(TargetGroupName,'$ENV')])" --output text)
if [ "$TG_COUNT" -gt 0 ]; then
  aws_resource_warning "$ENV 環境のターゲットグループが $TG_COUNT 個存在します"
  EXISTING_RESOURCES=true
fi

# 修正後
if [ "$EXISTING_RESOURCES" = "true" ]; then
  if [ "$STRICT_MODE" = "true" ]; then
    aws_resource_error "AWS環境が完全にクリーンではありません。リソースが残存しています。"
    echo -e "${YELLOW}残存しているリソース:${NC}"
    
    # 残存リソースの詳細表示
    if [ "$VPC_COUNT" -gt 0 ]; then
      echo -e "  - VPC: ${VPC_COUNT}個"
    fi
    if [ "$RDS_COUNT" -gt 0 ]; then
      echo -e "  - RDSインスタンス: ${RDS_COUNT}個"
    fi
    if [ "$ALB_COUNT" -gt 0 ]; then
      echo -e "  - ロードバランサー: ${ALB_COUNT}個"
    fi
    if ecs_cluster_exists "$CLUSTER_NAME"; then
      echo -e "  - ECSクラスター: 1個"
    fi
    if [ "$TG_COUNT" -gt 0 ]; then
      echo -e "  - ターゲットグループ: ${TG_COUNT}個"
    fi
    
    echo -e "${YELLOW}以下のコマンドでクリーンアップを実行してください:${NC}"
    echo -e "make stop-api-dev TF_ENV=$ENV"
    exit 1
  else
    aws_resource_warning "AWS環境に残存リソースがありますが、緩和モードで続行します"
  fi
else
  aws_resource_success "AWS環境はクリーンです"
fi

# 3. Terraform状態の整合性確認（オプション）
if [ "$SKIP_STATE_VERIFY" = "false" ]; then
  echo -e "${BLUE}ステップ 3/3: Terraform状態の整合性を検証${NC}"
  
  # Goツールを優先、なければシェルスクリプトにフォールバック
  VERIFY_BIN="$SCRIPT_DIR/../../bin/verify-terraform"
  VERIFY_SCRIPT="$SCRIPT_DIR/aws-terraform-verify.sh"
  
  if [ -f "$VERIFY_BIN" ]; then
    # Goツールを使用
    echo -e "${BLUE}Goツールを使用して検証します...${NC}"
    if "$VERIFY_BIN" -env "$ENV" --ignore-resource-errors; then
      aws_resource_success "Terraform状態の検証に成功しました"
    else
      VERIFY_EXIT_CODE=$?
      if [ "$STRICT_MODE" = "true" ]; then
        aws_resource_error "Terraform状態と実際の環境に不整合があります"
        echo -e "${YELLOW}以下のオプションがあります:${NC}"
        echo -e "1. make terraform-import TF_ENV=$ENV     - 既存リソースをインポート"
        echo -e "2. make terraform-reset TF_ENV=$ENV      - 状態をリセット"
        echo -e "3. SKIP_STATE_VERIFY=true を設定して続行"
        exit $VERIFY_EXIT_CODE
      else
        aws_resource_warning "Terraform状態に不整合がありますが、緩和モードで続行します"
      fi
    fi
  elif [ -f "$VERIFY_SCRIPT" ]; then
    # シェルスクリプトにフォールバック
    echo -e "${YELLOW}シェルスクリプトにフォールバックして検証します...${NC}"
    chmod +x "$VERIFY_SCRIPT"
    if "$VERIFY_SCRIPT" "$ENV"; then
      aws_resource_success "Terraform状態の検証に成功しました"
    else
      VERIFY_EXIT_CODE=$?
      if [ "$STRICT_MODE" = "true" ]; then
        aws_resource_error "Terraform状態と実際の環境に不整合があります"
        echo -e "${YELLOW}以下のオプションがあります:${NC}"
        echo -e "1. make terraform-import TF_ENV=$ENV     - 既存リソースをインポート"
        echo -e "2. make terraform-reset TF_ENV=$ENV      - 状態をリセット"
        echo -e "3. SKIP_STATE_VERIFY=true を設定して続行"
        exit $VERIFY_EXIT_CODE
      else
        aws_resource_warning "Terraform状態に不整合がありますが、緩和モードで続行します"
      fi
    fi
  else
    # どちらも見つからない場合
    aws_resource_warning "検証ツールが見つかりません。検証をスキップします。"
    if [ "$STRICT_MODE" = "true" ]; then
      # ビルドを試行
      echo -e "${YELLOW}検証ツールバイナリが見つかりません。ビルドを試行します...${NC}"
      make -C "$SCRIPT_DIR/../../" build-terraform-verify
      
      # 再確認
      if [ -f "$VERIFY_BIN" ]; then
        if "$VERIFY_BIN" -env "$ENV" --ignore-resource-errors; then
          aws_resource_success "Terraform状態の検証に成功しました"
        else
          VERIFY_EXIT_CODE=$?
          aws_resource_error "Terraform状態の検証に失敗しました"
          exit $VERIFY_EXIT_CODE
        fi
      else
        aws_resource_error "検証ツールが使用できないため、デプロイを中止します"
        exit 1
      fi
    fi
  fi
else
  echo -e "${YELLOW}ステップ 3/3: Terraform状態の検証をスキップします${NC}"
fi

aws_resource_success "すべての前提条件の検証に成功しました！デプロイを続行できます。"
exit 0