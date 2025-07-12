# frontend/makefiles/frontend-aws.mk
# フロントエンドAWS環境デプロイ用Makefile（バックエンドパターン踏襲）

#----------------------------------------
# 色設定（バックエンドと統一）
#----------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

#----------------------------------------
# 基本変数
#----------------------------------------
FRONTEND_TF_ENV ?= development
FRONTEND_DIR := $(shell pwd)
BACKEND_TF_DIR := ../deployments/terraform/environments/$(FRONTEND_TF_ENV)

#----------------------------------------
# フロントエンドAWSデプロイコマンド
#----------------------------------------
.PHONY: deploy-frontend-dev get-backend-dns build-frontend-assets upload-frontend-assets
.PHONY: create-frontend-s3-dev create-frontend-cloudfront-dev
.PHONY: cleanup-frontend-dev verify-frontend-health

# 完全デプロイコマンド（バックエンドパターン踏襲）
deploy-frontend-dev:
	@echo -e "${BLUE}フロントエンド開発環境をデプロイしています...${NC}"
	@echo -e "${BLUE}ステップ1: バックエンドDNS名を取得しています...${NC}"
	@make get-backend-dns FRONTEND_TF_ENV=$(FRONTEND_TF_ENV)
	@echo -e "${BLUE}ステップ2: フロントエンドをビルドしています...${NC}"
	@make build-frontend-assets FRONTEND_TF_ENV=$(FRONTEND_TF_ENV)
	@echo -e "${BLUE}ステップ3: AWS環境をプロビジョニングしています...${NC}"
	@make terraform-deploy-frontend FRONTEND_TF_ENV=$(FRONTEND_TF_ENV)
	@echo -e "${BLUE}ステップ4: フロントエンドファイルをアップロードしています...${NC}"
	@make upload-frontend-assets FRONTEND_TF_ENV=$(FRONTEND_TF_ENV)
	@echo -e "${BLUE}ステップ5: CloudFrontキャッシュを無効化しています...${NC}"
	@make invalidate-cloudfront-cache FRONTEND_TF_ENV=$(FRONTEND_TF_ENV)
	@echo -e "${BLUE}ステップ6: デプロイメントを検証しています...${NC}"
	@make verify-frontend-health FRONTEND_TF_ENV=$(FRONTEND_TF_ENV)
	@echo -e "${GREEN}フロントエンド開発環境のデプロイが完了しました${NC}"

# バックエンドDNS名取得（terraform outputから動的取得）
get-backend-dns:
	@echo -e "${BLUE}バックエンドGraphQL ALB DNS名を取得しています...${NC}"
	@if [ -d "$(BACKEND_TF_DIR)" ]; then \
		cd $(BACKEND_TF_DIR) && \
		GRAPHQL_ALB_DNS_NAME=$$(terraform output -raw graphql_alb_dns_name 2>/dev/null || echo ""); \
		if [ -n "$GRAPHQL_ALB_DNS_NAME" ]; then \
			echo "GRAPHQL_ALB_DNS_NAME=$$GRAPHQL_ALB_DNS_NAME" >> ~/.env.terraform; \
			echo -e "${GREEN}✓ GraphQL ALB DNS名を取得しました: $$GRAPHQL_ALB_DNS_NAME${NC}"; \
		else \
			echo -e "${YELLOW}⚠ GraphQL ALB DNS名が見つかりません。バックエンドがデプロイされているか確認してください${NC}"; \
		fi; \
	else \
		echo -e "${RED}エラー: バックエンドTerraformディレクトリが見つかりません: $(BACKEND_TF_DIR)${NC}"; \
		exit 1; \
	fi

# フロントエンドビルド（環境変数を動的生成）
build-frontend-assets:
	@echo -e "${BLUE}フロントエンドアセットをビルドしています...${NC}"
	@chmod +x scripts/build-frontend.sh
	@scripts/build-frontend.sh $(FRONTEND_TF_ENV)
	@echo -e "${GREEN}✓ フロントエンドビルドが完了しました${NC}"

# Terraformデプロイ実行
terraform-deploy-frontend:
	@echo -e "${BLUE}フロントエンドTerraformをデプロイしています...${NC}"
	@chmod +x scripts/deploy-frontend.sh
	@scripts/deploy-frontend.sh plan-apply $(FRONTEND_TF_ENV) all
	@echo -e "${GREEN}✓ Terraformデプロイが完了しました${NC}"

# S3へのファイルアップロード
upload-frontend-assets:
	@echo -e "${BLUE}S3バケットにファイルをアップロードしています...${NC}"
	@chmod +x scripts/upload-frontend.sh
	@scripts/upload-frontend.sh $(FRONTEND_TF_ENV)
	@echo -e "${GREEN}✓ ファイルアップロードが完了しました${NC}"

# CloudFrontキャッシュ無効化
invalidate-cloudfront-cache:
	@echo -e "${BLUE}CloudFrontキャッシュを無効化しています...${NC}"
	@chmod +x scripts/invalidate-cache.sh
	@scripts/invalidate-cache.sh $(FRONTEND_TF_ENV)
	@echo -e "${GREEN}✓ キャッシュ無効化が完了しました${NC}"

# 個別コンポーネントデプロイ（段階的デプロイ対応）
create-frontend-s3-dev:
	@echo -e "${BLUE}S3ホスティング環境を作成しています...${NC}"
	@scripts/deploy-frontend.sh plan-apply $(FRONTEND_TF_ENV) s3-hosting
	@echo -e "${GREEN}✓ S3ホスティング環境が作成されました${NC}"

create-frontend-cloudfront-dev:
	@echo -e "${BLUE}CloudFront配信環境を作成しています...${NC}"
	@scripts/deploy-frontend.sh plan-apply $(FRONTEND_TF_ENV) cloudfront
	@echo -e "${GREEN}✓ CloudFront配信環境が作成されました${NC}"

# フロントエンド検証
verify-frontend-health:
	@echo -e "${BLUE}フロントエンドの動作を検証しています...${NC}"
	@chmod +x scripts/verify-frontend-health.sh
	@scripts/verify-frontend-health.sh $(FRONTEND_TF_ENV)

# クリーンアップ
cleanup-frontend-dev:
	@echo -e "${BLUE}フロントエンド開発環境をクリーンアップしています...${NC}"
	@chmod +x scripts/cleanup-frontend.sh
	@scripts/cleanup-frontend.sh $(FRONTEND_TF_ENV)
	@echo -e "${GREEN}✓ フロントエンドクリーンアップが完了しました${NC}"

#----------------------------------------
# ユーティリティコマンド
#----------------------------------------
.PHONY: frontend-cost-estimate frontend-status

# フロントエンドリソースのコスト見積もり
frontend-cost-estimate:
	@echo -e "${BLUE}フロントエンドAWSリソースのコスト見積もりを取得しています...${NC}"
	@echo "S3バケット使用量:"
	@aws s3api list-objects-v2 --bucket $(FRONTEND_TF_ENV)-test-management-frontend --query 'Contents[].Size' --output text 2>/dev/null | awk '{sum += $1} END {printf "%.2f MB\n", sum/1024/1024}' || echo "バケット未作成またはファイルなし"
	@echo "CloudFrontディストリビューション:"
	@aws cloudfront list-distributions --query 'DistributionList.Items[?Comment==`$(FRONTEND_TF_ENV)-test-management-frontend CloudFront distribution`].{Id:Id,Status:Status,DomainName:DomainName}' --output table 2>/dev/null || echo "ディストリビューション未作成"

# フロントエンド環境状態確認
frontend-status:
	@echo -e "${BLUE}フロントエンド環境の状態を確認しています...${NC}"
	@echo "=== S3バケット状態 ==="
	@aws s3api head-bucket --bucket $(FRONTEND_TF_ENV)-test-management-frontend 2>/dev/null && echo "✓ S3バケット存在" || echo "✗ S3バケット未作成"
	@echo "=== CloudFront状態 ==="
	@aws cloudfront list-distributions --query 'DistributionList.Items[?Comment==`$(FRONTEND_TF_ENV)-test-management-frontend CloudFront distribution`].Status' --output text 2>/dev/null | head -1 | { read status; [ -n "$status" ] && echo "✓ CloudFront: $status" || echo "✗ CloudFront未作成"; }
	@echo "=== バックエンド接続 ==="
	@source ~/.env.terraform 2>/dev/null && [ -n "$GRAPHQL_ALB_DNS_NAME" ] && echo "✓ GraphQL ALB: $GRAPHQL_ALB_DNS_NAME" || echo "✗ GraphQL ALB DNS未設定"

#----------------------------------------
# 統合コマンド（バックエンド + フロントエンド）
#----------------------------------------
.PHONY: deploy-full-stack-dev cleanup-full-stack-dev

# フルスタックデプロイ（バックエンド → フロントエンド）
deploy-full-stack-dev:
	@echo -e "${BLUE}フルスタック環境をデプロイしています...${NC}"
	@echo -e "${BLUE}ステップ1: バックエンドをデプロイ中...${NC}"
	@cd .. && make deploy-graphql-new-dev TF_ENV=$(FRONTEND_TF_ENV)
	@echo -e "${BLUE}ステップ2: フロントエンドをデプロイ中...${NC}"
	@make deploy-frontend-dev FRONTEND_TF_ENV=$(FRONTEND_TF_ENV)
	@echo -e "${GREEN}フルスタック環境のデプロイが完了しました${NC}"
	@echo -e "${BLUE}=== アクセス情報 ===${NC}"
	@cd deployments/terraform/environments/$(FRONTEND_TF_ENV) && \
	echo -e "フロントエンドURL: $(terraform output -raw cloudfront_url)" && \
	echo -e "バックエンドAPI: $(terraform output -raw backend_graphql_alb_dns_name)"

# フルスタッククリーンアップ
cleanup-full-stack-dev:
	@echo -e "${BLUE}フルスタック環境をクリーンアップしています...${NC}"
	@echo -e "${BLUE}ステップ1: フロントエンドをクリーンアップ中...${NC}"
	@make cleanup-frontend-dev FRONTEND_TF_ENV=$(FRONTEND_TF_ENV) || echo -e "${YELLOW}フロントエンドクリーンアップで問題が発生しました${NC}"
	@echo -e "${BLUE}ステップ2: バックエンドをクリーンアップ中...${NC}"
	@cd .. && make safe-cleanup-graphql-dev TF_ENV=$(FRONTEND_TF_ENV) || echo -e "${YELLOW}バックエンドクリーンアップで問題が発生しました${NC}"
	@echo -e "${GREEN}フルスタック環境のクリーンアップが完了しました${NC}"