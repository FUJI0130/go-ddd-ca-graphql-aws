# 特化コマンド（開発者向け） - gqlgen実行を含む修正版
.PHONY: quick-update-graphql quick-update-api quick-update-grpc

quick-update-graphql:
	@echo -e "${BLUE}[0/5] GraphQLスキーマを更新しています...${NC}"
	@# gqlgen実行（ローカルインストール版を優先、フォールバック付き）
	@if command -v gqlgen >/dev/null 2>&1; then \
		echo -e "${GREEN}ローカルのgqlgenを使用します${NC}"; \
		gqlgen generate; \
	else \
		echo -e "${YELLOW}ローカルのgqlgenが見つかりません。go runで実行します${NC}"; \
		go mod tidy > /dev/null 2>&1 || true; \
		go run github.com/99designs/gqlgen generate; \
	fi
	@echo -e "${GREEN}✓ GraphQLスキーマ更新完了${NC}"
	@make update-image-only SERVICE_TYPE=graphql TF_ENV=development

quick-update-api:
	@make update-image-only SERVICE_TYPE=api TF_ENV=development

quick-update-grpc:
	@make update-image-only SERVICE_TYPE=grpc TF_ENV=development

# 最高速度版（検証スキップ） - gqlgen実行を含む修正版
.PHONY: fastest-update-graphql fastest-update-api fastest-update-grpc

fastest-update-graphql:
	@echo -e "${BLUE}========== 最速GraphQL更新開始 ==========${NC}"
	@START_TIME=$$(date +%s); \
	\
	echo -e "${BLUE}[1/5] GraphQLスキーマを更新しています...${NC}"; \
	if command -v gqlgen >/dev/null 2>&1; then \
		echo -e "${GREEN}ローカルのgqlgenを使用します${NC}"; \
		gqlgen generate; \
	else \
		echo -e "${YELLOW}ローカルのgqlgenが見つかりません。自動セットアップします${NC}"; \
		go mod tidy > /dev/null 2>&1 || true; \
		go run github.com/99designs/gqlgen generate; \
	fi; \
	echo -e "${GREEN}✓ GraphQLスキーマ更新完了${NC}"; \
	\
	echo -e "${BLUE}[2/5] 更新されたスキーマを確認しています...${NC}"; \
	if grep -q "deleteUser\|DeleteUser" internal/interface/graphql/generated/generated.go; then \
		echo -e "${GREEN}✓ deleteUser mutationが正常に生成されました${NC}"; \
	else \
		echo -e "${YELLOW}⚠ deleteUser mutationが見つかりません（既存の可能性）${NC}"; \
	fi; \
	\
	make update-image-only SERVICE_TYPE=graphql TF_ENV=development SKIP_VERIFY=1; \
	\
	END_TIME=$$(date +%s); \
	DURATION=$$((END_TIME - START_TIME)); \
	echo -e "${GREEN}========== 最速更新完了 (実行時間: $${DURATION}秒) ==========${NC}"

fastest-update-api:
	@make update-image-only SERVICE_TYPE=api TF_ENV=development SKIP_VERIFY=1

fastest-update-grpc:
	@make update-image-only SERVICE_TYPE=grpc TF_ENV=development SKIP_VERIFY=1

# デバッグ用：gqlgen状態確認
.PHONY: check-gqlgen-status

check-gqlgen-status:
	@echo -e "${BLUE}=== gqlgen環境状態確認 ===${NC}"
	@echo -n "ローカルgqlgen: "
	@if command -v gqlgen >/dev/null 2>&1; then \
		echo -e "${GREEN}インストール済み ($(gqlgen version))${NC}"; \
	else \
		echo -e "${RED}未インストール${NC}"; \
	fi
	@echo -n "go mod状態: "
	@if go mod verify >/dev/null 2>&1; then \
		echo -e "${GREEN}正常${NC}"; \
	else \
		echo -e "${YELLOW}要修復${NC}"; \
	fi
	@echo -n "generated.goの最終更新: "
	@if [ -f "internal/interface/graphql/generated/generated.go" ]; then \
		stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" internal/interface/graphql/generated/generated.go 2>/dev/null || \
		date -r internal/interface/graphql/generated/generated.go "+%Y-%m-%d %H:%M:%S" 2>/dev/null || \
		echo "確認できません"; \
	else \
		echo -e "${RED}ファイルが存在しません${NC}"; \
	fi
	@echo -n "deleteUser mutation: "
	@if grep -q "deleteUser\|DeleteUser" internal/interface/graphql/generated/generated.go 2>/dev/null; then \
		echo -e "${GREEN}生成済み${NC}"; \
	else \
		echo -e "${RED}未生成${NC}"; \
	fi

# トラブルシューティング用：完全リセット
.PHONY: reset-gqlgen-env

reset-gqlgen-env:
	@echo -e "${BLUE}=== gqlgen環境完全リセット ===${NC}"
	@echo -e "${YELLOW}警告: generated.goが再生成されます${NC}"
	@read -p "続行しますか？ (y/n): " CONFIRM && \
	if [ "$$CONFIRM" = "y" ] || [ "$$CONFIRM" = "Y" ]; then \
		echo -e "${BLUE}go mod tidyを実行中...${NC}"; \
		go mod tidy; \
		echo -e "${BLUE}gqlgenを最新版で再インストール中...${NC}"; \
		go install github.com/99designs/gqlgen@latest; \
		echo -e "${BLUE}generated.goを削除して再生成中...${NC}"; \
		rm -f internal/interface/graphql/generated/generated.go; \
		gqlgen generate; \
		echo -e "${GREEN}✓ 環境リセット完了${NC}"; \
		make check-gqlgen-status; \
	else \
		echo -e "${YELLOW}リセットをキャンセルしました${NC}"; \
	fi