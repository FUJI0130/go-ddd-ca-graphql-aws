# テスト関連Makefile
# makefiles/test.mk

#----------------------------------------
# テスト関連コマンド
#----------------------------------------
.PHONY: test-integration test-graphql test-all-integration test-graphql-resolver

# 統合テスト
test-integration: test-db-up
	@echo "リポジトリ層の統合テスト実行中..."
	go test ./internal/infrastructure/persistence/postgres/... -v
	make test-db-down

test-graphql: test-db-up
	@echo "GraphQL統合テスト実行中..."
	TEST_ENV=true go test -v ./internal/interface/graphql/... -tags=integration
	make test-db-down

test-all-integration: test-db-up
	@echo "すべての統合テスト実行中..."
	go test ./internal/infrastructure/persistence/postgres/... -v
	TEST_ENV=true go test -v ./internal/interface/graphql/... -tags=integration
	make test-db-down

test-graphql-resolver:
	go test ./internal/interface/graphql/resolver/... -v


# ローカル開発用ターゲット
.PHONY: local-env local-graphql-server local-setup local-test-full

# 環境変数ファイルの作成
local-env:
	@echo "Creating .env.local file..."
	@echo "export DB_HOST=localhost" > .env.local
	@echo "export DB_PORT=5433" >> .env.local
	@echo "export DB_USER=test_user" >> .env.local
	@echo "export DB_PASS=test_pass" >> .env.local
	@echo "export DB_NAME=test_db" >> .env.local
	@echo "export JWT_SECRET=test-jwt-secret-for-local-development" >> .env.local
	@echo "export PORT=8080" >> .env.local
	@echo ".env.local file created successfully"

# ローカル環境でGraphQLサーバー起動
local-graphql-server:
	@echo "Starting GraphQL server with local environment..."
	@if [ ! -f .env.local ]; then \
		echo ".env.local not found. Creating it..."; \
		$(MAKE) local-env; \
	fi
	@. $(CURDIR)/.env.local && go run cmd/graphql/main.go

# ローカル開発環境の完全セットアップ
local-setup:
	@echo "Setting up local development environment..."
	$(MAKE) local-env
	$(MAKE) test-db-up
	@sleep 5
	$(MAKE) test-migrate
	@echo "Local development environment setup complete"

# 完全ローカルテスト（DB起動→マイグレーション→サーバー起動）
local-test-full: local-setup
	@echo "Starting full local test environment..."
	@echo "GraphQL Playground will be available at: http://localhost:8080/"
	@echo "Health check at: http://localhost:8080/health"
	$(MAKE) local-graphql-server


# ======================================
# ローカルテストユーザー管理コマンド
# ======================================

# ローカル環境用テストユーザーの作成
.PHONY: local-test-users
local-test-users:
	@echo "ローカルテストユーザーを作成中..."
	@PGPASSWORD=test_pass psql -h localhost -p 5433 -U test_user -d test_db -f scripts/testdata/local-test-users.sql
	@echo "テストユーザーの作成が完了しました"

# テストユーザーの確認
.PHONY: local-check-users
local-check-users:
	@echo "現在のテストユーザー一覧:"
	@PGPASSWORD=test_pass psql -h localhost -p 5433 -U test_user -d test_db -c "\
	SELECT \
		id, \
		username, \
		role, \
		created_at, \
		updated_at \
	FROM users \
	ORDER BY created_at DESC;"

# テストユーザーの削除（クリーンアップ用）
.PHONY: local-clean-users
local-clean-users:
	@echo "テストユーザーを削除中..."
	@PGPASSWORD=test_pass psql -h localhost -p 5433 -U test_user -d test_db -c "\
	DELETE FROM users \
	WHERE username IN ('test_admin', 'test_manager', 'test_tester');"
	@echo "テストユーザーの削除が完了しました"

# ローカル環境の完全セットアップ（DB起動→マイグレーション→テストユーザー作成）
.PHONY: local-setup-complete
local-setup-complete: test-db-up test-migrate local-test-users
	@echo "========================================="
	@echo "ローカル環境のセットアップが完了しました"
	@echo "========================================="
	@echo "次のステップ:"
	@echo "1. make local-graphql-server でサーバー起動"
	@echo "2. http://localhost:8080/ でGraphQL Playground確認"
	@echo "3. ログイン情報："
	@echo "   - ユーザー名: test_admin"
	@echo "   - パスワード: password"
	@echo "========================================="

# ローカル環境の完全リセット（DB停止→起動→マイグレーション→テストユーザー作成）
.PHONY: local-reset
local-reset: test-db-down test-db-up test-migrate local-test-users
	@echo "ローカル環境の完全リセットが完了しました"