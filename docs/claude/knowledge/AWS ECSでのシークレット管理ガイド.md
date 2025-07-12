# AWS ECSでのシークレット管理ガイド

## 1. 概要

このガイドでは、AWS ECSでのシークレット管理の実装方法について解説します。特にTerraformを使用したECSタスク定義でのデータベースパスワードなどの機密情報の安全な管理方法に焦点を当てます。

## 2. シークレット管理の選択肢

ECS環境での主要なシークレット管理オプションを比較します：

### 2.1 AWS Systems Manager Parameter Store

**メリット**:
- 低コスト（基本的なパラメータは無料）
- シンプルな管理
- 階層的なパラメータ名構造
- KMS暗号化のサポート

**実装例**:
```hcl
# SSM パラメータの作成
resource "aws_ssm_parameter" "db_password" {
  name        = "/${var.environment}/database/password"
  description = "Database password for ${var.environment} environment"
  type        = "SecureString"
  value       = var.db_password
  
  tags = {
    Environment = var.environment
    Service     = "database"
  }
}

# ECSタスク定義での参照
container_definitions = jsonencode([
  {
    # 他の設定...
    secrets = [
      {
        name      = "DB_PASSWORD"
        valueFrom = aws_ssm_parameter.db_password.arn
      }
    ]
    # 他の設定...
  }
])
```

### 2.2 AWS Secrets Manager

**メリット**:
- 自動ローテーション機能
- データベース認証情報の専用管理
- より高度なアクセス制御
- 多種類のシークレットのサポート

**デメリット**:
- コストが高い（SSMパラメータストアと比較して）

**実装例**:
```hcl
# Secrets Managerシークレットの作成
resource "aws_secretsmanager_secret" "db_password" {
  name = "${var.environment}/database/password"
  
  tags = {
    Environment = var.environment
    Service     = "database"
  }
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = jsonencode({
    password = var.db_password
  })
}

# ECSタスク定義での参照
container_definitions = jsonencode([
  {
    # 他の設定...
    secrets = [
      {
        name      = "DB_PASSWORD"
        valueFrom = "${aws_secretsmanager_secret.db_password.arn}:password::"
      }
    ]
    # 他の設定...
  }
])
```

### 2.3 環境変数（開発環境限定）

**メリット**:
- シンプルな実装
- 追加リソースが不要

**デメリット**:
- セキュリティが低い（平文で保存）
- 本番環境では非推奨

**実装例**:
```hcl
container_definitions = jsonencode([
  {
    # 他の設定...
    environment = [
      { 
        name  = "DB_PASSWORD", 
        value = var.db_password 
      }
    ]
    # 他の設定...
  }
])
```

## 3. IAMアクセス権限の設定

シークレットにアクセスするために必要なIAM権限設定：

### 3.1 Systems Manager Parameter Storeへのアクセス

```hcl
resource "aws_iam_policy" "ssm_parameter_access" {
  name        = "${var.environment}-ssm-parameter-access"
  description = "Allow ECS tasks to access SSM parameters"
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ssm:GetParameters",
          "ssm:GetParameter"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:ssm:${var.region}:*:parameter/${var.environment}/*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "task_exec_ssm_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = aws_iam_policy.ssm_parameter_access.arn
}
```

### 3.2 Secrets Managerへのアクセス

```hcl
resource "aws_iam_policy" "secrets_manager_access" {
  name        = "${var.environment}-secrets-manager-access"
  description = "Allow ECS tasks to access Secrets Manager secrets"
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:secretsmanager:${var.region}:*:secret:/${var.environment}/*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "task_exec_secrets_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = aws_iam_policy.secrets_manager_access.arn
}
```

## 4. Terraformモジュール設計

効率的なシークレット管理のためのモジュール構造：

### 4.1 共有シークレットモジュール

```
modules/
└── shared/
    └── secrets/
        ├── main.tf        # SSMパラメータとIAMポリシー定義
        ├── variables.tf   # 変数定義
        └── outputs.tf     # 出力変数
```

### 4.2 ECSサービスへの統合

```hcl
module "secrets" {
  source = "../../modules/shared/secrets"
  
  environment             = var.environment
  region                  = var.region
  db_password             = var.db_password
  task_execution_role_name = module.shared_ecs_cluster.task_execution_role_name
}

module "service_api" {
  source = "../../modules/service/ecs-service"
  
  # 既存のパラメータ...
  
  # シークレット参照の追加
  db_password_arn = module.secrets.db_password_arn
}
```

## 5. 環境別の考慮事項

### 5.1 開発環境

- より簡易的なセキュリティ要件
- SSMパラメータストアを推奨
- デプロイ前の自動パラメータ作成が便利

### 5.2 本番環境

- より厳格なセキュリティ要件
- Secret Managerの使用を検討
- 自動ローテーションの設定
- 厳密なIAMアクセス制御
- 監査ログの有効化

## 6. デプロイワークフロー

### 6.1 シークレット確認・作成ステップの追加

```makefile
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
```

### 6.2 シークレット検証を含むデプロイコマンド

```makefile
deploy-api-with-params: verify-ssm-params deploy-api
deploy-graphql-with-params: verify-ssm-params deploy-graphql
deploy-grpc-with-params: verify-ssm-params deploy-grpc
deploy-all-with-params: verify-ssm-params deploy-all-services
```

## 7. ベストプラクティス

### 7.1 命名規則

- 環境ごとにプレフィックスを使用
  - 例：`/development/database/password`
  - 例：`/production/database/password`

### 7.2 アクセス制御

- 最小権限の原則に従う
- 細かいリソースARNの指定
- サービスごとに必要な権限のみを付与

### 7.3 暗号化

- SecureStringタイプの使用
- 必要に応じてカスタムKMSキーの使用
- 転送中の暗号化の確保

### 7.4 監視とログ

- パラメータアクセスの監査ログ記録
- 異常アクセスのアラート設定
- 定期的なアクセス権限の見直し

### 7.5 CI/CD統合

- 環境変数の安全な保存
- デプロイパイプラインでのシークレット管理
- 本番シークレットへのアクセス制限

## 8. トラブルシューティング

### 8.1 よくあるエラー

- `The Systems Manager parameter name specified for secret is invalid`
  - 原因: パラメータ名の形式が無効
  - 解決: 完全なARNを指定

- `AccessDenied: is not authorized to perform: ssm:GetParameters`
  - 原因: タスク実行ロールに必要な権限がない
  - 解決: 適切なIAMポリシーを追加

- `ParameterNotFound: Parameter not found`
  - 原因: 参照されたパラメータが存在しない
  - 解決: パラメータの存在を確認し、必要に応じて作成

### 8.2 検証コマンド

```bash
# SSMパラメータの確認
aws ssm get-parameter --name "/${ENVIRONMENT}/database/password" --with-decryption

# ECSタスク定義の確認
aws ecs describe-task-definition --task-definition ${TASK_DEFINITION_NAME}

# IAMポリシーの確認
aws iam get-policy --policy-arn ${POLICY_ARN}
aws iam list-policy-versions --policy-arn ${POLICY_ARN}
```