# ===================================================================
# ファイル名: terraform-workflow.mk
# 説明: Terraformを使用した開発ライフサイクル管理のMakefile
#
# 用途:
#  - REST API開発環境の段階的なデプロイと削除
#  - 開発環境のライフサイクル管理（開始、一時停止、再開、完全停止）
#  - terraform-deploy.shを使用したデプロイプロセスの実行
#  - デプロイ後の自動検証（verify-api-health）
#
# 注意:
#  - このMakefileはterraform-deploy.shを使用してTerraformコマンドを実行します
#  - 環境変数TF_ENVで対象環境を指定します（デフォルト: development）
#  - スクリプト実行にはAWS認証情報が正しく設定されている必要があります
#
# 主要コマンド:
#  - terraform-start-api-dev: API開発環境のデプロイ
#  - terraform-pause-api-dev: API開発環境の一時停止
#  - terraform-resume-api-dev: API開発環境の再開
#  - terraform-stop-api-dev: API開発環境の完全停止
# ===================================================================

# makefiles/terraform-workflow.mk に追加
.PHONY: deploy-api-dev terraform-cleanup-iam

terraform-cleanup-iam:
	@echo -e "${BLUE}IAMリソースをクリーンアップしています...${NC}"
	@chmod +x scripts/terraform/aws-iam-cleaner.sh
	@scripts/terraform/aws-iam-cleaner.sh $(TF_ENV)
	@echo -e "${GREEN}IAMリソースのクリーンアップが完了しました${NC}"

# APIアプリケーションとインフラを一括デプロイ
deploy-api-dev:
	@echo -e "${BLUE}APIアプリケーションと環境を準備しています...${NC}"
	@if [ "$(SKIP_IMAGE)" != "1" ]; then \
		make prepare-ecr-image SERVICE_TYPE=api TF_ENV=$(TF_ENV); \
		make update-tfvars SERVICE_TYPE=api TF_ENV=$(TF_ENV); \
	fi
	@make terraform-start-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}APIアプリケーションと環境のデプロイが完了しました${NC}"

# GraphQLアプリケーションとインフラを一括デプロイ
deploy-graphql-dev:
	@echo -e "${BLUE}GraphQLアプリケーションと環境を準備しています...${NC}"
	@if [ "$(SKIP_IMAGE)" != "1" ]; then \
		make prepare-ecr-image SERVICE_TYPE=graphql TF_ENV=$(TF_ENV); \
		make update-tfvars SERVICE_TYPE=graphql TF_ENV=$(TF_ENV); \
	fi
	@make terraform-start-graphql-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}GraphQLアプリケーションと環境のデプロイが完了しました${NC}"

# REST API専用の開発ライフサイクル管理
.PHONY: terraform-start-api-dev terraform-pause-api-dev terraform-resume-api-dev terraform-stop-api-dev terraform-test-api-dev

# API開発環境のデプロイ（コア環境＋APIサービス）
# 修正後のコード
terraform-start-api-dev:
	@echo -e "${BLUE}API開発環境を準備しています (Terraformベース)...${NC}"
	@chmod +x scripts/terraform/terraform-deploy.sh
 
	@echo -e "${BLUE}デプロイ前の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ前に検証問題があります${NC}"
 
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) network
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) database
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-cluster
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) secrets
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) target-group-api
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) lb-api
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-api
 
	@echo -e "${GREEN}API開発環境が準備されました${NC}"
	@echo -e "${BLUE}デプロイ後の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ後に検証問題があります${NC}"
	@make verify-api-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに失敗しましたが、サービスは起動している可能性があります${NC}"

# GraphQL開発環境のデプロイ
terraform-start-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を準備しています (Terraformベース)...${NC}"
	@chmod +x scripts/terraform/terraform-deploy.sh
 
	@echo -e "${BLUE}デプロイ前の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ前に検証問題があります${NC}"
 
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) network
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) database
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-cluster
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) secrets
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) target-group-graphql
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) lb-graphql
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-graphql
 
	@echo -e "${GREEN}GraphQL開発環境が準備されました${NC}"
	@echo -e "${BLUE}デプロイ後の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ後に検証問題があります${NC}"
	@make verify-graphql-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに失敗しましたが、サービスは起動している可能性があります${NC}"

# API開発環境の一時停止
terraform-pause-api-dev:
	@echo -e "${BLUE}API開発環境を一時停止しています (Terraformベース)...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV) ecs-api
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV) lb-api
	@echo -e "${GREEN}API開発環境は一時停止されました。コア基盤は維持されています。${NC}"

# GraphQL開発環境の一時停止
terraform-pause-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を一時停止しています (Terraformベース)...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV) ecs-graphql
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV) lb-graphql
	@echo -e "${GREEN}GraphQL開発環境は一時停止されました。コア基盤は維持されています。${NC}"

# API開発環境の再開
terraform-resume-api-dev:
	@echo -e "${BLUE}API開発環境を再開しています (Terraformベース)...${NC}"
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) target-group-api
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) lb-api
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-api
	@echo -e "${GREEN}API開発環境が再開されました${NC}"
	@make verify-api-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに失敗しましたが、サービスは起動している可能性があります${NC}"

# GraphQL開発環境の再開
terraform-resume-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を再開しています (Terraformベース)...${NC}"
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) target-group-graphql
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) lb-graphql
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-graphql
	@echo -e "${GREEN}GraphQL開発環境が再開されました${NC}"
	@make verify-graphql-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに失敗しましたが、サービスは起動している可能性があります${NC}"

# 通常のクリーンアップ（スクリプト実行はコメントアウト）
terraform-stop-api-dev:
	@echo -e "${BLUE}API開発環境を完全に停止しています (Terraformベース)...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@make terraform-cleanup-iam TF_ENV=$(TF_ENV)
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV)
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: 一部のリソースが残っている可能性があります${NC}"
# 残存リソースの確認
	@make check-resources TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}API開発環境は停止されました${NC}"

# GraphQL開発環境の完全停止
terraform-stop-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を完全に停止しています (Terraformベース)...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@make terraform-cleanup-iam TF_ENV=$(TF_ENV)
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV) ecs-graphql
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV) lb-graphql
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV) target-group-graphql
	@echo -e "${GREEN}GraphQL開発環境は停止されました${NC}"
  
# 完全クリーンアップ（スクリプト実行を明示的に選択）
terraform-full-cleanup:
	@echo -e "${RED}警告: すべてのリソースを完全にクリーンアップします！${NC}"
	@make terraform-stop-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}残存リソースを確認しています...${NC}"
	@make check-resources TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}タグベースのクリーンアップを実行します...${NC}"
	@chmod +x scripts/terraform/aws-resource-cleaner.sh
	@scripts/terraform/aws-resource-cleaner.sh $(TF_ENV)
	@echo -e "${BLUE}クリーンアップ後の検証を実行しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV)
	@make check-resources TF_ENV=$(TF_ENV)

# API開発環境のクイックテスト
terraform-test-api-dev:
	@echo -e "${BLUE}API一時テスト環境をデプロイしています (Terraformベース)...${NC}"
	@make terraform-start-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${YELLOW}テスト環境が準備されました。検証中...${NC}"
	@make verify-api-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに失敗しました${NC}"
	@echo -e "${BLUE}テスト完了。環境をクリーンアップしています...${NC}"
	@make terraform-stop-api-dev TF_ENV=$(TF_ENV)

# GraphQL開発環境のクイックテスト
terraform-test-graphql-dev:
	@echo -e "${BLUE}GraphQL一時テスト環境をデプロイしています (Terraformベース)...${NC}"
	@make terraform-start-graphql-dev TF_ENV=$(TF_ENV)
	@echo -e "${YELLOW}テスト環境が準備されました。検証中...${NC}"
	@make verify-graphql-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに失敗しました${NC}"
	@echo -e "${BLUE}テスト完了。環境をクリーンアップしています...${NC}"
	@make terraform-stop-graphql-dev TF_ENV=$(TF_ENV)

# 統合型ワークフローコマンド（新旧橋渡し）
.PHONY: start-api-dev pause-api-dev resume-api-dev stop-api-dev test-api-dev
.PHONY: start-graphql-dev pause-graphql-dev resume-graphql-dev stop-graphql-dev test-graphql-dev

# API開発環境のデプロイ（コア環境＋APIサービス）
# 修正後
start-api-dev:
	@echo -e "${BLUE}API開発環境を準備しています...${NC}"
	@chmod +x scripts/terraform/verify-deploy-prerequisites.sh
	@scripts/terraform/verify-deploy-prerequisites.sh $(TF_ENV) || exit 1
	@make terraform-start-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}API開発環境が準備されました${NC}"

# GraphQL開発環境のデプロイ
start-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を準備しています...${NC}"
	@chmod +x scripts/terraform/verify-deploy-prerequisites.sh
	@scripts/terraform/verify-deploy-prerequisites.sh $(TF_ENV) || exit 1
	@make terraform-start-graphql-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}GraphQL開発環境が準備されました${NC}"

# API開発環境の一時停止
pause-api-dev:
	@echo -e "${BLUE}API開発環境を一時停止しています...${NC}"
	@make terraform-pause-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}API開発環境は一時停止されました。コア基盤は維持されています。${NC}"

# GraphQL開発環境の一時停止
pause-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を一時停止しています...${NC}"
	@make terraform-pause-graphql-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}GraphQL開発環境は一時停止されました。コア基盤は維持されています。${NC}"

# API開発環境の再開
resume-api-dev:
	@echo -e "${BLUE}API開発環境を再開しています...${NC}"
	@make terraform-resume-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}API開発環境が再開されました${NC}"

# GraphQL開発環境の再開
resume-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を再開しています...${NC}"
	@make terraform-resume-graphql-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}GraphQL開発環境が再開されました${NC}"

# API開発環境の完全停止
stop-api-dev:
	@echo -e "${BLUE}API開発環境を完全に停止しています...${NC}"
	@make terraform-stop-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}残存リソースを確認しています...${NC}"
	@make check-resources TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}残存リソースがあれば削除を試みます...${NC}"
	@chmod +x scripts/terraform/aws-resource-cleaner.sh
	@scripts/terraform/aws-resource-cleaner.sh $(TF_ENV) auto
	@echo -e "${BLUE}クリーンアップ後の最終確認を行っています...${NC}"
	@make check-resources TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}API開発環境は完全に停止されました。すべてのリソースが削除されています。${NC}"

# GraphQL開発環境の完全停止
stop-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を完全に停止しています...${NC}"
	@make terraform-stop-graphql-dev TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}残存リソースを確認しています...${NC}"
	@make check-resources TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}残存リソースがあれば削除を試みます...${NC}"
	@chmod +x scripts/terraform/aws-resource-cleaner.sh
	@scripts/terraform/aws-resource-cleaner.sh $(TF_ENV) auto
	@echo -e "${BLUE}クリーンアップ後の最終確認を行っています...${NC}"
	@make check-resources TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}GraphQL開発環境は完全に停止されました。すべてのリソースが削除されています。${NC}"

# API開発環境のクイックテスト
test-api-dev:
	@echo -e "${BLUE}API一時テスト環境をデプロイしています...${NC}"
	@make terraform-test-api-dev TF_ENV=$(TF_ENV)

# GraphQL開発環境のクイックテスト
test-graphql-dev:
	@echo -e "${BLUE}GraphQL一時テスト環境をデプロイしています...${NC}"
	@make terraform-test-graphql-dev TF_ENV=$(TF_ENV)

# makefiles/terraform-workflow.mk に追加
.PHONY: terraform-verify terraform-import terraform-cleanup-safe tag-cleanup

# バージョン情報の定義（修正版）
VERSION_FILE := VERSION
VERIFY_TERRAFORM_VERSION := $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.1.0")
VERIFY_BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')  # ISO 8601形式に変更
VERIFY_GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# LDFLAGS構文の修正
VERIFY_LDFLAGS := -ldflags="-X main.version=$(VERIFY_TERRAFORM_VERSION) -X main.buildTime=$(VERIFY_BUILD_TIME) -X main.gitCommit=$(VERIFY_GIT_COMMIT)"

# ソースファイルと出力バイナリの指定
VERIFY_TERRAFORM_SRC := cmd/tools/verify-terraform.go
VERIFY_TERRAFORM_BIN := bin/verify-terraform

# AWS環境とTerraform状態の整合性を検証
terraform-verify:
	@echo -e "${BLUE}Terraform状態とAWS環境の整合性を検証しています...${NC}"
ifeq ($(USE_GO_RUN),1)
	@cd cmd/tools && go run verify-terraform.go -env $(TF_ENV) --ignore-resource-errors $(EXTRA_ARGS)
else
	@make build-terraform-verify
	@$(VERIFY_TERRAFORM_BIN) -env $(TF_ENV) --ignore-resource-errors $(EXTRA_ARGS)
endif


# ビルドルール - シンプルな実装
build-terraform-verify:
	@mkdir -p bin
	@echo -e "${BLUE}バイナリをビルドしています...${NC}"
	@go build $(VERIFY_LDFLAGS) -o $(VERIFY_TERRAFORM_BIN) $(VERIFY_TERRAFORM_SRC)
	@echo -e "${GREEN}ビルド完了: $(VERIFY_TERRAFORM_BIN) (v$(VERIFY_TERRAFORM_VERSION), $(VERIFY_BUILD_TIME))${NC}"

# バージョン表示
terraform-verify-version:
	@if [ -f "$(VERIFY_TERRAFORM_BIN)" ]; then \
		$(VERIFY_TERRAFORM_BIN) -version; \
	else \
	echo -e "${YELLOW}検証ツールがビルドされていません。ビルドを開始します...${NC}"; \
		make build-terraform-verify; \
		$(VERIFY_TERRAFORM_BIN) -version; \
	fi

# 既存リソースをTerraform状態にインポート
terraform-import:
	@echo -e "${BLUE}既存のAWSリソースをTerraform状態にインポートします...${NC}"
	@chmod +x scripts/terraform/aws-terraform-import.sh
	@scripts/terraform/aws-terraform-import.sh $(TF_ENV)
	@echo -e "${GREEN}インポートが完了しました${NC}"

# 安全なTerraform削除（検証後に実行）
terraform-cleanup-safe:
	@echo -e "${YELLOW}警告: terraform-cleanup-safe は非推奨です。代わりに stop-api-dev を使用してください${NC}"
	@echo -e "${BLUE}安全なTerraform削除を開始します...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@make stop-api-dev TF_ENV=$(TF_ENV)
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: 整合性の検証に失敗しました。${NC}"
	@echo -e "${GREEN}安全なクリーンアップが完了しました${NC}"

# タグベースのリソース削除（リカバリーオプション）
tag-cleanup:
	@echo -e "${RED}警告: タグベースでリソースを削除します！この操作は元に戻せません！${NC}"
	@chmod +x scripts/terraform/aws-resource-cleaner.sh
	@scripts/terraform/aws-resource-cleaner.sh $(TF_ENV)
	@echo -e "${GREEN}タグベースのクリーンアップが完了しました${NC}"

# GraphQLヘルスチェック検証ターゲット
.PHONY: verify-graphql-health

verify-graphql-health:
	@echo -e "${BLUE}GraphQLのヘルスチェックを検証しています...${NC}"
	@chmod +x scripts/verification/verify-graphql-health.sh
	@scripts/verification/verify-graphql-health.sh $(TF_ENV)