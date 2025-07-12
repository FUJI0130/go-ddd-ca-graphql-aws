# ===================================================================
# ファイル名: terraform-cleanup.mk
# 説明: Terraformベースのクリーンアップと状態管理のためのMakefile
#
# 用途:
#  - Terraformを使用したAWSリソースの段階的な削除
#  - Terraform状態ファイルの管理（バックアップ、リセット、検証）
#  - AWS環境とTerraform状態の整合性確保
#  - 安全なクリーンアップワークフローの提供
#
# 注意:
#  - このMakefileのコマンドはTerraform状態ファイルを更新します
#  - クリーンアップ前に自動的に状態ファイルのバックアップを作成します
#  - terraform-verifyでTerraform状態とAWS環境の整合性を検証できます
#
# 主要コマンド:
#  - terraform-backup: 状態ファイルのバックアップを作成
#  - terraform-reset: 状態ファイルをリセット（バックアップあり）
#  - terraform-verify: AWS環境とTerraform状態の整合性を検証
#  - terraform-cleanup-minimal: ECSサービスとALBを削除
#  - terraform-cleanup-standard: RDSを含むリソースを削除（コア基盤は維持）
#  - terraform-cleanup-complete: VPCを含む全リソースを削除
#  - terraform-safe-cleanup: バックアップ、削除、検証を一連で実行
# ===================================================================
.PHONY: terraform-backup terraform-reset terraform-verify
.PHONY: terraform-cleanup-minimal terraform-cleanup-standard terraform-cleanup-complete
.PHONY: terraform-safe-cleanup

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

# Terraformステートと実際のAWS環境の一致を検証
# terraform-verify ターゲットは terraform-workflow.mk に統合されました
# 使用方法: make terraform-verify TF_ENV=<環境名>

# 修正後
terraform-cleanup-minimal: terraform-backup
	@echo -e "${YELLOW}警告: terraform-cleanup-minimal は非推奨です。代わりに pause-api-dev を使用してください${NC}"
	@echo -e "${YELLOW}将来のバージョンでこのコマンドは削除されます${NC}"
	@read -p "続行しますか？(y/n) " CONTINUE; \
	if [ "$$CONTINUE" != "y" ]; then \
		echo "中止します"; \
		exit 1; \
	fi
	@echo -e "${BLUE}Terraformで最小限クリーンアップを実行しています...${NC}"
	@cd deployments/terraform/environments/$(TF_ENV) && \
	terraform destroy -auto-approve \
	-target=module.service_api -target=module.service_graphql -target=module.service_grpc \
	-target=module.loadbalancer_api -target=module.loadbalancer_graphql -target=module.loadbalancer_grpc \
	-target=module.target_group_api -target=module.target_group_graphql -target=module.target_group_grpc && \
	echo -e "${GREEN}Terraformでの最小限クリーンアップが完了しました。${NC}"

# 標準クリーンアップ（最小限 + RDS）- Terraformで実行
terraform-cleanup-standard: terraform-backup
	@echo -e "${YELLOW}警告: terraform-cleanup-standard は非推奨です。代わりに stop-api-dev を使用してください${NC}"
	@echo -e "${YELLOW}将来のバージョンでこのコマンドは削除されます${NC}"
	@read -p "続行しますか？(y/n) " CONTINUE; \
	if [ "$$CONTINUE" != "y" ]; then \
		echo "中止します"; \
		exit 1; \
	fi
	@echo -e "${BLUE}Terraformで標準クリーンアップを実行しています...${NC}"
	@cd deployments/terraform/environments/$(TF_ENV) && \
	terraform destroy -auto-approve \
	-target=module.service_api -target=module.service_graphql -target=module.service_grpc \
	-target=module.loadbalancer_api -target=module.loadbalancer_graphql -target=module.loadbalancer_grpc \
	-target=module.target_group_api -target=module.target_group_graphql -target=module.target_group_grpc \
	-target=module.database && \
	echo -e "${GREEN}Terraformでの標準クリーンアップが完了しました。${NC}"

# 完全クリーンアップ - Terraformで実行
terraform-cleanup-complete: terraform-backup
	@echo -e "${YELLOW}警告: terraform-cleanup-complete は非推奨です。代わりに stop-api-dev を使用してください${NC}"
	@echo -e "${YELLOW}将来のバージョンでこのコマンドは削除されます${NC}"
	@read -p "続行しますか？(y/n) " CONTINUE; \
	if [ "$$CONTINUE" != "y" ]; then \
		echo "中止します"; \
		exit 1; \
	fi
	@echo -e "${BLUE}Terraformで完全クリーンアップを実行しています...${NC}"
	@cd deployments/terraform/environments/$(TF_ENV) && \
	terraform destroy -auto-approve && \
	echo -e "${GREEN}Terraformでの完全クリーンアップが完了しました。${NC}"

# 安全なクリーンアップ（ステートバックアップ、クリーンアップ、検証）
terraform-safe-cleanup: terraform-backup
	@echo -e "${BLUE}安全なクリーンアップを実行しています...${NC}"
	@make terraform-cleanup-minimal TF_ENV=$(TF_ENV)
	@make terraform-verify TF_ENV=$(TF_ENV) || \
	(echo -e "${RED}警告: Terraformステートと実際のAWS環境に不一致があります。${NC}" && \
	echo -e "${YELLOW}修復するには: make terraform-reset TF_ENV=$(TF_ENV)${NC}")
	@echo -e "${GREEN}安全なクリーンアップ処理が完了しました。${NC}"