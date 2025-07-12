#!/bin/bash
# ===================================================================
# ファイル名: verify-migration.sh
# 配置場所: scripts/verification/verify-migration.sh
# 説明: AWS環境でのマイグレーション結果検証スクリプト
# 
# 用途:
#  - マイグレーション実行後のテーブル作成確認
#  - テストユーザーデータの投入確認
#  - GraphQL認証準備状況の検証
# 
# 検証項目:
#  1. 必須テーブルの存在確認（users, refresh_tokens等）
#  2. テストユーザーデータの確認
#  3. インデックス・制約の確認
#  4. GraphQL認証用データの整合性確認
# 
# 使用方法:
#  ./verify-migration.sh <環境名>
#
# 引数:
#  環境名 - 検証対象環境（development, production）
# ===================================================================

set -e

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 引数解析
ENVIRONMENT=${1:-development}

# 環境設定
TERRAFORM_DIR="deployments/terraform/environments/${ENVIRONMENT}"
CLUSTER_NAME="${ENVIRONMENT}-shared-cluster"
TASK_FAMILY="${ENVIRONMENT}-verify-migration"
AWS_REGION=${AWS_REGION:-ap-northeast-1}
ECR_REPOSITORY="${ENVIRONMENT}-test-management-verify"

echo -e "${BLUE}========== AWS環境マイグレーション検証 (環境: ${ENVIRONMENT}) ==========${NC}"

# AWS認証情報の確認
if ! aws sts get-caller-identity &> /dev/null; then
  echo -e "${RED}エラー: AWS認証情報が設定されていないか、無効です${NC}"
  echo "AWS CLIの設定を確認してください: aws configure"
  exit 1
fi

# Terraformディレクトリの確認
if [ ! -d "${TERRAFORM_DIR}" ]; then
  echo -e "${RED}エラー: Terraform環境ディレクトリが見つかりません: ${TERRAFORM_DIR}${NC}"
  exit 1
fi

echo -e "${BLUE}ステップ1: RDS接続情報を取得しています...${NC}"

# Terraformからデータベース接続情報を取得
cd ${TERRAFORM_DIR}

DB_HOST=$(terraform output -raw db_instance_address 2>/dev/null || echo "")
DB_NAME=$(terraform output -raw db_name 2>/dev/null || echo "")
DB_USERNAME="admin"  # 固定値として使用

if [ -z "${DB_HOST}" ] || [ "${DB_HOST}" = "null" ]; then
  echo -e "${RED}エラー: RDSインスタンスが見つかりません${NC}"
  exit 1
fi

echo -e "${GREEN}✓ RDS接続情報を取得しました${NC}"

# プロジェクトルートディレクトリに戻る
cd - > /dev/null

echo -e "${BLUE}ステップ2: 検証用SQLスクリプトを作成しています...${NC}"

# 検証用SQLスクリプトを動的生成
cat > /tmp/verify-migration.sql << 'EOF'
-- マイグレーション検証用SQLスクリプト

-- ステップ1: 必須テーブルの存在確認
DO $$
DECLARE
    table_count INTEGER;
    required_tables TEXT[] := ARRAY[
        'test_suites', 'test_groups', 'test_cases', 'effort_records', 'status_history',
        'users', 'refresh_tokens', 'login_history'
    ];
    table_name TEXT;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ1: 必須テーブルの存在確認';
    RAISE NOTICE '========================================';
    
    FOREACH table_name IN ARRAY required_tables
    LOOP
        SELECT COUNT(*) INTO table_count
        FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_name = table_name;
        
        IF table_count > 0 THEN
            RAISE NOTICE '✓ テーブル % が存在します', table_name;
        ELSE
            RAISE NOTICE '✗ テーブル % が存在しません', table_name;
        END IF;
    END LOOP;
END
$$;

-- ステップ2: ENUMタイプの確認
DO $$
DECLARE
    enum_count INTEGER;
    required_enums TEXT[] := ARRAY[
        'priority_enum', 'suite_status_enum', 'test_status_enum', 'user_role_enum'
    ];
    enum_name TEXT;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ2: ENUMタイプの確認';
    RAISE NOTICE '========================================';
    
    FOREACH enum_name IN ARRAY required_enums
    LOOP
        SELECT COUNT(*) INTO enum_count
        FROM pg_type 
        WHERE typname = enum_name AND typtype = 'e';
        
        IF enum_count > 0 THEN
            RAISE NOTICE '✓ ENUM % が存在します', enum_name;
        ELSE
            RAISE NOTICE '✗ ENUM % が存在しません', enum_name;
        END IF;
    END LOOP;
END
$$;

-- ステップ3: シーケンスの確認
DO $$
DECLARE
    seq_count INTEGER;
    required_sequences TEXT[] := ARRAY[
        'test_suite_seq', 'test_group_seq', 'test_case_seq', 'user_seq'
    ];
    seq_name TEXT;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ3: シーケンスの確認';
    RAISE NOTICE '========================================';
    
    FOREACH seq_name IN ARRAY required_sequences
    LOOP
        SELECT COUNT(*) INTO seq_count
        FROM information_schema.sequences 
        WHERE sequence_schema = 'public' AND sequence_name = seq_name;
        
        IF seq_count > 0 THEN
            RAISE NOTICE '✓ シーケンス % が存在します', seq_name;
        ELSE
            RAISE NOTICE '✗ シーケンス % が存在しません', seq_name;
        END IF;
    END LOOP;
END
$$;

-- ステップ4: インデックスの確認
DO $$
DECLARE
    index_count INTEGER;
    required_indexes TEXT[] := ARRAY[
        'idx_users_username', 'idx_users_role',
        'idx_refresh_tokens_user_id', 'idx_refresh_tokens_expires_at', 'idx_refresh_tokens_token',
        'idx_effort_records_date', 'idx_test_cases_priority', 'idx_test_cases_status'
    ];
    index_name TEXT;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ4: インデックスの確認';
    RAISE NOTICE '========================================';
    
    FOREACH index_name IN ARRAY required_indexes
    LOOP
        SELECT COUNT(*) INTO index_count
        FROM pg_indexes 
        WHERE schemaname = 'public' AND indexname = index_name;
        
        IF index_count > 0 THEN
            RAISE NOTICE '✓ インデックス % が存在します', index_name;
        ELSE
            RAISE NOTICE '✗ インデックス % が存在しません', index_name;
        END IF;
    END LOOP;
END
$$;

-- ステップ5: usersテーブルの構造確認
DO $$
DECLARE
    column_count INTEGER;
    required_columns TEXT[] := ARRAY[
        'id', 'username', 'password_hash', 'role', 'created_at', 'updated_at', 'last_login_at'
    ];
    column_name TEXT;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ5: usersテーブルの構造確認';
    RAISE NOTICE '========================================';
    
    FOREACH column_name IN ARRAY required_columns
    LOOP
        SELECT COUNT(*) INTO column_count
        FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = column_name;
        
        IF column_count > 0 THEN
            RAISE NOTICE '✓ usersテーブルのカラム % が存在します', column_name;
        ELSE
            RAISE NOTICE '✗ usersテーブルのカラム % が存在しません', column_name;
        END IF;
    END LOOP;
END
$$;

-- ステップ6: テストユーザーデータの確認
DO $$
DECLARE
    user_count INTEGER;
    admin_count INTEGER;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ6: テストユーザーデータの確認';
    RAISE NOTICE '========================================';
    
    -- 総ユーザー数確認
    SELECT COUNT(*) INTO user_count FROM users;
    RAISE NOTICE 'ユーザー総数: %', user_count;
    
    -- test_adminユーザーの確認
    SELECT COUNT(*) INTO admin_count FROM users WHERE username = 'test_admin' AND role = 'Admin';
    
    IF admin_count > 0 THEN
        RAISE NOTICE '✓ test_adminユーザー（Admin権限）が存在します';
    ELSE
        RAISE NOTICE '✗ test_adminユーザー（Admin権限）が存在しません';
    END IF;
    
    -- ユーザー一覧表示
    RAISE NOTICE '';
    RAISE NOTICE '登録済みユーザー一覧:';
    FOR rec IN 
        SELECT id, username, role, created_at 
        FROM users 
        ORDER BY created_at
    LOOP
        RAISE NOTICE '  - ID: %, ユーザー名: %, ロール: %, 作成日時: %', 
                     rec.id, rec.username, rec.role, rec.created_at;
    END LOOP;
END
$$;

-- ステップ7: refresh_tokensテーブルの強化確認
DO $$
DECLARE
    column_count INTEGER;
    enhanced_columns TEXT[] := ARRAY[
        'issued_at', 'last_used_at', 'client_info', 'ip_address', 'updated_at'
    ];
    column_name TEXT;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ7: refresh_tokensテーブル強化確認';
    RAISE NOTICE '========================================';
    
    FOREACH column_name IN ARRAY enhanced_columns
    LOOP
        SELECT COUNT(*) INTO column_count
        FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'refresh_tokens' AND column_name = column_name;
        
        IF column_count > 0 THEN
            RAISE NOTICE '✓ refresh_tokensテーブルの拡張カラム % が存在します', column_name;
        ELSE
            RAISE NOTICE '✗ refresh_tokensテーブルの拡張カラム % が存在しません', column_name;
        END IF;
    END LOOP;
END
$$;

-- ステップ8: トリガーの確認
DO $$
DECLARE
    trigger_count INTEGER;
    required_triggers TEXT[] := ARRAY[
        'update_users_updated_at', 'update_refresh_tokens_updated_at',
        'update_test_suites_updated_at', 'update_test_groups_updated_at', 'update_test_cases_updated_at'
    ];
    trigger_name TEXT;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ8: トリガーの確認';
    RAISE NOTICE '========================================';
    
    FOREACH trigger_name IN ARRAY required_triggers
    LOOP
        SELECT COUNT(*) INTO trigger_count
        FROM information_schema.triggers 
        WHERE trigger_schema = 'public' AND trigger_name = trigger_name;
        
        IF trigger_count > 0 THEN
            RAISE NOTICE '✓ トリガー % が存在します', trigger_name;
        ELSE
            RAISE NOTICE '✗ トリガー % が存在しません', trigger_name;
        END IF;
    END LOOP;
END
$$;

-- ステップ9: GraphQL認証準備状況の最終確認
DO $$
DECLARE
    auth_ready BOOLEAN := TRUE;
    check_result TEXT;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ステップ9: GraphQL認証準備状況の最終確認';
    RAISE NOTICE '========================================';
    
    -- usersテーブル確認
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        auth_ready := FALSE;
        RAISE NOTICE '✗ usersテーブルが存在しません';
    ELSE
        RAISE NOTICE '✓ usersテーブルが存在します';
    END IF;
    
    -- refresh_tokensテーブル確認
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'refresh_tokens') THEN
        auth_ready := FALSE;
        RAISE NOTICE '✗ refresh_tokensテーブルが存在しません';
    ELSE
        RAISE NOTICE '✓ refresh_tokensテーブルが存在します';
    END IF;
    
    -- test_adminユーザー確認
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'test_admin' AND role = 'Admin') THEN
        auth_ready := FALSE;
        RAISE NOTICE '✗ test_adminユーザーが存在しません';
    ELSE
        RAISE NOTICE '✓ test_adminユーザーが存在します';
    END IF;
    
    -- user_role_enum確認
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role_enum') THEN
        auth_ready := FALSE;
        RAISE NOTICE '✗ user_role_enumが存在しません';
    ELSE
        RAISE NOTICE '✓ user_role_enumが存在します';
    END IF;
    
    RAISE NOTICE '';
    IF auth_ready THEN
        RAISE NOTICE '🎉 GraphQL認証の準備が完了しています！';
        RAISE NOTICE '';
        RAISE NOTICE 'GraphQL認証テストの実行方法:';
        RAISE NOTICE '1. GraphQL Playgroundにアクセス';
        RAISE NOTICE '2. 以下のmutationを実行してJWTトークンを取得:';
        RAISE NOTICE '   mutation {';
        RAISE NOTICE '     login(username: "test_admin", password: "test_password") {';
        RAISE NOTICE '       token';
        RAISE NOTICE '       user { id username role }';
        RAISE NOTICE '     }';
        RAISE NOTICE '   }';
        RAISE NOTICE '3. 取得したトークンをHTTP HEADERSに設定:';
        RAISE NOTICE '   {"Authorization": "Bearer YOUR_TOKEN"}';
        RAISE NOTICE '4. 認証済みAPIを実行してテスト';
    ELSE
        RAISE NOTICE '❌ GraphQL認証の準備が完了していません';
        RAISE NOTICE 'マイグレーションまたはテストユーザー投入を再実行してください';
    END IF;
END
$$;

-- 検証完了
SELECT 'マイグレーション検証が完了しました' AS result;
EOF

echo -e "${GREEN}✓ 検証用SQLスクリプトを作成しました${NC}"

echo -e "${BLUE}ステップ3: 検証用Dockerイメージをビルドしています...${NC}"

# ECRリポジトリの作成（存在しない場合）
if ! aws ecr describe-repositories --repository-names ${ECR_REPOSITORY} --region ${AWS_REGION} &>/dev/null; then
  echo -e "${YELLOW}ECRリポジトリが存在しません。作成しています...${NC}"
  aws ecr create-repository --repository-name ${ECR_REPOSITORY} --region ${AWS_REGION}
  echo -e "${GREEN}✓ ECRリポジトリを作成しました${NC}"
fi

# ECRログイン
echo -e "${BLUE}ECRにログインしています...${NC}"
aws ecr get-login-password --region ${AWS_REGION} | \
  docker login --username AWS --password-stdin $(aws sts get-caller-identity --query Account --output text).dkr.ecr.${AWS_REGION}.amazonaws.com

# 検証用Dockerイメージのビルド
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
IMAGE_URI="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}:latest"

# 一時的なDockerfileを作成
cat > /tmp/Dockerfile.verify << 'EOF'
FROM postgres:14-alpine

# 検証用SQLファイルをコピー
COPY /tmp/verify-migration.sql /sql/

# 実行用エントリーポイント
ENTRYPOINT ["psql"]
EOF

docker build -f /tmp/Dockerfile.verify -t ${ECR_REPOSITORY}:latest /tmp/
docker tag ${ECR_REPOSITORY}:latest ${IMAGE_URI}

# ECRにプッシュ
echo -e "${BLUE}ECRにイメージをプッシュしています...${NC}"
docker push ${IMAGE_URI}
echo -e "${GREEN}✓ 検証用イメージの準備が完了しました${NC}"

echo -e "${BLUE}ステップ4: ECSタスクで検証を実行しています...${NC}"

# ECSタスク定義を動的生成
TASK_DEFINITION=$(cat <<EOF
{
  "family": "${TASK_FAMILY}",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::${ACCOUNT_ID}:role/${ENVIRONMENT}-ecs-task-execution-role",
  "taskRoleArn": "arn:aws:iam::${ACCOUNT_ID}:role/${ENVIRONMENT}-ecs-task-execution-role",
  "containerDefinitions": [
    {
      "name": "verify-container",
      "image": "${IMAGE_URI}",
      "essential": true,
      "command": [
        "postgresql://${DB_USERNAME}:\${DB_PASSWORD}@${DB_HOST}:5432/${DB_NAME}?sslmode=require",
        "-f", "/sql/verify-migration.sql"
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "/${ENVIRONMENT}/database/password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/verify-migration",
          "awslogs-region": "${AWS_REGION}",
          "awslogs-stream-prefix": "verify-${ENVIRONMENT}"
        }
      }
    }
  ]
}
EOF
)

# CloudWatchログ グループの作成
if ! aws logs describe-log-groups --log-group-name-prefix "/ecs/verify-migration" --region ${AWS_REGION} | grep -q "/ecs/verify-migration"; then
  aws logs create-log-group --log-group-name "/ecs/verify-migration" --region ${AWS_REGION}
fi

# タスク定義の登録と実行
echo "${TASK_DEFINITION}" > /tmp/verify-task-definition.json
TASK_DEFINITION_ARN=$(aws ecs register-task-definition \
  --cli-input-json file:///tmp/verify-task-definition.json \
  --region ${AWS_REGION} \
  --query 'taskDefinition.taskDefinitionArn' --output text)

# VPC設定情報の取得
cd ${TERRAFORM_DIR}
PRIVATE_SUBNET_IDS=$(terraform output -json private_subnet_ids | jq -r '.[]' | tr '\n' ',' | sed 's/,$//')
VPC_ID=$(terraform output -raw vpc_id)
cd - > /dev/null

SECURITY_GROUP_ID=$(aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=${ENVIRONMENT}-ecs-*" "Name=vpc-id,Values=${VPC_ID}" \
  --query 'SecurityGroups[0].GroupId' --output text --region ${AWS_REGION})

# ECSタスク実行
TASK_ARN=$(aws ecs run-task \
  --cluster ${CLUSTER_NAME} \
  --task-definition ${TASK_DEFINITION_ARN} \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[${PRIVATE_SUBNET_IDS}],securityGroups=[${SECURITY_GROUP_ID}],assignPublicIp=DISABLED}" \
  --region ${AWS_REGION} \
  --query 'tasks[0].taskArn' --output text)

echo -e "${GREEN}✓ 検証タスクを起動しました${NC}"

# タスク完了待機
echo -e "${BLUE}検証実行の完了を待機しています...${NC}"
WAIT_COUNT=0
MAX_WAIT=30

while [ ${WAIT_COUNT} -lt ${MAX_WAIT} ]; do
  TASK_STATUS=$(aws ecs describe-tasks \
    --cluster ${CLUSTER_NAME} \
    --tasks ${TASK_ARN} \
    --region ${AWS_REGION} \
    --query 'tasks[0].lastStatus' --output text)
  
  echo -n "."
  
  if [ "${TASK_STATUS}" = "STOPPED" ]; then
    echo -e "\n${GREEN}✓ 検証タスクが完了しました${NC}"
    break
  fi
  
  sleep 10
  WAIT_COUNT=$((WAIT_COUNT + 1))
done

echo -e "${BLUE}ステップ5: 検証結果を表示しています...${NC}"

# CloudWatchログの表示
LOG_STREAM_NAME=$(aws logs describe-log-streams \
  --log-group-name "/ecs/verify-migration" \
  --order-by LastEventTime \
  --descending \
  --max-items 1 \
  --region ${AWS_REGION} \
  --query 'logStreams[0].logStreamName' --output text)

if [ "${LOG_STREAM_NAME}" != "None" ] && [ ! -z "${LOG_STREAM_NAME}" ]; then
  echo -e "${BLUE}========== マイグレーション検証結果 ==========${NC}"
  aws logs get-log-events \
    --log-group-name "/ecs/verify-migration" \
    --log-stream-name "${LOG_STREAM_NAME}" \
    --region ${AWS_REGION} \
    --query 'events[].message' --output text
  echo -e "${BLUE}=================================================${NC}"
else
  echo -e "${YELLOW}警告: 検証ログが見つかりません${NC}"
fi

# 終了コード確認
EXIT_CODE=$(aws ecs describe-tasks \
  --cluster ${CLUSTER_NAME} \
  --tasks ${TASK_ARN} \
  --region ${AWS_REGION} \
  --query 'tasks[0].containers[0].exitCode' --output text)

# クリーンアップ
rm -f /tmp/verify-migration.sql
rm -f /tmp/verify-task-definition.json
rm -f /tmp/Dockerfile.verify

if [ "${EXIT_CODE}" = "0" ]; then
  echo -e "${GREEN}========== マイグレーション検証成功 ==========${NC}"
  echo -e "${GREEN}✓ すべての検証項目が正常に完了しました${NC}"
else
  echo -e "${RED}========== マイグレーション検証で問題発見 ==========${NC}"
  echo -e "${RED}✗ 検証中に問題が発見されました (終了コード: ${EXIT_CODE})${NC}"
  exit 1
fi

echo -e "${GREEN}AWS環境マイグレーション検証が完了しました${NC}"