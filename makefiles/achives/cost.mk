# ===================================================================
# ファイル名: cost.mk
# 説明: AWS環境のコスト管理とリソース最適化のためのMakefile
#
# 用途:
#  - AWS環境のコスト見積もりと表示
#  - AWS環境のリソース状態確認
#  - レガシーコマンドの管理と新コマンドへの移行サポート
#
# 注意:
#  - cost-estimateはAWS環境の実際のリソースに基づいた見積もりを提供
#  - 古いクリーンアップコマンドは非推奨とし、代替コマンドを案内
#  - 非推奨コマンドは10秒のキャンセル猶予付きで実行されます
#
# 主要コマンド:
#  - cost-estimate: 現在のAWS環境のコスト見積もりを表示
#  - check-resources: 現在のAWSリソース状態を確認
#  - cleanup-*: [非推奨] AWS CLIによるリソース削除（Terraformステート非更新）
#
# 推奨代替コマンド:
#  - pause-dev/resume-dev: 環境の一時停止と再開
#  - stop-dev: 環境の完全停止
# ===================================================================
.PHONY: cost-estimate check-resources

# コスト見積もり
cost-estimate:
	@echo -e "${BLUE}AWS環境のコスト見積もりを取得しています...${NC}"
	@chmod +x scripts/terraform/cost-estimate.sh
	@scripts/terraform/cost-estimate.sh $(TF_ENV)


# リソース状態確認
check-resources:
	@echo -e "${BLUE}現在のAWSリソース状態を確認しています...${NC}"
	@chmod +x scripts/terraform/aws-status.sh
	@scripts/terraform/aws-status.sh $(TF_ENV)

#----------------------------------------
# 非推奨コマンド (互換性のために残す)
#----------------------------------------
# 廃止計画:
# 2025-05: 実行前にYを入力しないとキャンセルする仕様に変更
# 2025-06: エイリアスとして新コマンドを呼び出す形式に変更
# 2025-07: 完全削除
.PHONY: cleanup-minimal cleanup-standard cleanup-complete
.PHONY: verify-and-cleanup-api verify-and-cleanup-all temporary-deploy-api
.PHONY: force-cleanup-complete

# 最小限クリーンアップ（旧コマンド）- 現在は新コマンドへのエイリアス
cleanup-minimal:
	@echo -e "${RED}警告: このコマンドは非推奨です。${NC}"
	@echo -e "${YELLOW}代わりに 'make pause-api-dev TF_ENV=$(TF_ENV)' を使用してください。${NC}"
	@echo -e "${BLUE}3秒後に新コマンドを実行します...${NC}"
	@sleep 3
	@make pause-api-dev TF_ENV=$(TF_ENV)


# 標準クリーンアップ（旧コマンド）- 現在は新コマンドへのエイリアス
cleanup-standard:
	@echo -e "${RED}警告: このコマンドは非推奨です。${NC}"
	@echo -e "${YELLOW}代わりに 'make stop-api-dev TF_ENV=$(TF_ENV)' を使用してください。${NC}"
	@echo -e "${BLUE}3秒後に新コマンドを実行します...${NC}"
	@sleep 3
	@make stop-api-dev TF_ENV=$(TF_ENV)


# 完全クリーンアップ（旧コマンド）- 現在は新コマンドへのエイリアス
cleanup-complete:
	@echo -e "${RED}警告: このコマンドは非推奨です。${NC}"
	@echo -e "${YELLOW}代わりに 'make stop-api-dev TF_ENV=$(TF_ENV)' を使用してください。${NC}"
	@echo -e "${BLUE}3秒後に新コマンドを実行します...${NC}"
	@sleep 3
	@make stop-api-dev TF_ENV=$(TF_ENV)


# 検証後のクリーンアップ
verify-and-cleanup-api:
	@echo -e "${RED}警告: このコマンドは非推奨です。${NC}"
	@echo -e "${YELLOW}代わりに 'make test-api-dev TF_ENV=$(TF_ENV)' を使用してください。${NC}"
	@echo -e "${BLUE}3秒後に新コマンドを実行します...${NC}"
	@sleep 3
	@make test-api-dev TF_ENV=$(TF_ENV)

verify-and-cleanup-all:
	@echo -e "${RED}警告: このコマンドは非推奨です。${NC}"
	@echo -e "${YELLOW}代わりに 'make test-api-dev TF_ENV=$(TF_ENV)' を使用してください。${NC}"
	@echo -e "${BLUE}3秒後に新コマンドを実行します...${NC}"
	@sleep 3
	@make test-api-dev TF_ENV=$(TF_ENV)

# 一時デプロイと検証
temporary-deploy-api:
	@echo -e "${RED}警告: このコマンドは非推奨です。${NC}"
	@echo -e "${YELLOW}代わりに 'make test-api-dev TF_ENV=$(TF_ENV)' を使用してください。${NC}"
	@echo -e "${BLUE}3秒後に新コマンドを実行します...${NC}"
	@sleep 3
	@make test-api-dev TF_ENV=$(TF_ENV)

# 強制削除（旧コマンド）- 現在は新コマンドへのエイリアス
force-cleanup-complete:
	@echo -e "${RED}警告: このコマンドは非推奨です。${NC}"
	@echo -e "${YELLOW}代わりに 'make stop-api-dev TF_ENV=$(TF_ENV)' を使用してください。${NC}"
	@echo -e "${BLUE}3秒後に新コマンドを実行します...${NC}"
	@sleep 3
	@make stop-api-dev TF_ENV=$(TF_ENV)