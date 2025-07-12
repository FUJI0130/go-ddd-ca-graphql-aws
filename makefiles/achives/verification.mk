# ===================================================================
# ファイル名: verification.mk
# 説明: AWS環境の検証用Makefile
#
# 用途:
#  - デプロイされたサービスのヘルスチェック検証
#  - REST API、GraphQL、gRPCサービスの個別検証
#  - 検証スクリプトの実行とステータス報告
#
# 注意:
#  - 検証スクリプトはAWS CLIを使用してリソースにアクセスします
#  - 環境変数TF_ENVで対象環境を指定します（デフォルト: development）
#  - サービスが完全に起動するまで検証が成功しない場合があります
#
# 主要コマンド:
#  - verify-api-health: REST APIのヘルスチェックを検証
#  - verify-all-services: すべてのサービスのヘルスチェックを検証
# ===================================================================

#----------------------------------------
# サービス検証コマンド
#----------------------------------------
.PHONY: verify-api-health verify-graphql-health verify-grpc-health verify-all-services

# REST APIヘルスチェック検証
verify-api-health:
	@echo -e "${BLUE}REST APIのヘルスチェックを検証しています...${NC}"
	@chmod +x scripts/verification/verify-api-health.sh
	@scripts/verification/verify-api-health.sh $(TF_ENV)

# GraphQLヘルスチェック検証（TODO: 実装）
verify-graphql-health:
	@echo -e "${YELLOW}注意: GraphQLヘルスチェック検証は未実装です${NC}"
	@echo -e "${BLUE}GraphQLのヘルスチェックを検証しています...${NC}"
	@echo -e "${YELLOW}GraphQLサービスのヘルスチェック検証スクリプトは今後実装予定です${NC}"

# gRPCヘルスチェック検証（TODO: 実装）
verify-grpc-health:
	@echo -e "${YELLOW}注意: gRPCヘルスチェック検証は未実装です${NC}"
	@echo -e "${BLUE}gRPCのヘルスチェックを検証しています...${NC}"
	@echo -e "${YELLOW}gRPCサービスのヘルスチェック検証スクリプトは今後実装予定です${NC}"

# すべてのサービスヘルスチェック検証
verify-all-health:
	@echo -e "${BLUE}すべてのサービスのヘルスチェックを検証しています...${NC}"
	@make verify-api-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}APIヘルスチェックに失敗しました${NC}"
	@make verify-graphql-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}GraphQLヘルスチェックに失敗しました${NC}"
	@make verify-grpc-health TF_ENV=$(TF_ENV) || echo -e "${YELLOW}gRPCヘルスチェックに失敗しました${NC}"
	@echo -e "${GREEN}すべてのサービス検証が完了しました${NC}"