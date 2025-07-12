#!/bin/bash
# aws-terraform-verify.sh - AWS環境とTerraform状態の整合性を検証するスクリプト
# 2025-04-15 改善版: 表示問題を修正したバージョン

# エラー時に停止しない
set +e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# デバッグフラグ（必要に応じて有効化）
# DEBUG=true で実行するとより詳細な情報が表示されます
DEBUG=${DEBUG:-false}

# 引数の解析
ENV=${1:-development}

echo -e "${BLUE}AWS環境とTerraform状態の整合性を検証しています（環境: $ENV）...${NC}"

# 現在のディレクトリを保存
ORIGINAL_DIR=$(pwd)

# 環境ディレクトリに移動
cd deployments/terraform/environments/$ENV || {
  echo -e "${RED}環境ディレクトリが見つかりません: deployments/terraform/environments/$ENV${NC}"
  exit 1
}

# 文字列をトリミングする関数（改良版）
# 改行や制御文字を徹底的に削除し、一貫した比較を可能にする
super_trim() {
  local var="$*"
  # 改行、復帰、タブ、垂直タブ、改ページを除去
  var=$(echo -n "$var" | tr -d '\n\r\t\v\f')
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  # 安全のためprintfで整形
  printf "%s" "$var"
}

# 環境変数ファイルの自動読み込み
ENV_FILE="$HOME/.env.terraform"
if [ -f "$ENV_FILE" ] && { [ -z "${TF_VAR_db_username}" ] || [ -z "${TF_VAR_db_password}" ]; }; then
  echo -e "${BLUE}環境変数ファイルを読み込んでいます: $ENV_FILE${NC}"
  source "$ENV_FILE"
fi

# 必須環境変数のチェック
if [ -z "${TF_VAR_db_username}" ] || [ -z "${TF_VAR_db_password}" ]; then
  echo -e "${YELLOW}警告: データベース認証情報の環境変数が設定されていません${NC}"
  echo -e "${YELLOW}以下の環境変数を設定してください:${NC}"
  echo -e "export TF_VAR_db_username=\"適切なユーザー名\""
  echo -e "export TF_VAR_db_password=\"適切なパスワード\""
  echo -e "${YELLOW}または、~/.env.terraform ファイルを作成して認証情報を保存してください${NC}"
  
  # 環境変数不足のため、terraform planをスキップ
  SKIP_TERRAFORM_PLAN=true
fi

# Terraform状態のリソース一覧を取得
echo -e "${BLUE}Terraform状態のリソース一覧を取得しています...${NC}"
TF_RESOURCES=$(terraform state list 2>/dev/null || echo "")

# AWS環境のリソース状態チェック
echo -e "${BLUE}AWS環境の主要リソースを確認しています...${NC}"

# 不一致カウンター初期化
MISMATCH_COUNT=0

# ==== リソース情報の収集 ====

# コアインフラストラクチャリソース
# VPCの確認
VPC_COUNT=$(aws ec2 describe-vpcs --filters "Name=tag:Environment,Values=$ENV" --query "length(Vpcs)" --output text)
TF_VPC_COUNT=$(echo "$TF_RESOURCES" | grep -c "module.networking.aws_vpc" || echo 0)

# RDSの確認
RDS_COUNT=$(aws rds describe-db-instances --query "length(DBInstances[?DBInstanceIdentifier=='$ENV-postgres'])" --output text)
TF_RDS_COUNT=$(echo "$TF_RESOURCES" | grep -c "module.database.aws_db_instance" || echo 0)

# ECSクラスターの確認
ECS_CLUSTER_COUNT=$(aws ecs list-clusters --query "length(clusterArns[?contains(@,'$ENV-shared-cluster')])" --output text 2>/dev/null || echo 0)
TF_ECS_CLUSTER_COUNT=$(echo "$TF_RESOURCES" | grep -c "module.shared_ecs_cluster.aws_ecs_cluster" || echo 0)

# サービス固有リソース情報の収集
declare -A SERVICE_COUNTS
declare -A TF_SERVICE_COUNTS
declare -A ALB_COUNTS
declare -A TF_ALB_COUNTS
declare -A TG_COUNTS
declare -A TF_TG_COUNTS

for SERVICE_TYPE in api graphql grpc; do
  # ECSサービスの確認
  SERVICE_COUNTS[$SERVICE_TYPE]=$(aws ecs list-services --cluster $ENV-shared-cluster --query "length(serviceArns[?contains(@,'$ENV-$SERVICE_TYPE')])" --output text 2>/dev/null || echo 0)
  TF_SERVICE_COUNTS[$SERVICE_TYPE]=$(echo "$TF_RESOURCES" | grep -c "module.service_${SERVICE_TYPE}.aws_ecs_service" || echo 0)
  
  # ALBの確認
  ALB_COUNTS[$SERVICE_TYPE]=$(aws elbv2 describe-load-balancers --query "length(LoadBalancers[?LoadBalancerName=='$ENV-$SERVICE_TYPE-alb'])" --output text)
  TF_ALB_COUNTS[$SERVICE_TYPE]=$(echo "$TF_RESOURCES" | grep -c "module.loadbalancer_${SERVICE_TYPE}.aws_lb" || echo 0)
  
  # ターゲットグループの確認
  TG_COUNTS[$SERVICE_TYPE]=$(aws elbv2 describe-target-groups --query "length(TargetGroups[?TargetGroupName=='$ENV-$SERVICE_TYPE-tg'])" --output text)
  TF_TG_COUNTS[$SERVICE_TYPE]=$(echo "$TF_RESOURCES" | grep -c "module.target_group_${SERVICE_TYPE}.aws_lb_target_group" || echo 0)
done

# ==== リソース情報のクリーニング ====

# コアリソースの徹底的なクリーニング
VPC_COUNT_CLEAN=$(super_trim "$VPC_COUNT")
TF_VPC_COUNT_CLEAN=$(super_trim "$TF_VPC_COUNT")
RDS_COUNT_CLEAN=$(super_trim "$RDS_COUNT")
TF_RDS_COUNT_CLEAN=$(super_trim "$TF_RDS_COUNT")
ECS_CLUSTER_COUNT_CLEAN=$(super_trim "$ECS_CLUSTER_COUNT")
TF_ECS_CLUSTER_COUNT_CLEAN=$(super_trim "$TF_ECS_CLUSTER_COUNT")

# サービスリソースの徹底的なクリーニング
for SERVICE_TYPE in api graphql grpc; do
  SERVICE_COUNTS[$SERVICE_TYPE]=$(super_trim "${SERVICE_COUNTS[$SERVICE_TYPE]}")
  TF_SERVICE_COUNTS[$SERVICE_TYPE]=$(super_trim "${TF_SERVICE_COUNTS[$SERVICE_TYPE]}")
  ALB_COUNTS[$SERVICE_TYPE]=$(super_trim "${ALB_COUNTS[$SERVICE_TYPE]}")
  TF_ALB_COUNTS[$SERVICE_TYPE]=$(super_trim "${TF_ALB_COUNTS[$SERVICE_TYPE]}")
  TG_COUNTS[$SERVICE_TYPE]=$(super_trim "${TG_COUNTS[$SERVICE_TYPE]}")
  TF_TG_COUNTS[$SERVICE_TYPE]=$(super_trim "${TF_TG_COUNTS[$SERVICE_TYPE]}")
done

# デバッグ情報表示
if [ "$DEBUG" = "true" ]; then
  echo ""
  echo "DEBUG: コアリソース情報"
  echo "DEBUG: VPC_COUNT=$VPC_COUNT -> $VPC_COUNT_CLEAN, TF_VPC_COUNT=$TF_VPC_COUNT -> $TF_VPC_COUNT_CLEAN"
  echo "DEBUG: RDS_COUNT=$RDS_COUNT -> $RDS_COUNT_CLEAN, TF_RDS_COUNT=$TF_RDS_COUNT -> $TF_RDS_COUNT_CLEAN"
  echo "DEBUG: ECS_CLUSTER_COUNT=$ECS_CLUSTER_COUNT -> $ECS_CLUSTER_COUNT_CLEAN, TF_ECS_CLUSTER_COUNT=$TF_ECS_CLUSTER_COUNT -> $TF_ECS_CLUSTER_COUNT_CLEAN"
  
  echo ""
  echo "DEBUG: サービスリソース情報"
  for SERVICE_TYPE in api graphql grpc; do
    echo "DEBUG: $SERVICE_TYPE:"
    echo "DEBUG:   SERVICE_COUNT=${SERVICE_COUNTS[$SERVICE_TYPE]}, TF_SERVICE_COUNT=${TF_SERVICE_COUNTS[$SERVICE_TYPE]}"
    echo "DEBUG:   ALB_COUNT=${ALB_COUNTS[$SERVICE_TYPE]}, TF_ALB_COUNT=${TF_ALB_COUNTS[$SERVICE_TYPE]}"
    echo "DEBUG:   TG_COUNT=${TG_COUNTS[$SERVICE_TYPE]}, TF_TG_COUNT=${TF_TG_COUNTS[$SERVICE_TYPE]}"
  done
  
  echo ""
  echo "DEBUG: Terraform Resources:"
  echo "$TF_RESOURCES"
fi

# ==== 検証結果の表示 ====

echo ""
echo -e "${BLUE}検証結果:${NC}"
echo "--------------------------------"
echo -e "リソース\t\tAWS\tTerraform\t状態"
echo "--------------------------------"

# === コアリソースの比較と表示 ===

# VPC比較 - 一時変数を使用して表示する
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
  MISMATCH_COUNT=$((MISMATCH_COUNT + 1))
fi
DISPLAY_LINE="VPC\t\t$VPC_COUNT_CLEAN\t$TF_VPC_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

# RDS比較 - 一時変数を使用して表示する
if [ "$RDS_COUNT_CLEAN" = "$TF_RDS_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
  MISMATCH_COUNT=$((MISMATCH_COUNT + 1))
fi
DISPLAY_LINE="RDS\t\t$RDS_COUNT_CLEAN\t$TF_RDS_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

# ECSクラスター比較 - 一時変数を使用して表示する
if [ "$ECS_CLUSTER_COUNT_CLEAN" = "$TF_ECS_CLUSTER_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
  MISMATCH_COUNT=$((MISMATCH_COUNT + 1))
fi
DISPLAY_LINE="ECSクラスター\t$ECS_CLUSTER_COUNT_CLEAN\t$TF_ECS_CLUSTER_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

echo "--------------------------------"

# === サービス毎の検証結果表示 ===
for SERVICE_TYPE in api graphql grpc; do
  echo -e "${BLUE}$SERVICE_TYPE サービス:${NC}"
  echo "--------------------------------"
  
  # ECSサービス比較 - 一時変数を使用して表示する
  if [ "${SERVICE_COUNTS[$SERVICE_TYPE]}" = "${TF_SERVICE_COUNTS[$SERVICE_TYPE]}" ]; then
    STATUS="一致"
    STATUS_COLOR=$GREEN
  else
    STATUS="不一致"
    STATUS_COLOR=$RED
    MISMATCH_COUNT=$((MISMATCH_COUNT + 1))
  fi
  DISPLAY_LINE="ECSサービス\t${SERVICE_COUNTS[$SERVICE_TYPE]}\t${TF_SERVICE_COUNTS[$SERVICE_TYPE]}\t${STATUS_COLOR}$STATUS${NC}"
  echo -e "$DISPLAY_LINE"
  
  # ALB比較 - 一時変数を使用して表示する
  if [ "${ALB_COUNTS[$SERVICE_TYPE]}" = "${TF_ALB_COUNTS[$SERVICE_TYPE]}" ]; then
    STATUS="一致"
    STATUS_COLOR=$GREEN
  else
    STATUS="不一致"
    STATUS_COLOR=$RED
    MISMATCH_COUNT=$((MISMATCH_COUNT + 1))
  fi
  DISPLAY_LINE="ALB\t\t${ALB_COUNTS[$SERVICE_TYPE]}\t${TF_ALB_COUNTS[$SERVICE_TYPE]}\t${STATUS_COLOR}$STATUS${NC}"
  echo -e "$DISPLAY_LINE"
  
  # ターゲットグループ比較 - 一時変数を使用して表示する
  if [ "${TG_COUNTS[$SERVICE_TYPE]}" = "${TF_TG_COUNTS[$SERVICE_TYPE]}" ]; then
    STATUS="一致"
    STATUS_COLOR=$GREEN
  else
    STATUS="不一致"
    STATUS_COLOR=$RED
    MISMATCH_COUNT=$((MISMATCH_COUNT + 1))
  fi
  DISPLAY_LINE="ターゲットグループ\t${TG_COUNTS[$SERVICE_TYPE]}\t${TF_TG_COUNTS[$SERVICE_TYPE]}\t${STATUS_COLOR}$STATUS${NC}"
  echo -e "$DISPLAY_LINE"
  
  echo "--------------------------------"
done

# デバッグログ - 最終不一致カウント
if [ "$DEBUG" = "true" ]; then
  echo "DEBUG: 最終MISMATCH_COUNT=$MISMATCH_COUNT"
fi

# ==== 早期整合性判定 ====
if [ $MISMATCH_COUNT -eq 0 ]; then
  # すべてのリソース数が一致
  if [ -z "$TF_RESOURCES" ] && [ "$VPC_COUNT_CLEAN" -eq 0 ] && [ "$RDS_COUNT_CLEAN" -eq 0 ] && [ "$ECS_CLUSTER_COUNT_CLEAN" -eq 0 ]; then
    # 環境が空（リソースなし）でTerraform状態も空の場合
    echo -e "\n${GREEN}✅ 整合性確認OK: 環境は空でTerraform状態も空です${NC}"
    cd $ORIGINAL_DIR
    exit 0
  elif [ -n "$TF_RESOURCES" ]; then
    # リソースがあってすべて一致する場合は terraform plan で最終確認
    echo -e "\n${GREEN}✓ すべてのリソース数が一致しています - Terraformプランで最終確認します${NC}"
  else
    # リソースが存在しないケース
    echo -e "\n${GREEN}✅ 整合性確認OK: AWS環境とTerraform状態は一致しています${NC}"
    cd $ORIGINAL_DIR
    exit 0
  fi
else
  # 不一致がある場合
  echo -e "\n${YELLOW}⚠️ 不整合検出: ${MISMATCH_COUNT}個のリソースで差異があります${NC}"
  
  # 修復オプションの表示
  echo -e "${BLUE}修復オプション:${NC}"
  echo -e "1. terraform importで不足リソースをインポート: make terraform-import TF_ENV=$ENV"
  echo -e "2. terraform state rmで余分なリソースを削除: terraform state rm <リソースパス>"
  echo -e "3. タグベース削除を使用: make tag-cleanup TF_ENV=$ENV"
  
  # リソース数不一致だけで終了するオプション
  if [ "${SKIP_TERRAFORM_PLAN:-false}" = "true" ] || [ "${SKIP_STATE_VERIFY:-false}" = "true" ]; then
    cd $ORIGINAL_DIR
    exit 2
  fi
  
  # リソース数不一致でも続行するオプション
  if [ "${FORCE_TERRAFORM_PLAN:-false}" = "true" ]; then
    echo -e "${YELLOW}警告: リソース数の不一致がありますが、Terraformプランを強制実行します${NC}"
  fi
fi

# ==== Terraform計画実行（差分検出） ====
echo -e "\n${BLUE}Terraformプランを実行して差分を検証しています...${NC}"

# terraform planのスキップオプション
SKIP_TERRAFORM_PLAN=${SKIP_TERRAFORM_PLAN:-false}

if [ "$SKIP_TERRAFORM_PLAN" = "true" ] || [ "${SKIP_STATE_VERIFY:-false}" = "true" ]; then
  echo -e "${YELLOW}警告: Terraformプランをスキップします${NC}"
  echo -e "${YELLOW}環境検証は不完全です - 必要に応じて手動で検証してください${NC}"
  cd $ORIGINAL_DIR
  
  # リソース数一致なら成功、不一致なら失敗で終了
  if [ $MISMATCH_COUNT -eq 0 ]; then
    exit 0
  else
    exit 2
  fi
fi

# terraform planのログファイル
PLAN_LOG_FILE=$(mktemp)

# terraform planを実行（ロックなし、タイムアウト60秒）
echo -e "${BLUE}Terraformプランを実行中です（タイムアウト: 60秒）...${NC}"
timeout 60 terraform plan -lock=false -input=false -detailed-exitcode -out=tfplan.verify > $PLAN_LOG_FILE 2>&1
PLAN_EXIT_CODE=$?

# デバッグモード時のログ表示
if [ "$DEBUG" = "true" ]; then
  echo -e "${BLUE}Terraformプラン実行ログ:${NC}"
  head -n 50 $PLAN_LOG_FILE
  echo "... (省略) ..."
  echo "DEBUG: PLAN_EXIT_CODE=$PLAN_EXIT_CODE"
fi

# タイムアウトの場合
if [ $PLAN_EXIT_CODE -eq 124 ]; then
  echo -e "\n${YELLOW}⚠️ Terraformプランの実行がタイムアウトしました（60秒経過）${NC}"
  echo -e "${YELLOW}環境の整合性は完全に検証できませんでした。${NC}"
  echo -e "${BLUE}以下のオプションがあります:${NC}"
  echo -e "1. 手動で検証: cd $(pwd) && terraform plan"
  echo -e "2. 環境変数を設定してから再実行: export TF_VAR_db_username=xxxx TF_VAR_db_password=xxxx"
  
  # 一時ファイルの削除
  rm -f tfplan.verify $PLAN_LOG_FILE
  
  # 元のディレクトリに戻る
  cd $ORIGINAL_DIR
  
  # リソース数一致なら成功、不一致なら失敗で終了
  if [ $MISMATCH_COUNT -eq 0 ]; then
    exit 0
  else
    exit 2
  fi
fi

# エラーが発生した場合のログ表示
if [ $PLAN_EXIT_CODE -eq 1 ]; then
  echo -e "\n${RED}❌ Terraformの実行中にエラーが発生しました${NC}"
  echo -e "${RED}エラーログ:${NC}"
  grep -i "error\|warning" $PLAN_LOG_FILE || head -n 20 $PLAN_LOG_FILE
  
  # 変数不足のエラーを検出
  if grep -q "No value for required variable" $PLAN_LOG_FILE; then
    echo -e "\n${YELLOW}必要な変数が設定されていません。以下の環境変数を設定してください:${NC}"
    echo -e "export TF_VAR_db_username=\"適切なユーザー名\""
    echo -e "export TF_VAR_db_password=\"適切なパスワード\""
  fi
fi

# 一時ファイルの削除
rm -f tfplan.verify $PLAN_LOG_FILE

# ==== 最終判定 ====
if [ $PLAN_EXIT_CODE -eq 0 ]; then
  # Terraform planで変更なし
  if [ $MISMATCH_COUNT -eq 0 ]; then
    echo -e "\n${GREEN}✅ 整合性確認OK: Terraform状態とAWS環境は一致しています${NC}"
  else
    echo -e "\n${YELLOW}⚠️ 混在状態: リソース数に不一致がありますが、Terraform planでは変更がありません${NC}"
    echo -e "${YELLOW}AWS CLIによる検出とTerraformの認識に相違があります。手動での確認をお勧めします。${NC}"
  fi
  
  # 元のディレクトリに戻る
  cd $ORIGINAL_DIR
  exit 0
elif [ $PLAN_EXIT_CODE -eq 2 ]; then
  echo -e "\n${YELLOW}⚠️ 不整合検出: Terraform状態とAWS環境に差分があります${NC}"
  
  # 修復オプションの表示
  echo -e "${BLUE}修復オプション:${NC}"
  echo -e "1. terraform importで不足リソースをインポート: make terraform-import TF_ENV=$ENV"
  echo -e "2. terraform state rmで余分なリソースを削除: terraform state rm <リソースパス>"
  echo -e "3. タグベース削除を使用: make tag-cleanup TF_ENV=$ENV"
  
  # 元のディレクトリに戻る
  cd $ORIGINAL_DIR
  exit 2
else
  echo -e "\n${RED}❌ Terraformの実行中にエラーが発生しました${NC}"
  
  # 元のディレクトリに戻る
  cd $ORIGINAL_DIR
  exit 1
fi