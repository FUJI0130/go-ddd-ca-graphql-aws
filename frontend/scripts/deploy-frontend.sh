#!/bin/bash
# frontend/scripts/deploy-frontend.sh
# フロントエンドAWS環境デプロイスクリプト（terraform-deploy.shパターン踏襲）

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# コマンドライン引数の解析
COMMAND=$1
ENVIRONMENT=${2:-development}
MODULE=$3

# フロントエンド用の環境設定
case "${ENVIRONMENT}" in
  development)
    TF_DIR="deployments/terraform/environments/development"
    ;;
  staging)
    TF_DIR="deployments/terraform/environments/staging"
    ;;
  production)
    TF_DIR="deployments/terraform/environments/production"
    ;;
  *)
    echo -e "${RED}エラー: サポートされていない環境です: ${ENVIRONMENT}${NC}"
    echo "サポートされている環境: development, staging, production"
    exit 1
    ;;
esac

# 環境変数からバケット名とテーブル名を取得するか、デフォルト値を使用
STATE_BUCKET=${STATE_BUCKET:-"terraform-state-testmgmt"}
STATE_DYNAMODB=${STATE_DYNAMODB:-"terraform-state-lock"}
AWS_REGION=${AWS_REGION:-"ap-northeast-1"}

# モジュールターゲットの設定
MODULE_TARGET=""
if [ ! -z "${MODULE}" ]; then
  case "${MODULE}" in
    s3-hosting|s3)
      MODULE_TARGET="-target=module.s3_hosting"
      ;;
    cloudfront|cf)
      MODULE_TARGET="-target=module.cloudfront"
      ;;
    all|frontend)
      # すべてのフロントエンドモジュールを含む
      MODULE_TARGET="-target=module.s3_hosting -target=module.cloudfront"
      ;;
    *)
      echo -e "${RED}エラー: サポートされていないモジュールです: ${MODULE}${NC}"
      echo "サポートされているモジュール: s3-hosting, cloudfront, all"
      exit 1
      ;;
  esac
fi

# 必要なツールの確認
check_dependencies() {
  echo -e "${BLUE}依存関係を確認しています...${NC}"
  
  if ! command -v terraform &> /dev/null; then
    echo -e "${RED}エラー: Terraformがインストールされていません${NC}"
    exit 1
  fi
  
  if ! command -v aws &> /dev/null; then
    echo -e "${RED}エラー: AWS CLIがインストールされていません${NC}"
    exit 1
  fi
  
  # AWS認証情報の確認
  if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}エラー: AWS認証情報が設定されていないか、無効です${NC}"
    echo "AWS CLIの設定を確認してください: aws configure"
    exit 1
  fi
  
  echo -e "${GREEN}依存関係の確認が完了しました${NC}"
}

# Terraformの初期化
init_terraform() {
  echo -e "${BLUE}Terraformを初期化しています (環境: ${ENVIRONMENT})...${NC}"
  
  cd ${TF_DIR}
  terraform init \
    -backend-config="bucket=${STATE_BUCKET}" \
    -backend-config="key=frontend/${ENVIRONMENT}/terraform.tfstate" \
    -backend-config="region=${AWS_REGION}" \
    -backend-config="dynamodb_table=${STATE_DYNAMODB}" \
    -reconfigure
  
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}初期化が完了しました${NC}"
  else
    echo -e "${RED}エラー: 初期化中に問題が発生しました${NC}"
    exit 1
  fi
  
  cd - > /dev/null
}

# 計画と適用を連続して実行する関数
plan_apply() {
  echo -e "${BLUE}フロントエンドモジュール ${MODULE} の計画・適用を行います...${NC}"
  
  cd ${TF_DIR}
  
  # モジュール固有の計画ファイル名を使用
  local plan_file="tfplan_frontend_${MODULE}"
  
  # 計画作成
  echo -e "${BLUE}Terraformプランを作成しています...${NC}"
  if [ ! -z "${MODULE_TARGET}" ]; then
      terraform plan ${MODULE_TARGET} -out=${plan_file}
  else
      terraform plan -out=${plan_file}
  fi
  
  # 確認メッセージ
  echo -e "${YELLOW}以下のフロントエンドリソースが変更されます:${NC}"
  if [ ! -z "${MODULE_TARGET}" ]; then
      terraform show ${plan_file} | grep -E "(will be created|will be updated|will be destroyed)" || echo "変更はありません"
  fi
  
  read -p "本当にこの計画を適用しますか？ (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo -e "${YELLOW}適用はキャンセルされました${NC}"
      rm -f ${plan_file}
      cd - > /dev/null
      return 0
  fi
  
  # 適用実行
  echo -e "${BLUE}Terraformプランを適用しています...${NC}"
  terraform apply ${plan_file}
  
  # 使用後に計画ファイルを削除
  rm -f ${plan_file}
  
  if [ $? -eq 0 ]; then
      echo -e "${GREEN}フロントエンドモジュール ${MODULE} のデプロイが完了しました${NC}"
      cd - > /dev/null
      return 0
  else
      echo -e "${RED}エラー: フロントエンドモジュール ${MODULE} のデプロイ中に問題が発生しました${NC}"
      cd - > /dev/null
      return 1
  fi
}

# Terraformの破棄
destroy_terraform() {
  echo -e "${BLUE}フロントエンドインフラストラクチャを破棄しています (環境: ${ENVIRONMENT})...${NC}"
  
  cd ${TF_DIR}
  
  # 確認メッセージ
  echo -e "${RED}警告: これにより、${ENVIRONMENT}環境のフロントエンドAWSリソースが破棄されます${NC}"
  read -p "本当にフロントエンドAWSリソースを破棄しますか？ (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}破棄はキャンセルされました${NC}"
    cd - > /dev/null
    return
  fi
  
  # 再確認
  read -p "最終確認: 本当に破棄しますか？ (yes/no) " CONFIRM
  if [[ "${CONFIRM}" != "yes" ]]; then
    echo -e "${YELLOW}破棄はキャンセルされました${NC}"
    cd - > /dev/null
    return
  fi
  
  # 破棄実行
  if [ ! -z "${MODULE_TARGET}" ]; then
    echo "モジュール: ${MODULE} のみを破棄します"
    terraform destroy ${MODULE_TARGET} -auto-approve
  else
    terraform destroy -auto-approve
  fi
  
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}フロントエンドインフラストラクチャの破棄が完了しました${NC}"
  else
    echo -e "${RED}エラー: 破棄中に問題が発生しました${NC}"
    exit 1
  fi
  
  cd - > /dev/null
}

# メインプロセス
case "$COMMAND" in
  init)
    check_dependencies
    init_terraform
    ;;
  plan)
    check_dependencies
    init_terraform
    cd ${TF_DIR}
    if [ ! -z "${MODULE_TARGET}" ]; then
      terraform plan ${MODULE_TARGET}
    else
      terraform plan
    fi
    cd - > /dev/null
    ;;
  apply)
    check_dependencies
    cd ${TF_DIR}
    if [ -f "tfplan_frontend_${MODULE}" ]; then
      terraform apply tfplan_frontend_${MODULE}
    else
      echo -e "${YELLOW}警告: プランファイルが見つかりません。プランを作成します...${NC}"
      terraform plan -out=tfplan_frontend_${MODULE}
      terraform apply tfplan_frontend_${MODULE}
    fi
    cd - > /dev/null
    ;;
  destroy)
    check_dependencies
    init_terraform
    destroy_terraform
    ;;
  plan-apply)
    check_dependencies
    init_terraform
    plan_apply
    ;;
  *)
    echo -e "${BLUE}フロントエンドAWS環境 - Terraformデプロイスクリプト${NC}"
    echo
    echo "使用方法: $0 [init|plan|apply|destroy|plan-apply] [environment] [module]"
    echo
    echo "コマンド:"
    echo "  init       - Terraformを初期化します"
    echo "  plan       - 変更計画を作成します"
    echo "  apply      - インフラストラクチャに変更を適用します"
    echo "  destroy    - インフラストラクチャをすべて破棄します"
    echo "  plan-apply - 計画の作成と適用を連続して行います"
    echo
    echo "環境:"
    echo "  development - 開発環境 (デフォルト)"
    echo "  staging     - ステージング環境"
    echo "  production  - 本番環境"
    echo
    echo "モジュール (オプション):"
    echo "  s3-hosting  - S3ホスティングモジュールのみを対象にします"
    echo "  cloudfront  - CloudFrontモジュールのみを対象にします"
    echo "  all         - すべてのフロントエンドモジュールを対象にします"
    echo
    echo "例:"
    echo "  $0 init development"
    echo "  $0 plan-apply development s3-hosting"
    echo "  $0 plan-apply development all"
    echo "  $0 destroy development"
    exit 1
    ;;
esac

exit 0