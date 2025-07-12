#!/bin/bash
# ===================================================================
# ファイル名: terraform-deploy.sh
# 説明: Terraformデプロイのラッパースクリプト
# 
# 用途:
#  - Terraformコマンドのラッパーとして機能
#  - 環境に応じた設定の適用とリモートステートの管理
#  - モジュール単位での段階的なデプロイをサポート
#  - S3バケットとDynamoDBテーブルの自動作成
#  - Terraformコマンド実行前の依存関係チェック
# 
# 注意:
#  - このスクリプトはTerraform状態ファイルを更新します
#  - AWS環境の変更を伴うため、実行前に計画を確認してください
#  - DB認証情報は環境変数で設定するか対話的に入力できます
# 
# 使用方法:
#  ./terraform-deploy.sh <コマンド> <環境> [モジュール]
#
# 引数:
#  コマンド - Terraformコマンド（init, plan, apply, destroy, plan-apply, all）
#  環境    - デプロイ環境（development, production）
#  モジュール - オプション：特定のモジュール（network, database, ecs-cluster, ecs-api等）
# ===================================================================

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# コマンドライン引数の解析
COMMAND=$1
ENVIRONMENT=${2:-development}  # 環境が指定されていない場合はデフォルトでdevelopment
MODULE=$3

# 環境設定
case "${ENVIRONMENT}" in
  development)
    TF_DIR="deployments/terraform/environments/development"
    ;;
  production)
    TF_DIR="deployments/terraform/environments/production"
    ;;
  *)
    echo -e "${RED}エラー: サポートされていない環境です: ${ENVIRONMENT}${NC}"
    echo "サポートされている環境: development, production"
    exit 1
    ;;
esac

# 環境変数からバケット名とテーブル名を取得するか、デフォルト値を使用
STATE_BUCKET=${STATE_BUCKET:-"terraform-state-$(aws sts get-caller-identity --query Account --output text)-${ENVIRONMENT}"}
STATE_DYNAMODB=${STATE_DYNAMODB:-"terraform-lock-$(aws sts get-caller-identity --query Account --output text)-${ENVIRONMENT}"}
AWS_REGION=${AWS_REGION:-"ap-northeast-1"}

# モジュールターゲットの設定
MODULE_TARGET=""
if [ ! -z "${MODULE}" ]; then
  case "${MODULE}" in
    network|networking)
      MODULE_TARGET="-target=module.networking"
      ;;
    database|db)
      MODULE_TARGET="-target=module.database"
      ;;
    ecs-cluster|cluster)
      MODULE_TARGET="-target=module.shared_ecs_cluster"
      ;;
    ecs|container)
      # すべてのサービスを含む
      MODULE_TARGET="-target=module.shared_ecs_cluster -target=module.service_api -target=module.service_graphql -target=module.service_grpc"
      ;;
    ecs-api|api)
      MODULE_TARGET="-target=module.service_api"
      ;;
    ecs-graphql|graphql)
      MODULE_TARGET="-target=module.service_graphql"
      ;;
    ecs-grpc|grpc)
      MODULE_TARGET="-target=module.service_grpc.aws_ecs_service.app -target=module.service_grpc.aws_ecs_task_definition.app -target=module.service_grpc.aws_cloudwatch_log_group.app -target=module.service_grpc.aws_appautoscaling_target.app -target=module.service_grpc.aws_appautoscaling_policy.cpu -target=module.service_grpc.aws_appautoscaling_policy.memory -target=module.service_grpc.aws_iam_role.ecs_task_role"
      ;;
    ecs-grpc-native)
      MODULE_TARGET="-target=module.target_group_grpc_native -target=aws_lb_listener.grpc_https"
      ;;
    # セキュリティグループのみを対象とする新しいターゲット
    security-group-grpc)
      MODULE_TARGET="-target=module.service_grpc.aws_security_group.ecs_tasks"
      ;;

    # ECSサービスとタスク定義のみを対象とする新しいターゲット
    ecs-service-grpc)
      MODULE_TARGET="-target=module.service_grpc.aws_ecs_service.app -target=module.service_grpc.aws_ecs_task_definition.app"
      ;;

    # ロギングとスケーリングのみを対象とする新しいターゲット
    scaling-grpc)
      MODULE_TARGET="-target=module.service_grpc.aws_cloudwatch_log_group.app -target=module.service_grpc.aws_appautoscaling_target.app -target=module.service_grpc.aws_appautoscaling_policy.cpu -target=module.service_grpc.aws_appautoscaling_policy.memory"
      ;;
    # ここに追加: 新しいservice-grpc-nativeターゲットを追加する
    service-grpc-native)
      MODULE_TARGET="-target=module.service_grpc_native"
      ;;
    # 追加するコード
    secrets)
      MODULE_TARGET="-target=module.secrets"
      ;;
    https-listener)
      MODULE_TARGET="-target=aws_lb_listener.grpc_https"
      ;;
    loadbalancer|lb)
      # すべてのロードバランサーを含む
      MODULE_TARGET="-target=module.loadbalancer_api -target=module.loadbalancer_graphql -target=module.loadbalancer_grpc"
      ;;
    lb-api)
      MODULE_TARGET="-target=module.loadbalancer_api"
      ;;
    lb-graphql)
      MODULE_TARGET="-target=module.loadbalancer_graphql"
      ;;
    lb-grpc)
      MODULE_TARGET="-target=module.loadbalancer_grpc"
      ;;
    target-group-api)
      MODULE_TARGET="-target=module.target_group_api"
      ;;
    target-group-graphql)
      MODULE_TARGET="-target=module.target_group_graphql"
      ;;
    target-group-grpc)
      MODULE_TARGET="-target=module.target_group_grpc"
      ;;
    target-group-grpc-native)
      MODULE_TARGET="-target=module.target_group_grpc_native"
      ;;
    certificates)
      MODULE_TARGET="-target=module.certificates"
      ;;
    target-group-api-new)
      MODULE_TARGET="-target=module.target_group_api_new"
      ;;
    lb-api-new)
      MODULE_TARGET="-target=module.loadbalancer_api_new"
      ;;
    ecs-api-new)
      MODULE_TARGET="-target=module.service_api_new"
      ;;
    all-api-new)
      MODULE_TARGET="-target=module.service_api_new -target=module.loadbalancer_api_new -target=module.target_group_api_new"
      ;;
    target-group-grpc-new)
      MODULE_TARGET="-target=module.target_group_grpc_new"
      ;;
    target-group-grpc-native-new)
      MODULE_TARGET="-target=module.target_group_grpc_native_new"
      ;;
    lb-grpc-new)
      MODULE_TARGET="-target=module.loadbalancer_grpc_new"
      ;;
    ecs-grpc-new)
      MODULE_TARGET="-target=module.service_grpc_new"
      ;;
    https-listener-new)
      MODULE_TARGET="-target=aws_lb_listener.grpc_https_new"
      ;;
    all-grpc-new)
      MODULE_TARGET="-target=module.service_grpc_new -target=module.loadbalancer_grpc_new -target=module.target_group_grpc_new -target=module.target_group_grpc_native_new -target=aws_lb_listener.grpc_https_new"
      ;;
    target-group-graphql-new)
      MODULE_TARGET="-target=module.target_group_graphql_new"
      ;;
    lb-graphql-new)
      MODULE_TARGET="-target=module.loadbalancer_graphql_new"
      ;;
    ecs-graphql-new)
      MODULE_TARGET="-target=module.service_graphql_new"
      ;;
    all-graphql-new)
      MODULE_TARGET="-target=module.service_graphql_new -target=module.loadbalancer_graphql_new -target=module.target_group_graphql_new"
      ;;
    *)
      echo -e "${RED}エラー: サポートされていないモジュールです: ${MODULE}${NC}"
      echo "サポートされているモジュール: network, database, ecs-cluster, ecs, ecs-api, ecs-graphql, ecs-grpc, loadbalancer, lb-api, lb-graphql, lb-grpc, target-group-api, target-group-graphql, target-group-grpc"
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

# Terraformファイルのフォーマット
format_terraform() {
  echo -e "${BLUE}Terraformファイルをフォーマットしています...${NC}"
  terraform -chdir=deployments/terraform fmt -recursive
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}フォーマットが完了しました${NC}"
  else
    echo -e "${YELLOW}警告: フォーマット中に問題が発生しました${NC}"
  fi
}

# Terraformファイルの検証
validate_terraform() {
  echo -e "${BLUE}Terraformファイルを検証しています...${NC}"
  terraform -chdir=${TF_DIR} validate
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}検証が完了しました${NC}"
  else
    echo -e "${RED}エラー: Terraform構成に問題があります${NC}"
    exit 1
  fi
}

# リモートステートのセットアップ
setup_remote_state() {
  echo -e "${BLUE}リモートステートの設定を確認しています...${NC}"
  echo -e "使用するS3バケット名: ${STATE_BUCKET}"
  echo -e "使用するDynamoDBテーブル名: ${STATE_DYNAMODB}"
  
  # S3バケットの存在確認
  if aws s3api head-bucket --bucket "${STATE_BUCKET}" 2>/dev/null; then
    echo -e "${GREEN}S3バケット ${STATE_BUCKET} は既に存在します${NC}"
  else
    echo -e "${YELLOW}S3バケット ${STATE_BUCKET} を作成しています...${NC}"
    
    # バケット作成
    if ! aws s3api create-bucket \
      --bucket "${STATE_BUCKET}" \
      --region "${AWS_REGION}" \
      --create-bucket-configuration LocationConstraint="${AWS_REGION}" 2>/dev/null; then
      
      echo -e "${RED}エラー: S3バケット ${STATE_BUCKET} の作成に失敗しました${NC}"
      echo -e "${YELLOW}別のバケット名を指定するか、既存のバケットへのアクセス権を確認してください${NC}"
      echo -e "例: export STATE_BUCKET=your-unique-bucket-name"
      exit 1
    fi
    
    # バケット作成成功後の設定
    aws s3api put-bucket-versioning \
      --bucket "${STATE_BUCKET}" \
      --versioning-configuration Status=Enabled
    
    aws s3api put-bucket-encryption \
      --bucket "${STATE_BUCKET}" \
      --server-side-encryption-configuration '{"Rules": [{"ApplyServerSideEncryptionByDefault": {"SSEAlgorithm": "AES256"}}]}'
    
    echo -e "${GREEN}S3バケットが作成されました${NC}"
  fi
  
  # DynamoDBテーブルの存在確認
  if aws dynamodb describe-table --table-name "${STATE_DYNAMODB}" --region "${AWS_REGION}" &>/dev/null; then
    echo -e "${GREEN}DynamoDBテーブル ${STATE_DYNAMODB} は既に存在します${NC}"
  else
    echo -e "${YELLOW}DynamoDBテーブル ${STATE_DYNAMODB} を作成しています...${NC}"
    
    # テーブル作成
    if ! aws dynamodb create-table \
      --table-name "${STATE_DYNAMODB}" \
      --attribute-definitions AttributeName=LockID,AttributeType=S \
      --key-schema AttributeName=LockID,KeyType=HASH \
      --billing-mode PAY_PER_REQUEST \
      --region "${AWS_REGION}" > /dev/null; then
      
      echo -e "${RED}エラー: DynamoDBテーブル ${STATE_DYNAMODB} の作成に失敗しました${NC}"
      echo -e "${YELLOW}別のテーブル名を指定するか、既存のテーブルへのアクセス権を確認してください${NC}"
      echo -e "例: export STATE_DYNAMODB=your-unique-table-name"
      exit 1
    fi
    
    # テーブルが作成されるまで待機
    echo "DynamoDBテーブルの作成を待機しています..."
    aws dynamodb wait table-exists --table-name "${STATE_DYNAMODB}" --region "${AWS_REGION}"
    
    echo -e "${GREEN}DynamoDBテーブルが作成されました${NC}"
  fi
  
  # バックエンド設定の有効化
  enable_backend
}

# バックエンド設定の有効化
enable_backend() {
  echo -e "${BLUE}バックエンド設定を確認しています...${NC}"
  
  # main.tfファイルのパス
  MAIN_TF_FILE="${TF_DIR}/main.tf"
  
  # バックエンド設定がコメントアウトされているか確認
  if grep -q "# backend \"s3\"" "${MAIN_TF_FILE}"; then
    echo -e "${YELLOW}バックエンド設定が無効になっています。有効化します...${NC}"
    
    # バックエンド設定のコメントを解除
    sed -i 's/# backend "s3"/backend "s3"/g' "${MAIN_TF_FILE}"
    
    echo -e "${GREEN}バックエンド設定を有効化しました${NC}"
  else
    echo -e "${GREEN}バックエンド設定は既に有効です${NC}"
  fi
}

# Terraformの初期化
init_terraform() {
  echo -e "${BLUE}Terraformを初期化しています (環境: ${ENVIRONMENT})...${NC}"
  
  # バックエンド設定が有効かどうかに基づいて初期化
  if grep -q "backend \"s3\"" "${TF_DIR}/main.tf" && ! grep -q "# backend \"s3\"" "${TF_DIR}/main.tf"; then
    echo "リモートバックエンドで初期化します"
    terraform -chdir=${TF_DIR} init \
      -backend-config="bucket=${STATE_BUCKET}" \
      -backend-config="key=${ENVIRONMENT}/terraform.tfstate" \
      -backend-config="region=${AWS_REGION}" \
      -backend-config="dynamodb_table=${STATE_DYNAMODB}" \
      -reconfigure
  else
    echo "ローカルバックエンドで初期化します"
    terraform -chdir=${TF_DIR} init
  fi
  
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}初期化が完了しました${NC}"
  else
    echo -e "${RED}エラー: 初期化中に問題が発生しました${NC}"
    exit 1
  fi
}

# 計画と適用を連続して実行する関数
plan_apply() {
  echo -e "${BLUE}モジュール ${MODULE} の計画・適用を行います...${NC}"
  
  # モジュール固有の計画ファイル名を使用
  local plan_file="tfplan_${MODULE}"
  
  # 計画作成
  if [ ! -z "${MODULE_TARGET}" ]; then
      terraform -chdir=${TF_DIR} plan ${MODULE_TARGET} -out=${plan_file}
  else
      terraform -chdir=${TF_DIR} plan -out=${plan_file}
  fi
  
  # 計画に変更がないか確認
  if grep -q "No changes" ${TF_DIR}/terraform_plan.log 2>/dev/null; then
      echo -e "${GREEN}モジュール ${MODULE} に変更はありません. スキップします${NC}"
      return 0
  fi
  
  # 確認メッセージ
  read -p "本当にこの計画を適用しますか？ (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo -e "${YELLOW}適用はキャンセルされました${NC}"
      return 0
  fi
  
  # 適用実行
  terraform -chdir=${TF_DIR} apply ${plan_file}
  
  # 使用後に計画ファイルを削除
  rm -f ${TF_DIR}/${plan_file}
  
  if [ $? -eq 0 ]; then
      echo -e "${GREEN}モジュール ${MODULE} のデプロイが完了しました${NC}"
      return 0
  else
      echo -e "${RED}エラー: モジュール ${MODULE} のデプロイ中に問題が発生しました${NC}"
      return 1
  fi
}

# Terraformプランの表示
plan_terraform() {
  echo -e "${BLUE}デプロイ計画を表示しています (環境: ${ENVIRONMENT})...${NC}"
  
  # DB認証情報の確認
  check_db_credentials
  
  # プラン実行
  if [ ! -z "${MODULE_TARGET}" ]; then
    echo "モジュール: ${MODULE} のプランを作成します"
    terraform -chdir=${TF_DIR} plan ${MODULE_TARGET} -out=tfplan
  else
    terraform -chdir=${TF_DIR} plan -out=tfplan
  fi
  
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}計画の作成が完了しました${NC}"
  else
    echo -e "${RED}エラー: 計画の作成中に問題が発生しました${NC}"
    exit 1
  fi
}

# Terraformの適用
apply_terraform() {
  echo -e "${BLUE}インフラストラクチャをデプロイしています (環境: ${ENVIRONMENT})...${NC}"
  
  # モジュールターゲットが指定されている場合の表示
  if [ ! -z "${MODULE_TARGET}" ]; then
    echo "モジュール: ${MODULE} のみをデプロイします"
  fi
  
  # tfplanファイルの存在確認
  if [ ! -f "${TF_DIR}/tfplan" ]; then
    echo -e "${YELLOW}警告: プランファイルが見つかりません。プランを作成します...${NC}"
    plan_terraform
  fi
  
  # 確認メッセージ
  read -p "本当にAWSリソースをデプロイしますか？ (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}デプロイはキャンセルされました${NC}"
    return
  fi
  
  # 適用実行
  terraform -chdir=${TF_DIR} apply tfplan
  
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}デプロイが完了しました${NC}"
  else
    echo -e "${RED}エラー: デプロイ中に問題が発生しました${NC}"
    exit 1
  fi
}

# Terraformの破棄
destroy_terraform() {
  echo -e "${BLUE}インフラストラクチャを破棄しています (環境: ${ENVIRONMENT})...${NC}"
  
  # DB認証情報の確認
  check_db_credentials
  
  # 確認メッセージ
  echo -e "${RED}警告: これにより、${ENVIRONMENT}環境のすべてのAWSリソースが破棄されます${NC}"
  read -p "本当にAWSリソースを破棄しますか？ (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}破棄はキャンセルされました${NC}"
    return
  fi
  
  # 再確認
  read -p "最終確認: 本当に破棄しますか？ (yes/no) " CONFIRM
  if [[ "${CONFIRM}" != "yes" ]]; then
    echo -e "${YELLOW}破棄はキャンセルされました${NC}"
    return
  fi
  
  # 破棄実行
  if [ ! -z "${MODULE_TARGET}" ]; then
    echo "モジュール: ${MODULE} のみを破棄します"
    terraform -chdir=${TF_DIR} destroy ${MODULE_TARGET}
  else
    terraform -chdir=${TF_DIR} destroy
  fi
  
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}インフラストラクチャの破棄が完了しました${NC}"
  else
    echo -e "${RED}エラー: 破棄中に問題が発生しました${NC}"
    exit 1
  fi
}

# DB認証情報確認
check_db_credentials() {
  if [ -z "${TF_VAR_db_username}" ] || [ -z "${TF_VAR_db_password}" ]; then
    echo -e "${YELLOW}警告: データベース認証情報が環境変数で設定されていません${NC}"
    echo "以下のように環境変数を設定してください:"
    echo "export TF_VAR_db_username=admin"
    echo "export TF_VAR_db_password=your_secure_password"
    
    # 環境変数がない場合、対話的に設定
    read -p "データベースユーザー名を入力してください: " DB_USER
    read -sp "データベースパスワードを入力してください: " DB_PASS
    echo
    
    export TF_VAR_db_username="${DB_USER}"
    export TF_VAR_db_password="${DB_PASS}"
    
    echo -e "${GREEN}DB認証情報を設定しました${NC}"
  else
    echo -e "${GREEN}DB認証情報が設定されています${NC}"
  fi
}

# 完全なデプロイプロセスを実行
deploy_all() {
  check_dependencies
  format_terraform
  validate_terraform
  setup_remote_state
  init_terraform
  plan_terraform
  apply_terraform
}

# メインプロセス
case "$COMMAND" in
  init)
    check_dependencies
    setup_remote_state
    init_terraform
    ;;
  plan)
    check_dependencies
    format_terraform
    validate_terraform
    init_terraform
    plan_terraform
    ;;
  apply)
    check_dependencies
    apply_terraform
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
  all)
    deploy_all
    ;;
  *)
    echo -e "${BLUE}テストケース管理システム - Terraformデプロイスクリプト${NC}"
    echo
    echo "使用方法: $0 [init|plan|apply|destroy|plan-apply|all] [environment] [module]"
    echo
    echo "コマンド:"
    echo "  init       - リモートステートをセットアップし、Terraformを初期化します"
    echo "  plan       - 変更計画を作成します"
    echo "  apply      - インフラストラクチャに変更を適用します"
    echo "  destroy    - インフラストラクチャをすべて破棄します"
    echo "  plan-apply - 計画の作成と適用を連続して行います"
    echo "  all        - 完全なデプロイプロセスを実行します（format → validate → init → plan → apply）"
    echo
    echo "環境:"
    echo "  development - 開発環境 (デフォルト)"
    echo "  production  - 本番環境"
    echo
    echo "モジュール (オプション):"
    echo "  network      - ネットワークモジュールのみを対象にします"
    echo "  database     - データベースモジュールのみを対象にします"
    echo "  ecs-cluster  - 共有ECSクラスターのみを対象にします"
    echo "  ecs          - すべてのECSサービスを対象にします"
    echo "  ecs-api      - APIサービスのみを対象にします"
    echo "  ecs-graphql  - GraphQLサービスのみを対象にします"
    echo "  ecs-grpc     - gRPCサービスのみを対象にします"
    echo "  loadbalancer - すべてのロードバランサーを対象にします"
    echo "  lb-api       - APIのロードバランサーのみを対象にします"
    echo "  lb-graphql   - GraphQLのロードバランサーのみを対象にします"
    echo "  lb-grpc      - gRPCのロードバランサーのみを対象にします"
    echo "  target-group-api     - APIのターゲットグループのみを対象にします"
    echo "  target-group-graphql - GraphQLのターゲットグループのみを対象にします"
    echo "  target-group-grpc    - gRPCのターゲットグループのみを対象にします"
    echo
    echo "環境変数:"
    echo "  STATE_BUCKET   - Terraformの状態を保存するS3バケット名"
    echo "  STATE_DYNAMODB - ステートロック用のDynamoDBテーブル名"
    echo "  AWS_REGION     - AWSリージョン (デフォルト: ap-northeast-1)"
    echo
    echo "例:"
    echo "  $0 init development"
    echo "  $0 plan production"
    echo "  $0 apply development network"
    echo "  $0 plan-apply development database"
    echo "  STATE_BUCKET=my-unique-bucket $0 init development"
    exit 1
    ;;
esac

exit 0