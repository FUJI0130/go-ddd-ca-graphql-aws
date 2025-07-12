# ===================================================================
# ファイル名: terraform-deploy-verify.mk
# 説明: AWS環境のデプロイ検証とライフサイクル管理のためのMakefile
#
# 用途:
#  - Terraformベースのデプロイプロセスの自動検証
#  - デプロイ→ヘルスチェック→クリーンアップの完全サイクル検証
#  - AWS環境の効率的な検証とリソース管理
#
# 注意:
#  - このMakefileはterraform-workflow.mkとcost.mkの機能を使用します
#  - 環境変数TF_ENVで対象環境を指定します（デフォルト: development）
#  - 検証完了後はリソースを適切にクリーンアップしてコスト最適化を図ってください
#
# 主要コマンド:
#  - api-deploy-verify: APIサービスのデプロイとヘルスチェックを実行
#  - full-cycle-verify: デプロイ、検証、クリーンアップの完全サイクルを実行
#  - quick-verify: タイムアウト付きの一時検証環境を作成（自動クリーンアップ）
# ===================================================================

# デプロイ前の前提条件確認（AWS環境とTerraform状態の両方）
.PHONY: verify-deploy-prerequisites

verify-deploy-prerequisites:
	@echo -e "${BLUE}デプロイ前の前提条件を確認しています...${NC}"
	@chmod +x scripts/terraform/verify-deploy-prerequisites.sh
	@scripts/terraform/verify-deploy-prerequisites.sh $(TF_ENV)

# APIデプロイ検証コマンド（ビルド〜デプロイ〜ヘルスチェック）
api-deploy-verify:
	@echo -e "${BLUE}APIデプロイの検証を開始します...${NC}"
	@echo -e "${YELLOW}ステップ 0/4: デプロイ前提条件を確認${NC}"
	@make verify-deploy-prerequisites TF_ENV=$(TF_ENV) || exit 1
	@echo -e "${YELLOW}ステップ 1/4: APIサービスをビルド${NC}"
	@make build SERVICE_TYPE=api
	@echo -e "${YELLOW}ステップ 2/4: ECRイメージを準備${NC}"
	@make prepare-ecr-image SERVICE_TYPE=api TF_ENV=$(TF_ENV)
	@echo -e "${YELLOW}ステップ 3/4: API環境をデプロイ${NC}"
	@make start-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${YELLOW}ステップ 4/4: APIヘルスチェックを実行${NC}"
	@make verify-api-health TF_ENV=$(TF_ENV) || \
		(echo -e "${RED}ヘルスチェックに失敗しました。${NC}" && exit 1)
	@echo -e "\n${GREEN}デプロイ検証成功: API環境が正常に動作しています${NC}"
	@make check-resources TF_ENV=$(TF_ENV)
	@make cost-estimate TF_ENV=$(TF_ENV)

# 完全サイクル検証（デプロイ〜検証〜削除）
full-cycle-verify:
	@echo -e "${BLUE}完全サイクル検証を開始します...${NC}"
	@make verify-deploy-prerequisites TF_ENV=$(TF_ENV) || exit 1
	@make api-deploy-verify TF_ENV=$(TF_ENV) || \
		(echo -e "${RED}デプロイ検証に失敗しました。環境をクリーンアップします...${NC}" && \
		make stop-api-dev TF_ENV=$(TF_ENV) && exit 1)
	@echo -e "${YELLOW}検証成功。環境のクリーンアップを開始します...${NC}"
	@make stop-api-dev TF_ENV=$(TF_ENV)
	@echo -e "${GREEN}完全サイクル検証が成功しました。環境は正常にクリーンアップされました${NC}"

# タイムアウト付き一時検証環境
quick-verify: export TIMEOUT_MINUTES ?= 30
quick-verify:
	@echo -e "${BLUE}タイムアウト付き検証環境を開始します (${TIMEOUT_MINUTES}分後に自動削除)${NC}"
	@make verify-deploy-prerequisites TF_ENV=$(TF_ENV) || exit 1
	@make api-deploy-verify TF_ENV=$(TF_ENV) || \
		(echo -e "${RED}デプロイに失敗しました。${NC}" && exit 1)
	@echo -e "${GREEN}検証環境が準備されました。${TIMEOUT_MINUTES}分後に自動削除されます${NC}"
	@echo -e "現在の日時: $(shell date)"
	@echo -e "予定削除時刻: $(shell date -d "+${TIMEOUT_MINUTES} minutes")"
	@( sleep $(shell echo "${TIMEOUT_MINUTES}*60" | bc) && make stop-api-dev TF_ENV=$(TF_ENV) ) &
	@echo $$! > .auto-cleanup-pid
	@echo -e "${YELLOW}注意: 自動削除をキャンセルするには 'kill $$(cat .auto-cleanup-pid)' を実行してください${NC}"

# GraphQLサービスのデプロイ検証（今後の拡張用）
graphql-deploy-verify:
	@echo -e "${BLUE}GraphQLデプロイの検証を開始します...${NC}"
	@echo -e "${YELLOW}この機能は現在準備中です${NC}"
	@echo -e "${RED}未実装${NC}"

# gRPCサービスのデプロイ検証（今後の拡張用）
grpc-deploy-verify:
	@echo -e "${BLUE}gRPCデプロイの検証を開始します...${NC}"
	@echo -e "${YELLOW}この機能は現在準備中です${NC}"
	@echo -e "${RED}未実装${NC}"