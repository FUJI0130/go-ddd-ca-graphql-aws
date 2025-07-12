# データベース関連Makefile
# makefiles/db.mk

#----------------------------------------
# データベース操作
#----------------------------------------
.PHONY: db-up db-down migrate migrate-down test-db-up test-db-down

# データベース起動/停止
db-up:
	docker compose -f deployments/docker/docker-compose.yml up -d

db-down:
	docker compose -f deployments/docker/docker-compose.yml down

# マイグレーション
migrate:
	migrate -path scripts/migrations -database "postgresql://testuser:testpass@localhost:5432/test_management?sslmode=disable" up

migrate-down:
	migrate -path scripts/migrations -database "postgresql://testuser:testpass@localhost:5432/test_management?sslmode=disable" down

test-migrate:
	migrate -path scripts/migrations -database "postgresql://test_user:test_pass@localhost:5433/test_db?sslmode=disable" up

test-migrate-down:
	migrate -path scripts/migrations -database "postgresql://test_user:test_pass@localhost:5433/test_db?sslmode=disable" down

# テスト用DB操作
test-db-up:
	docker compose -f test/integration/postgres/docker-compose.test.yml up -d
	@echo "テスト用DB(ポート5433)を起動しました"
	@./scripts/setup/wait-for-db.sh

test-db-down:
	docker compose -f test/integration/postgres/docker-compose.test.yml down
	@echo "テスト用DBを停止しました"