# デプロイ関連Makefile
# makefiles/deploy.mk

#----------------------------------------
# インフラストラクチャコンポーネントデプロイ
#----------------------------------------
.PHONY: deploy-network deploy-database deploy-db-complete deploy-ecs-cluster

# 個別コンポーネントデプロイ
deploy-network:
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh apply $(TF_ENV) network

deploy-database:
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh apply $(TF_ENV) database

deploy-db-complete:
	@echo "データベースの完全デプロイを開始します..."
	@if [ -f ~/.env.terraform ]; then \
		echo "環境変数を読み込んでいます..."; \
		. ~/.env.terraform; \
	else \
		echo "警告: ~/.env.terraform が見つかりません。データベース認証情報が設定されているか確認してください。"; \
	fi
	@echo "現在のインフラ状況を確認しています..."
	@make tf-status
	@echo "データベースデプロイのプランを作成しています..."
	@make tf-plan MODULE=database
	@echo "データベースをデプロイしています..."
	@make deploy-database
	@echo "デプロイ結果を確認しています..."
	@make tf-status
	@echo "データベースデプロイプロセスが完了しました"

deploy-ecs-cluster:
	@echo "共有ECSクラスターをデプロイします（環境: $(TF_ENV)）..."
	@make tf-plan TF_ENV=$(TF_ENV) MODULE=shared-ecs-cluster
	@make tf-apply TF_ENV=$(TF_ENV)

#----------------------------------------
# サービスデプロイコマンド
#----------------------------------------
.PHONY: deploy-api deploy-graphql deploy-grpc deploy-all-services deploy-infrastructure
.PHONY: deploy-api-with-params deploy-graphql-with-params deploy-grpc-with-params deploy-all-with-params
.PHONY: redeploy-api deploy-service-with-logs

# サービスデプロイ (ECS + LB)
deploy-api: 
	@echo "APIサービスをデプロイします（環境: $(TF_ENV)）..."
	@make deploy-ecs-service SERVICE_TYPE=api TF_ENV=$(TF_ENV)
	@make deploy-service-lb SERVICE_TYPE=api TF_ENV=$(TF_ENV)

deploy-graphql:
	@echo "GraphQLサービスをデプロイします（環境: $(TF_ENV)）..."
	@make deploy-ecs-service SERVICE_TYPE=graphql TF_ENV=$(TF_ENV)
	@make deploy-service-lb SERVICE_TYPE=graphql TF_ENV=$(TF_ENV)

deploy-grpc:
	@echo "gRPCサービスをデプロイします（環境: $(TF_ENV)）..."
	@make deploy-ecs-service SERVICE_TYPE=grpc TF_ENV=$(TF_ENV)
	@make deploy-service-lb SERVICE_TYPE=grpc TF_ENV=$(TF_ENV)

deploy-all-services: deploy-api deploy-graphql deploy-grpc
	@echo "全サービスのデプロイが完了しました"

# パラメータ確認付きサービスデプロイ
deploy-api-with-params: verify-ssm-params deploy-api
deploy-graphql-with-params: verify-ssm-params deploy-graphql
deploy-grpc-with-params: verify-ssm-params deploy-grpc
deploy-all-with-params: verify-ssm-params deploy-all-services

# インフラ全体のデプロイ
deploy-infrastructure: deploy-ecs-cluster deploy-all-services
	@echo "インフラストラクチャのデプロイが完了しました"

# APIサービスの再デプロイ
redeploy-api:
	@echo "APIサービスを再デプロイしています..."
	@echo "1. サービスを停止中..."
	aws ecs update-service --cluster $(TF_ENV)-shared-cluster --service $(TF_ENV)-api --desired-count 0
	@echo "サービスが停止するまで待機中（30秒）..."
	sleep 30
	@echo "2. サービスを強制再デプロイで再開中..."
	aws ecs update-service --cluster $(TF_ENV)-shared-cluster --service $(TF_ENV)-api --desired-count 1 --force-new-deployment
	@echo "3. デプロイ状況を確認..."
	aws ecs describe-services --cluster $(TF_ENV)-shared-cluster --services $(TF_ENV)-api --query 'services[0].{Status:status,RunningCount:runningCount,DesiredCount:desiredCount}'
	@echo "4. ログ確認のために60秒待機..."
	sleep 60
	@echo "5. 最新のログを確認..."
	$(eval LOG_GROUP := /ecs/$(TF_ENV)-api)
	$(eval LOG_STREAM := $(shell aws logs describe-log-streams --log-group-name $(LOG_GROUP) --order-by LastEventTime --descending --limit 1 --query 'logStreams[0].logStreamName' --output text))
	aws logs get-log-events --log-group-name $(LOG_GROUP) --log-stream-name $(LOG_STREAM) --limit 30

#----------------------------------------
# サービスコンポーネントデプロイ補助コマンド
#----------------------------------------
.PHONY: deploy-ecs-service deploy-service-lb verify-service verify-all-services
.PHONY: prepare-service

# ECSサービスデプロイ
deploy-ecs-service:
	@if [ -z "$(SERVICE_TYPE)" ]; then \
		echo -e "${RED}エラー: SERVICE_TYPE環境変数を指定してください${NC}"; \
		echo "例: SERVICE_TYPE=api make deploy-ecs-service"; \
		exit 1; \
	fi
	@echo "ECSサービスをデプロイします（サービス: $(SERVICE_TYPE), 環境: $(TF_ENV)）..."
	@make tf-plan TF_ENV=$(TF_ENV) MODULE=ecs-service SERVICE_TYPE=$(SERVICE_TYPE)
	@make tf-apply TF_ENV=$(TF_ENV)

# ロードバランサーデプロイ
deploy-service-lb:
	@if [ -z "$(SERVICE_TYPE)" ]; then \
		echo -e "${RED}エラー: SERVICE_TYPE環境変数を指定してください${NC}"; \
		echo "例: SERVICE_TYPE=api make deploy-service-lb"; \
		exit 1; \
	fi
	@echo "サービス用ロードバランサーをデプロイします（サービス: $(SERVICE_TYPE), 環境: $(TF_ENV)）..."
	@make tf-plan TF_ENV=$(TF_ENV) MODULE=load-balancer SERVICE_TYPE=$(SERVICE_TYPE)
	@make tf-apply TF_ENV=$(TF_ENV)

# サービス検証
verify-service:
	@if [ -z "$(SERVICE_TYPE)" ]; then \
		echo -e "${RED}エラー: SERVICE_TYPE環境変数を指定してください${NC}"; \
		echo "例: SERVICE_TYPE=api make verify-service"; \
		exit 1; \
	fi
	@echo "サービスの動作を検証しています（サービス: $(SERVICE_TYPE), 環境: $(TF_ENV)）..."
	@chmod +x scripts/verification/verify-ecs-$(SERVICE_TYPE).sh
	@scripts/verification/verify-ecs-$(SERVICE_TYPE).sh $(TF_ENV)

verify-all-services:
	@echo "全サービスの動作を検証しています（環境: $(TF_ENV)）..."
	@make verify-service SERVICE_TYPE=api TF_ENV=$(TF_ENV) || echo "警告: APIサービスの検証に失敗しました"
	@make verify-service SERVICE_TYPE=graphql TF_ENV=$(TF_ENV) || echo "警告: GraphQLサービスの検証に失敗しました"
	@make verify-service SERVICE_TYPE=grpc TF_ENV=$(TF_ENV) || echo "警告: gRPCサービスの検証に失敗しました"
	@echo "全サービスの検証が完了しました"

# 単一サービス準備
prepare-service:
	@echo "サービスの準備を開始します ($(SERVICE_TYPE))..."
	@make prepare-ecr-image SERVICE_TYPE=$(SERVICE_TYPE) TF_ENV=$(TF_ENV)
	@make update-tfvars SERVICE_TYPE=$(SERVICE_TYPE) TF_ENV=$(TF_ENV)
	@make verify-ssm-params TF_ENV=$(TF_ENV)

# サービスデプロイとログ確認
deploy-service-with-logs:
	@echo "サービスのデプロイと検証を開始します ($(SERVICE_TYPE))..."
	@make deploy-$(SERVICE_TYPE) TF_ENV=$(TF_ENV)
	@echo "デプロイの安定化を待機しています (30秒)..."
	@sleep 30
	@LOG_GROUP="/ecs/$(TF_ENV)-$(SERVICE_TYPE)" && \
	LOG_STREAM=$$(aws logs describe-log-streams --log-group-name $$LOG_GROUP --order-by LastEventTime --descending --limit 1 --query 'logStreams[0].logStreamName' --output text) && \
	echo "最新のログを確認しています..." && \
	aws logs get-log-events --log-group-name $$LOG_GROUP --log-stream-name $$LOG_STREAM --limit 20
	@make verify-service SERVICE_TYPE=$(SERVICE_TYPE) TF_ENV=$(TF_ENV) || (echo "警告: $(SERVICE_TYPE)サービスの検証に失敗しました。ログを確認してください。" && exit 1)

#----------------------------------------
# アプリケーションデプロイワークフロー
#----------------------------------------
.PHONY: deploy-app-workflow deploy-app-workflow-secure

# 完全デプロイワークフロー
deploy-app-workflow:
	@echo "アプリケーションのデプロイワークフローを開始します..."
	@echo "1. ECRイメージを準備しています..."
	@make prepare-all-ecr-images TF_ENV=$(TF_ENV)
	@echo "2. terraform.tfvarsを更新しています..."
	@make update-tfvars-all TF_ENV=$(TF_ENV)
	@echo "3. インフラストラクチャをデプロイしています..."
	@make deploy-infrastructure TF_ENV=$(TF_ENV)
	@echo "4. デプロイの安定化を待機しています (60秒)..."
	@sleep 60
	@echo "5. デプロイ結果を検証しています..."
	@make verify-all-services TF_ENV=$(TF_ENV) || (echo "警告: 一部のサービス検証に失敗しました。詳細なログを確認してください。" && exit 1)
	@echo "アプリケーションデプロイワークフローが完了しました"

# セキュアデプロイワークフロー (パラメータ確認付き)
deploy-app-workflow-secure:
	@echo "アプリケーションのセキュアなデプロイワークフローを開始します..."
	@echo "1. SSMパラメータを確認・作成しています..."
	@make verify-ssm-params TF_ENV=$(TF_ENV)
	@echo "2. ECRイメージを準備しています..."
	@make prepare-all-ecr-images TF_ENV=$(TF_ENV)
	@echo "3. terraform.tfvarsを更新しています..."
	@make update-tfvars-all TF_ENV=$(TF_ENV)
	@echo "4. インフラストラクチャをデプロイしています..."
	@make deploy-infrastructure TF_ENV=$(TF_ENV)
	@echo "5. デプロイの安定化を待機しています (60秒)..."
	@sleep 60
	@echo "6. デプロイ結果を検証しています..."
	@make verify-all-services TF_ENV=$(TF_ENV) || (echo "警告: 一部のサービス検証に失敗しました。詳細なログを確認してください。" && exit 1)
	@echo "アプリケーションデプロイワークフローが完了しました"