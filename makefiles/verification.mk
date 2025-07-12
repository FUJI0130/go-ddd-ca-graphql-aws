# ===================================================================
# ファイル名: verification.mk
# 説明: AWS環境の検証用Makefile
#
# 用途:
#  - AWS環境とTerraform状態の整合性検証
#  - デプロイされたサービスのヘルスチェック検証
#  - REST API、GraphQLサービスの個別検証
#
# 注意:
#  - 検証スクリプトはAWS CLIを使用してリソースにアクセスします
#  - 環境変数TF_ENVで対象環境を指定します（デフォルト: development）
#  - サービスが完全に起動するまで検証が成功しない場合があります
#
# 主要コマンド:
#  - terraform-verify: AWS環境とTerraform状態の整合性を検証
#  - verify-api-health: REST APIのヘルスチェックを検証
#  - verify-graphql-health: GraphQLのヘルスチェックを検証
# ===================================================================

#----------------------------------------
# 検証コマンド
#----------------------------------------
.PHONY: terraform-verify verify-api-health verify-graphql-health

# AWS環境とTerraform状態の整合性を検証
terraform-verify:
	@echo -e "${BLUE}Terraform状態とAWS環境の整合性を検証しています...${NC}"
ifeq ($(USE_GO_RUN),1)
	@cd cmd/tools && go run verify-terraform.go -env $(TF_ENV) --ignore-resource-errors $(EXTRA_ARGS)
else
	@make build-terraform-verify
	@bin/verify-terraform -env $(TF_ENV) --ignore-resource-errors $(EXTRA_ARGS)
endif

# ビルドルール
build-terraform-verify:
	@mkdir -p bin
	@echo -e "${BLUE}バイナリをビルドしています...${NC}"
	@go build -o bin/verify-terraform cmd/tools/verify-terraform.go
	@echo -e "${GREEN}ビルド完了: bin/verify-terraform${NC}"

# バージョン表示
terraform-verify-version:
	@if [ -f "bin/verify-terraform" ]; then \
		bin/verify-terraform -version; \
	else \
		echo -e "${YELLOW}検証ツールがビルドされていません。ビルドを開始します...${NC}"; \
		make build-terraform-verify; \
		bin/verify-terraform -version; \
	fi

# REST APIヘルスチェック検証
verify-api-health:
	@echo -e "${BLUE}REST APIのヘルスチェックを検証しています...${NC}"
	@chmod +x scripts/verification/verify-api-health.sh
	@scripts/verification/verify-api-health.sh $(TF_ENV)

# GraphQLヘルスチェック検証
verify-graphql-health:
	@echo -e "${BLUE}GraphQLのヘルスチェックを検証しています...${NC}"
	@chmod +x scripts/verification/verify-graphql-health.sh
	@scripts/verification/verify-graphql-health.sh $(TF_ENV)

# gRPCネイティブヘルスチェック検証
verify-grpc-native-health:
	@echo -e "${BLUE}gRPCネイティブヘルスチェックを検証しています...${NC}"
	@chmod +x scripts/verification/verify-grpc-native-health.sh
	@scripts/verification/verify-grpc-native-health.sh $(TF_ENV)

# gRPCヘルスチェック検証
verify-grpc-health:
	@echo -e "${BLUE}gRPCのヘルスチェックを検証しています...${NC}"
	@chmod +x scripts/verification/verify-grpc-health.sh
	@scripts/verification/verify-grpc-health.sh $(TF_ENV)