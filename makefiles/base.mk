# ベースMakefile - 共通変数と基本開発コマンド
# makefiles/base.mk

# 環境変数のデフォルト値設定
TF_ENV ?= development
SERVICE_TYPE ?= api
AWS_REGION ?= ap-northeast-1

# カラー設定
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# ヘルプタスク (使用方法説明用)
.PHONY: help
help:
	@echo -e "${BLUE}テストケース管理システム - Makefile使用ガイド${NC}"
	@echo ""
	@echo "=== 開発用コマンド ==="
	@echo "  run               - REST APIサーバーを起動"
	@echo "  run-graphql       - GraphQLサーバーを起動"
	@echo "  run-grpc          - gRPCサーバーを起動"
	@echo "  build             - バイナリをビルド"
	@echo "  proto             - Protocol Buffersコードを生成"
	@echo "	 tree			   - project file tree を表示"
	@echo ""
	@echo "=== テスト関連コマンド ==="
	@echo "  test              - 通常のテストを実行"
	@echo "  test-integration  - リポジトリ層の統合テストを実行"
	@echo "  test-graphql      - GraphQL統合テストを実行"
	@echo "  test-all-integration - すべての統合テストを実行"
	@echo "  test-graphql-resolver - GraphQLリゾルバーのテスト"
	@echo ""
	@echo "=== データベース操作 ==="
	@echo "  db-up             - データベースコンテナを起動"
	@echo "  db-down           - データベースコンテナを停止"
	@echo "  migrate           - マイグレーションを実行"
	@echo "  migrate-down      - マイグレーションをロールバック"
	@echo "  test-db-up        - テスト用データベースを起動"
	@echo "  test-db-down      - テスト用データベースを停止"
	@echo ""
	@echo "=== Docker関連コマンド ==="
	@echo "  docker-build-{service} - serviceのDockerイメージをビルド (service: api|graphql|grpc)"
	@echo "  docker-run-{service}   - serviceのコンテナを起動 (service: api|graphql|grpc)"
	@echo "  test-docker-{service}  - serviceのDockerテスト実行 (service: api|graphql|grpc)"
	@echo "  test-docker-all        - すべてのサービスのDockerテスト実行"
	@echo "  docker-compose-up-{service} - Docker Composeでサービスを起動"
	@echo "  docker-compose-down    - Docker Composeでコンテナを停止"
	@echo ""
	@echo "=== AWS/Terraform関連コマンド ==="
	@echo "  tf-status         - AWSリソースの状態を確認"
	@echo "  tf-init           - Terraformを初期化"
	@echo "  tf-plan           - TerraformプランをMODULE指定で作成"
	@echo "  tf-apply          - Terraformプランを適用"
	@echo "  tf-destroy        - Terraformリソースを破棄"
	@echo "  verify-ssm-params - SSMパラメータを確認・作成"
	@echo "  update-tfvars     - terraform.tfvarsを更新"
	@echo "  prepare-ecr-image - ECRイメージを準備 (SERVICE_TYPE指定)"
	@echo ""
	@echo "=== デプロイ関連コマンド ==="
	@echo "  deploy-network    - ネットワークリソースをデプロイ"
	@echo "  deploy-database   - データベースをデプロイ"
	@echo "  deploy-ecs-cluster - ECSクラスターをデプロイ"
	@echo "  deploy-api        - APIサービスをデプロイ"
	@echo "  deploy-graphql    - GraphQLサービスをデプロイ"
	@echo "  deploy-grpc       - gRPCサービスをデプロイ"
	@echo "  deploy-all-services - すべてのサービスをデプロイ"
	@echo "  verify-service    - サービスの動作を検証 (SERVICE_TYPE指定)"
	@echo "  redeploy-api      - APIサービスを再デプロイ"
	@echo "  deploy-app-workflow - アプリケーション完全デプロイを実行"
	@echo ""
	@echo "=== AWS コスト最適化 ==="
	@echo "  cost-estimate       - AWS環境のコスト見積もりを表示"
	@echo "  cleanup-minimal     - ECSサービスとロードバランサーを削除（最小限のクリーンアップ）"
	@echo "  cleanup-standard    - ECSサービス、ロードバランサー、RDSを削除"
	@echo "  cleanup-complete    - すべてのAWSリソースを削除（完全クリーンアップ）"
	@echo "  verify-and-cleanup-api - APIサービスを検証後、自動クリーンアップ"
	@echo "  verify-and-cleanup-all - 全サービスを検証後、自動クリーンアップ"
	@echo "  temporary-deploy-api   - APIサービスを一時デプロイして検証後、自動クリーンアップ"
	@echo ""
	@echo "=== インストラクション管理 ==="
	@echo "  instructions-aws      - AWS作業用インストラクションを生成"
	@echo "  instructions-backend  - バックエンド開発用インストラクションを生成"
	@echo "  instructions-frontend - フロントエンド開発用インストラクションを生成"
	@echo "  instructions-problem  - 問題解決用インストラクションを生成"
	@echo "  instructions-aws-problem - AWS問題解決用インストラクションを生成"
	@echo "  instructions-backend-problem - バックエンド問題解決用インストラクションを生成" 
	@echo "  instructions-frontend-problem - フロントエンド問題解決用インストラクションを生成"
	@echo "  instructions-custom   - カスタム組み合わせのインストラクションを生成 (MODULES指定)"
	@echo "  instructions-all     - すべての事前定義インストラクションを生成"
	@echo "  instructions-clean   - 生成されたインストラクションをクリーンアップ"
	@echo ""
	@echo "=== 環境変数 ==="
	@echo "  TF_ENV            - 環境 (development|production) デフォルト: ${TF_ENV}"
	@echo "  SERVICE_TYPE      - サービスタイプ (api|graphql|grpc) デフォルト: ${SERVICE_TYPE}"
	@echo "  MODULES           - インストラクションモジュール (instructions-custom用)"
	@echo ""
	@echo "=== 使用例 ==="
	@echo "  make run                           - APIサーバー起動"
	@echo "  make test-docker-all               - すべてのサービスのテスト実行"
	@echo "  make deploy-network TF_ENV=production - 本番環境のネットワークデプロイ"
	@echo "  make prepare-ecr-image SERVICE_TYPE=graphql - GraphQLのECRイメージ準備"
	@echo "  make instructions-aws              - AWS作業用インストラクションを生成"
	@echo "  make instructions-custom MODULES=\"aws backend\" - カスタムインストラクションを生成"
	@echo "=== Terraform状態管理 ==="
	@echo "  terraform-backup          - Terraformステートのバックアップを作成"
	@echo "  terraform-reset           - Terraformステートをリセット（バックアップ後）"
	@echo "  terraform-cleanup-minimal - Terraformでサービスとロードバランサーを削除"
	@echo "  terraform-cleanup-standard - Terraformでサービス、ロードバランサー、RDSを削除"
	@echo "  terraform-cleanup-complete - Terraformですべてのリソースを削除"
	@echo "  terraform-verify          - Terraformステートと実際のAWS環境の一致を検証"
	@echo ""
	@echo "注意: 従来の cleanup-* コマンドはAWSリソースのみを削除し、Terraformステートを更新しません。"
	@echo "      代わりに terraform-cleanup-* コマンドの使用を推奨します。"

#----------------------------------------
# 基本開発コマンド
#----------------------------------------
.PHONY: run run-graphql run-grpc build test

# アプリケーション起動
run:
	go run cmd/api/main.go

run-graphql:
	go run cmd/graphql/main.go

run-grpc:
	go run cmd/grpc/main.go

# ビルド
build:
	@echo -e "${BLUE}ビルドを開始します: ${SERVICE_TYPE}${NC}"
	@mkdir -p bin
	@if [ "${SERVICE_TYPE}" = "api" ]; then \
		go build -o bin/api cmd/api/main.go; \
		echo -e "${GREEN}APIサービスのビルドが完了しました${NC}"; \
	elif [ "${SERVICE_TYPE}" = "graphql" ]; then \
		go build -o bin/graphql cmd/graphql/main.go; \
		echo -e "${GREEN}GraphQLサービスのビルドが完了しました${NC}"; \
	elif [ "${SERVICE_TYPE}" = "grpc" ]; then \
		go build -o bin/grpc cmd/grpc/main.go; \
		echo -e "${GREEN}gRPCサービスのビルドが完了しました${NC}"; \
	else \
		echo -e "${RED}無効なサービスタイプです: ${SERVICE_TYPE}${NC}"; \
		exit 1; \
	fi

# 基本テスト
test:
	go test ./...

#----------------------------------------
# Protocol Buffers関連
#----------------------------------------
.PHONY: proto

proto:
	protoc --proto_path=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/testsuite/v1/*.proto
tree:
	tree -a -I 'build|debug|target|.git|aws|frontend|node_modules|dist|build|coverage|.next|.nuxt|.output'