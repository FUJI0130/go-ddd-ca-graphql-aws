# AWS/Terraform関連Makefile
# makefiles/aws.mk

#----------------------------------------
# AWS/Terraform関連コマンド
#----------------------------------------
.PHONY: tf-status tf-init tf-plan tf-apply tf-destroy
.PHONY: verify-ssm-params update-tfvars update-tfvars-all
.PHONY: prepare-ecr-image prepare-all-ecr-images

# 基本Terraformコマンド
tf-status:
	@chmod +x scripts/terraform/aws-status.sh
	@scripts/terraform/aws-status.sh $(TF_ENV)

tf-init:
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh init $(TF_ENV)

tf-plan:
	@echo "Terraformプランを作成します（環境: $(TF_ENV), モジュール: $(MODULE)）..."
	@cd deployments/terraform/environments/$(TF_ENV) && \
	terraform init && \
	terraform plan -out=tfplan

tf-apply:
	@echo "Terraformプランを適用します（環境: $(TF_ENV)）..."
	@cd deployments/terraform/environments/$(TF_ENV) && \
	terraform apply -auto-approve tfplan

tf-destroy:
	@chmod +x scripts/terraform/terraform-deploy.sh
	@scripts/terraform/terraform-deploy.sh destroy $(TF_ENV) $(MODULE)

# SSMパラメータ管理
verify-ssm-params:
	@echo "SSMパラメータの存在を確認しています..."
	@if aws ssm get-parameter --name "/${TF_ENV}/database/password" --with-decryption >/dev/null 2>&1; then \
		echo "SSMパラメータは既に存在します"; \
	else \
		echo "SSMパラメータが存在しません。作成します..."; \
		if [ -z "$(TF_VAR_db_password)" ]; then \
			echo "DB_PASSWORDが設定されていません"; \
			read -sp "データベースパスワードを入力してください: " DB_PASS; \
			echo; \
			aws ssm put-parameter --name "/${TF_ENV}/database/password" --type SecureString --value "$$DB_PASS"; \
		else \
			aws ssm put-parameter --name "/${TF_ENV}/database/password" --type SecureString --value "$(TF_VAR_db_password)"; \
		fi; \
		echo "SSMパラメータを作成しました"; \
	fi

# terraform.tfvars更新
update-tfvars:
	@if [ -z "$(SERVICE_TYPE)" ]; then \
		echo -e "${RED}エラー: SERVICE_TYPE環境変数を指定してください${NC}"; \
		echo "例: SERVICE_TYPE=api make update-tfvars"; \
		exit 1; \
	fi
	@echo "terraform.tfvarsを更新しています (サービス: $(SERVICE_TYPE))..."
	@chmod +x scripts/terraform/update-tfvars.sh
	@scripts/terraform/update-tfvars.sh $(TF_ENV) $(SERVICE_TYPE)

update-tfvars-all:
	@echo "すべてのサービスタイプのterraform.tfvarsを更新しています..."
	@chmod +x scripts/terraform/update-tfvars.sh
	@scripts/terraform/update-tfvars.sh $(TF_ENV)

# ECRイメージ準備
prepare-ecr-image:
	@if [ -z "$(SERVICE_TYPE)" ]; then \
		echo -e "${RED}エラー: SERVICE_TYPE環境変数を指定してください${NC}"; \
		echo "例: SERVICE_TYPE=api make prepare-ecr-image"; \
		exit 1; \
	fi
	@chmod +x scripts/docker/prepare-ecr-image.sh
	@scripts/docker/prepare-ecr-image.sh $(SERVICE_TYPE) $(TF_ENV)

prepare-all-ecr-images:
	@echo "全サービスのECRイメージを準備しています..."
	@make prepare-ecr-image SERVICE_TYPE=api TF_ENV=$(TF_ENV)
	@make prepare-ecr-image SERVICE_TYPE=graphql TF_ENV=$(TF_ENV)
	@make prepare-ecr-image SERVICE_TYPE=grpc TF_ENV=$(TF_ENV)
	@echo "全サービスのECRイメージ準備が完了しました"