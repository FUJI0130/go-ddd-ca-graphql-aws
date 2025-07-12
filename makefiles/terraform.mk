# ===================================================================
# ファイル名: terraform.mk
# 説明: AWS環境のデプロイとクリーンアップのためのMakefile
#
# 用途:
#  - Terraformを使用したAWS環境のデプロイ
#  - Terraformを使用したAWS環境のクリーンアップ
#  - Terraform状態ファイルの管理（バックアップ、リセットなど）
#
# 注意:
#  - 環境変数TF_ENVで対象環境を指定します（デフォルト: development）
#  - SERVICE_TYPEでサービス種別を指定します（api, graphql, grpc）
#  - クリーンアップ前に自動的に状態ファイルのバックアップを作成します
#
# 主要コマンド:
#  - deploy-api-dev: REST APIサービスをデプロイ
#  - deploy-graphql-dev: GraphQLサービスをデプロイ
#  - cleanup-api-dev: REST APIサービスとリソースを削除
#  - cleanup-graphql-dev: GraphQLサービスとリソースを削除
#  - terraform-backup: 状態ファイルのバックアップを作成
#  - terraform-reset: 状態ファイルをリセット（バックアップあり）
# ===================================================================

# Terraformクリーンアップコマンド
.PHONY: cleanup-api-dev cleanup-graphql-dev
# Terraform状態管理
.PHONY: terraform-backup terraform-reset

.PHONY: deploy-api-new-dev

# 新モジュールREST APIデプロイ
deploy-api-new-dev:
	@echo -e "${BLUE}新モジュールAPIアプリケーションを準備しています...${NC}"
	@if [ "$(SKIP_IMAGE)" != "1" ]; then \
		make prepare-ecr-image SERVICE_TYPE=api TF_ENV=$(TF_ENV); \
		make update-tfvars SERVICE_TYPE=api TF_ENV=$(TF_ENV); \
	fi
	@chmod +x scripts/terraform/terraform-deploy.sh
	@echo -e "${BLUE}デプロイ前の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ前に検証問題があります${NC}"
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) network
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) database
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-cluster
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) secrets
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) target-group-api-new
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) lb-api-new
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-api-new
	@echo -e "${GREEN}新モジュールAPI環境が準備されました${NC}"
	@echo -e "${BLUE}デプロイ後の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ後に検証問題があります${NC}"
	@make verify-api-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに失敗しましたが、サービスは起動している可能性があります${NC}"

# gRPCサービス再設計版デプロイ
.PHONY: deploy-grpc-new-dev

deploy-grpc-new-dev:
	@echo -e "${BLUE}新モジュールgRPCアプリケーションを準備しています...${NC}"
	@if [ "$(SKIP_IMAGE)" != "1" ]; then \
		make prepare-ecr-image SERVICE_TYPE=grpc TF_ENV=$(TF_ENV); \
		make update-tfvars SERVICE_TYPE=grpc TF_ENV=$(TF_ENV); \
	fi
	@chmod +x scripts/terraform/terraform-deploy.sh
	@echo -e "${BLUE}デプロイ前の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ前に検証問題があります${NC}"
	
	# 共有リソースのデプロイ（スキップ可能）
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) network
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) database
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-cluster
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) secrets
	
	# gRPC特化モジュールのデプロイ
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) target-group-grpc-new
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) target-group-grpc-native-new
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) lb-grpc-new
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) https-listener-new
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-grpc-new
	
	@echo -e "${GREEN}新モジュールgRPC環境が準備されました${NC}"
	
	# デプロイ後の検証
	@echo -e "${BLUE}デプロイ後の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ後に検証問題があります${NC}"
	@make verify-grpc-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: HTTP経由のヘルスチェックに失敗しましたが、サービスは起動している可能性があります${NC}"
	@make verify-grpc-native-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: gRPCネイティブヘルスチェックに失敗しましたが、サービスは起動している可能性があります${NC}"


# GraphQL新モジュール版デプロイ
.PHONY: deploy-graphql-new-dev

deploy-graphql-new-dev:
	@echo -e "${BLUE}新モジュールGraphQLアプリケーションを準備しています...${NC}"
	
	# 🆕 GraphQLスキーマ更新を最初に実行
	@echo -e "${BLUE}[0/8] GraphQLスキーマを更新しています...${NC}"
	@if command -v gqlgen >/dev/null 2>&1; then \
		echo -e "${GREEN}ローカルのgqlgenを使用します${NC}"; \
		gqlgen generate; \
	else \
		echo -e "${YELLOW}ローカルのgqlgenが見つかりません。自動セットアップします${NC}"; \
		go mod tidy > /dev/null 2>&1 || true; \
		go run github.com/99designs/gqlgen generate; \
	fi
	@echo -e "${GREEN}✓ GraphQLスキーマ更新完了${NC}"
	
	# スキーマ更新確認
	@if grep -q "deleteUser\|DeleteUser" internal/interface/graphql/generated/generated.go; then \
		echo -e "${GREEN}✓ deleteUser mutationが正常に生成されました${NC}"; \
	else \
		echo -e "${YELLOW}⚠ deleteUser mutationが見つかりません${NC}"; \
	fi
	
	# 既存のECRイメージ準備とデプロイ処理
	@if [ "$(SKIP_IMAGE)" != "1" ]; then \
		make prepare-ecr-image SERVICE_TYPE=graphql TF_ENV=$(TF_ENV); \
		make update-tfvars SERVICE_TYPE=graphql TF_ENV=$(TF_ENV); \
	fi
	@chmod +x scripts/terraform/terraform-deploy.sh
	@echo -e "${BLUE}デプロイ前の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ前に検証問題があります${NC}"
	
	# 共有リソースのデプロイ（スキップ可能）
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) network
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) database
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-cluster
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) secrets
	
	# GraphQL特化モジュールのデプロイ
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) target-group-graphql-new
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) lb-graphql-new
	@scripts/terraform/terraform-deploy.sh plan-apply $(TF_ENV) ecs-graphql-new
	
	@echo -e "${GREEN}新モジュールGraphQL環境が準備されました${NC}"
	
	# デプロイ後の検証
	@echo -e "${BLUE}デプロイ後の環境状態を検証しています...${NC}"
	@make terraform-verify TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: デプロイ後に検証問題があります${NC}"
	@make verify-graphql-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに失敗しましたが、サービスは起動している可能性があります${NC}"



.PHONY: terraform-cleanup-iam

terraform-cleanup-iam:
	@echo -e "${BLUE}IAMリソースをクリーンアップしています...${NC}"
	@chmod +x scripts/terraform/aws-iam-cleaner.sh
	@scripts/terraform/aws-iam-cleaner.sh $(TF_ENV)
	@echo -e "${GREEN}IAMリソースのクリーンアップが完了しました${NC}"

# 複合クリーンアップコマンド（API+gRPC）
.PHONY: cleanup-all-dev

# 5/16 動作確認OK
cleanup-all-dev:
	@echo -e "${BLUE}すべてのサービス（REST API + gRPC）を順次クリーンアップしています...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	
	@echo -e "${BLUE}ステップ1: gRPCサービスの安全クリーンアップ...${NC}"
	@make safe-cleanup-grpc-dev TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: gRPCクリーンアップで一部エラーが発生しました${NC}"
	@sleep 30  # リソース解放待機
	
	@echo -e "${BLUE}ステップ2: REST APIサービスとリソースのクリーンアップ...${NC}"
	@make cleanup-api-dev TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: REST APIクリーンアップで一部エラーが発生しました${NC}"
	
	@echo -e "${BLUE}ステップ3: GRAPHQLサービスとリソースのクリーンアップ...${NC}"
	@make safe-cleanup-graphql-dev TF_ENV=$(TF_ENV) || echo -e "${YELLOW}警告: GRAPHQLクリーンアップで一部エラーが発生しました${NC}"
	

	@echo -e "${BLUE}リソース状態を確認しています...${NC}"
	@make cost-estimate TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}すべてのサービスとリソースがクリーンアップされました${NC}"



# GraphQLクリーンアップ (safe-cleanup-graphql.sh方式)
.PHONY: safe-cleanup-graphql-dev

safe-cleanup-graphql-dev:
	@echo -e "${BLUE}GraphQL開発環境を安全に停止しています...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@chmod +x scripts/terraform/safe-cleanup-graphql.sh
	@scripts/terraform/safe-cleanup-graphql.sh $(TF_ENV)
	@echo -e "${BLUE}リソース状態を確認しています...${NC}"
	@make cost-estimate TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}GraphQL開発環境は安全に停止されました${NC}"



# REST APIクリーンアップ
cleanup-api-dev:
	@echo -e "${BLUE}API開発環境を完全に停止しています...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@make terraform-cleanup-iam TF_ENV=$(TF_ENV)
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV)
	@echo -e "${BLUE}リソース状態を確認しています...${NC}"
	@make cost-estimate TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}API開発環境は停止されました${NC}"



# 追加: 安全なgRPCクリーンアップ（段階的なリソース削除）
safe-cleanup-grpc-dev:
	@echo -e "${BLUE}gRPC開発環境を安全に停止しています...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@chmod +x scripts/terraform/safe-cleanup-grpc.sh
	@scripts/terraform/safe-cleanup-grpc.sh $(TF_ENV)
	@echo -e "${BLUE}リソース状態を確認しています...${NC}"
	@make cost-estimate TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}gRPC開発環境は安全に停止されました${NC}"

# gRPCクリーンアップ
cleanup-grpc-dev:
	@echo -e "${BLUE}gRPC開発環境を完全に停止しています...${NC}"
	@make terraform-backup TF_ENV=$(TF_ENV)
	@make terraform-cleanup-iam TF_ENV=$(TF_ENV)
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV)
	@echo -e "${BLUE}リソース状態を確認しています...${NC}"
	@make cost-estimate TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}gRPC開発環境は停止されました${NC}"

# Terraformステートのバックアップ
terraform-backup:
	@echo -e "${BLUE}Terraformステートをバックアップしています...${NC}"
	@cd deployments/terraform/environments/$(TF_ENV) && \
	mkdir -p backup-$(shell date +%Y%m%d) && \
	cp -r .terraform* terraform.tfstate* backup-$(shell date +%Y%m%d)/ 2>/dev/null || true && \
	echo -e "${GREEN}Terraformステートのバックアップが完了しました。${NC}"

# Terraformステートをリセット（バックアップ後）
terraform-reset: terraform-backup
	@echo -e "${BLUE}Terraformステートをリセットしています...${NC}"
	@cd deployments/terraform/environments/$(TF_ENV) && \
	rm -f terraform.tfstate* && \
	terraform init && \
	echo -e "${GREEN}Terraformステートのリセットが完了しました。${NC}"


# ===================================================================
# AWS環境マイグレーション関連ターゲット
# makefiles/terraform.mk への追加部分
# 
# 用途:
#  - AWS環境でのデータベースマイグレーション実行
#  - テストユーザーデータの投入
#  - マイグレーション結果の検証
# 
# 主要コマンド:
#  - migrate-aws-ci-dev: GitLab CI環境でマイグレーション実行
#  - migrate-aws-local-dev: ローカル環境でマイグレーション実行
#  - seed-test-users-dev: テストユーザーデータ投入
#  - deploy-with-migrate-dev: デプロイ→マイグレーション統合実行
#  - verify-migration-dev: マイグレーション結果検証
# ===================================================================

# AWS環境マイグレーション関連　
.PHONY: migrate-aws-ci-dev migrate-aws-local-dev migrate-aws-dev seed-test-users-dev deploy-with-migrate-dev verify-migration-dev

# GitLab CI環境でマイグレーション実行（新規）
migrate-aws-ci-dev:
	@echo -e "${BLUE}GitLab CI環境でマイグレーションを実行しています (環境: $(TF_ENV))...${NC}"
	@chmod +x scripts/terraform/aws-migrate-ci.sh
	@scripts/terraform/aws-migrate-ci.sh $(TF_ENV)
	@echo -e "${GREEN}GitLab CI環境マイグレーションが完了しました${NC}"

# ローカル環境でマイグレーション実行（既存）
migrate-aws-local-dev:
	@echo -e "${BLUE}ローカル環境でマイグレーションを実行しています (環境: $(TF_ENV))...${NC}"
	@chmod +x scripts/terraform/aws-migrate-local.sh
	@scripts/terraform/aws-migrate-local.sh $(TF_ENV)
	@echo -e "${GREEN}ローカル環境マイグレーションが完了しました${NC}"

# デフォルトマイグレーション（CI環境優先）
migrate-aws-dev:
	@echo -e "${BLUE}AWS環境でマイグレーションを実行しています (環境: $(TF_ENV))...${NC}"
	@echo -e "${YELLOW}注意: CI環境用スクリプトを使用します${NC}"
	@make migrate-aws-ci-dev TF_ENV=$(TF_ENV)

# テストユーザーデータ投入
seed-test-users-dev:
	@echo -e "${BLUE}テストユーザーデータを投入しています (環境: $(TF_ENV))...${NC}"
	@chmod +x scripts/terraform/aws-seed-users.sh  
	@scripts/terraform/aws-seed-users.sh $(TF_ENV)
	@echo -e "${GREEN}テストユーザーデータ投入が完了しました${NC}"

# 統合実行：デプロイ→マイグレーション→テストデータ→検証
deploy-with-migrate-dev:
	@echo -e "${BLUE}統合デプロイとマイグレーションを実行しています (環境: $(TF_ENV))...${NC}"
	@echo -e "${BLUE}ステップ1: GraphQLサービスをデプロイしています...${NC}"
	@make deploy-graphql-new-dev TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}ステップ2: CI環境でマイグレーションを実行しています...${NC}"
	@make migrate-aws-ci-dev TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}ステップ3: テストユーザーを投入しています...${NC}"
	@make seed-test-users-dev TF_ENV=$(TF_ENV)
	@echo -e "${BLUE}ステップ4: GraphQLヘルスチェックを実行しています...${NC}"
	@make verify-graphql-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}注意: ヘルスチェックに問題がありますが、続行します${NC}"
	@echo -e "${GREEN}統合デプロイとマイグレーションが完了しました${NC}"
	@echo -e "${BLUE}次のステップ: GraphQL認証テストを実行してください${NC}"
	@echo -e "  GraphQL Playground: http://\$$(cd deployments/terraform/environments/$(TF_ENV) && terraform output -raw graphql_alb_dns_name)/query"

# マイグレーション結果検証
verify-migration-dev:
	@echo -e "${BLUE}マイグレーション結果を検証しています (環境: $(TF_ENV))...${NC}"
	@chmod +x scripts/verification/verify-migration.sh
	@scripts/verification/verify-migration.sh $(TF_ENV)
	@echo -e "${GREEN}マイグレーション検証が完了しました${NC}"

# 段階的実行パターン
.PHONY: deploy-step-by-step-dev

deploy-step-by-step-dev:
	@echo -e "${BLUE}段階的デプロイを開始します (環境: $(TF_ENV))...${NC}"
	@echo -e "${YELLOW}各ステップで確認しながら進めます${NC}"
	@echo ""
	@echo -e "${BLUE}=== ステップ1: GraphQLサービスデプロイ ===${NC}"
	@read -p "GraphQLサービスをデプロイしますか？ (y/n): " CONFIRM_DEPLOY; \
	if [[ "$$CONFIRM_DEPLOY" =~ ^[Yy]$$ ]]; then \
		make deploy-graphql-new-dev TF_ENV=$(TF_ENV); \
		echo -e "${GREEN}✓ GraphQLサービスデプロイ完了${NC}"; \
	else \
		echo -e "${YELLOW}GraphQLサービスデプロイをスキップしました${NC}"; \
	fi
	@echo ""
	@echo -e "${BLUE}=== ステップ2: マイグレーション実行 ===${NC}"
	@read -p "CI環境でマイグレーションを実行しますか？ (y/n): " CONFIRM_MIGRATE; \
	if [[ "$$CONFIRM_MIGRATE" =~ ^[Yy]$$ ]]; then \
		make migrate-aws-ci-dev TF_ENV=$(TF_ENV); \
		echo -e "${GREEN}✓ マイグレーション実行完了${NC}"; \
	else \
		echo -e "${YELLOW}マイグレーション実行をスキップしました${NC}"; \
	fi
	@echo ""
	@echo -e "${BLUE}=== ステップ3: テストユーザー投入 ===${NC}"
	@read -p "テストユーザーデータを投入しますか？ (y/n): " CONFIRM_SEED; \
	if [[ "$$CONFIRM_SEED" =~ ^[Yy]$$ ]]; then \
		make seed-test-users-dev TF_ENV=$(TF_ENV); \
		echo -e "${GREEN}✓ テストユーザー投入完了${NC}"; \
	else \
		echo -e "${YELLOW}テストユーザー投入をスキップしました${NC}"; \
	fi
	@echo ""
	@echo -e "${BLUE}=== ステップ4: 動作確認 ===${NC}"
	@read -p "GraphQL動作確認を実行しますか？ (y/n): " CONFIRM_VERIFY; \
	if [[ "$$CONFIRM_VERIFY" =~ ^[Yy]$$ ]]; then \
		make verify-graphql-health TF_ENV=$(TF_ENV); \
		echo -e "${GREEN}✓ 動作確認完了${NC}"; \
	else \
		echo -e "${YELLOW}動作確認をスキップしました${NC}"; \
	fi
	@echo ""
	@echo -e "${GREEN}段階的デプロイが完了しました${NC}"

# デバッグ用：AWS環境の状態確認
.PHONY: debug-aws-migration-env

debug-aws-migration-env:
	@echo -e "${BLUE}AWS環境のマイグレーション関連状態を確認しています...${NC}"
	@echo -e "${BLUE}=== RDS情報（CI方式） ===${NC}"
	@aws rds describe-db-instances \
		--query 'DBInstances[?contains(DBInstanceIdentifier, `development`) && DBInstanceStatus == `available`].[DBInstanceIdentifier,Endpoint.Address,Engine,DBInstanceStatus]' \
		--output table --region $(AWS_REGION) 2>/dev/null || echo "RDS情報取得失敗"
	@echo ""
	@echo -e "${BLUE}=== ECSクラスター情報 ===${NC}"
	@aws ecs describe-clusters --clusters $(TF_ENV)-shared-cluster --region $(AWS_REGION) --query 'clusters[0].{Name:clusterName,Status:status,ActiveServicesCount:activeServicesCount,RunningTasksCount:runningTasksCount}' --output table 2>/dev/null || echo "ECSクラスター情報取得失敗"
	@echo ""
	@echo -e "${BLUE}=== ECRリポジトリ情報 ===${NC}"
	@aws ecr describe-repositories --repository-names $(TF_ENV)-test-management-migration --region $(AWS_REGION) --query 'repositories[0].{Name:repositoryName,URI:repositoryUri}' --output table 2>/dev/null || echo "マイグレーション用ECRリポジトリは未作成"
	@echo ""
	@echo -e "${BLUE}=== CloudWatchログ グループ ===${NC}"
	@aws logs describe-log-groups --log-group-name-prefix "/ecs/migration" --region $(AWS_REGION) --query 'logGroups[].{Name:logGroupName,CreationTime:creationTime}' --output table 2>/dev/null || echo "マイグレーション用ログ グループは未作成"

# トラブルシューティング用：マイグレーション関連リソースのクリーンアップ
.PHONY: cleanup-migration-resources-dev

cleanup-migration-resources-dev:
	@echo -e "${BLUE}マイグレーション関連リソースをクリーンアップしています...${NC}"
	@echo -e "${YELLOW}注意: これはマイグレーション用の一時リソースのみを削除します${NC}"
	@read -p "続行しますか？ (y/n): " CONFIRM_CLEANUP && \
	if [ "$$CONFIRM_CLEANUP" = "y" ] || [ "$$CONFIRM_CLEANUP" = "Y" ]; then \
		echo -e "${BLUE}ECRリポジトリを削除しています...${NC}"; \
		aws ecr delete-repository --repository-name $(TF_ENV)-test-management-migration --region $(AWS_REGION) --force 2>/dev/null || echo "マイグレーション用ECRリポジトリは存在しません"; \
		aws ecr delete-repository --repository-name $(TF_ENV)-test-management-seed --region $(AWS_REGION) --force 2>/dev/null || echo "シード用ECRリポジトリは存在しません"; \
		echo -e "${BLUE}CloudWatchログ グループを削除しています...${NC}"; \
		aws logs delete-log-group --log-group-name "/ecs/migration" --region $(AWS_REGION) 2>/dev/null || echo "マイグレーション用ログ グループは存在しません"; \
		aws logs delete-log-group --log-group-name "/ecs/seed-users" --region $(AWS_REGION) 2>/dev/null || echo "シード用ログ グループは存在しません"; \
		echo -e "${GREEN}マイグレーション関連リソースのクリーンアップが完了しました${NC}"; \
	else \
		echo -e "${YELLOW}クリーンアップをキャンセルしました${NC}"; \
	fi