# AWS環境構築ガイド：設定管理ベストプラクティス

## 1. 概要

本ガイドでは、テストケース管理システムのAWS環境における設定管理のベストプラクティスについて解説します。特に12-Factor原則に準拠した環境変数優先の設定管理アーキテクチャと、AWSの各サービスとの統合方法に焦点を当てています。

## 2. 12-Factor準拠の設定管理

### 2.1 基本原則

12-Factor Appの第3原則「設定」では、以下の原則を推奨しています：

- 環境ごとに異なる設定を環境変数に保存する
- コードと設定を明確に分離する
- 設定グループを環境ごとにコードリポジトリで管理しない

これらの原則に従い、以下の設計を実装しました：

1. **環境変数優先の階層構造**:
   - 環境変数 > 設定ファイル > デフォルト値
   - クラウド環境では設定ファイルを完全に無視

2. **クラウド環境の自動検出**:
   ```go
   isCloudEnv := os.Getenv("APP_ENVIRONMENT") == "production" ||
     os.Getenv("IS_CLOUD_ENV") == "true" ||
     os.Getenv("ECS_CONTAINER_METADATA_URI") != "" ||
     os.Getenv("KUBERNETES_SERVICE_HOST") != ""
   ```

3. **設定ソースの透明性**:
   ```go
   log.Printf("Database.Host: %s (source: %s)", config.Database.Host, 
     getSettingSource(chainedProvider, "database.host", "DB_HOST"))
   ```

### 2.2 AWS環境での実装

AWS環境では以下の実装パターンを採用しています：

1. **ECSタスク定義での環境変数設定**:
   ```json
   "environment": [
     {"name": "DB_HOST", "value": "development-postgres.xxxx.ap-northeast-1.rds.amazonaws.com"},
     {"name": "DB_PORT", "value": "5432"},
     {"name": "APP_ENVIRONMENT", "value": "production"}
   ]
   ```

2. **Secretsの使用**:
   ```json
   "secrets": [
     {"name": "DB_PASSWORD", "valueFrom": "arn:aws:ssm:region:account:parameter/development/database/password"}
   ]
   ```

3. **コンテナイメージの最適化**:
   - Dockerfileから設定ファイルを含めない
   - 明示的に設定ファイルディレクトリを削除
   ```dockerfile
   # 設定ファイルコピー行の削除
   # COPY --from=builder /app/configs /app/configs
   
   # 明示的に設定ファイルディレクトリを削除
   RUN rm -rf /app/configs
   ```

## 3. AWS Systems Manager Parameter Storeの活用

### 3.1 パラメータストアの設計

Parameter Store（SSM）を使用して機密情報を管理する際の設計パターン：

1. **階層的パラメータ名**:
   ```
   /environment/category/name
   ```
   例: `/development/database/password`

2. **パラメータタイプ**:
   - 標準文字列: 通常の設定値
   - セキュア文字列: 機密情報（パスワードなど）

3. **アクセス制御**:
   - IAMポリシーによるきめ細かいアクセス制御
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": ["ssm:GetParameters"],
         "Resource": "arn:aws:ssm:region:account:parameter/development/*"
       }
     ]
   }
   ```

### 3.2 パラメータの作成と管理

```bash
# セキュア文字列パラメータの作成
aws ssm put-parameter \
  --name "/development/database/password" \
  --type "SecureString" \
  --value "yourSecurePassword"

# パラメータの取得（復号化）
aws ssm get-parameter \
  --name "/development/database/password" \
  --with-decryption

# パラメータの更新
aws ssm put-parameter \
  --name "/development/database/password" \
  --type "SecureString" \
  --value "newPassword" \
  --overwrite
```

### 3.3 ECSタスク定義での参照

ECSタスク定義で、環境変数としてSSMパラメータを参照する方法：

```json
"containerDefinitions": [
  {
    "name": "app",
    "secrets": [
      {
        "name": "DB_PASSWORD",
        "valueFrom": "arn:aws:ssm:ap-northeast-1:xxxxxxxxxxxx:parameter/development/database/password"
      }
    ]
  }
]
```

この設定により、コンテナ起動時にパラメータストアから値が取得され、環境変数として設定されます。

## 4. 環境別設定管理

### 4.1 環境分離

異なる環境（開発、ステージング、本番）で設定を分離する方法：

1. **環境変数プレフィックス**:
   - `APP_ENVIRONMENT=development`
   - `APP_ENVIRONMENT=staging`
   - `APP_ENVIRONMENT=production`

2. **SSMパラメータの階層化**:
   - `/development/database/password`
   - `/staging/database/password`
   - `/production/database/password`

3. **Terraformの環境変数**:
   ```hcl
   variable "environment" {
     description = "Environment (development, staging, production)"
     type        = string
   }
   ```

### 4.2 terraform.tfvarsによる環境設定

```hcl
# development/terraform.tfvars
environment = "development"
app_environment = "development"

# production/terraform.tfvars
environment = "production"
app_environment = "production"
```

### 4.3 Makefileによる環境指定

```makefile
# 環境を指定してデプロイ
deploy-api:
	@echo "Deploying API service to $(TF_ENV) environment..."
	terraform -chdir=$(TF_DIR) apply -auto-approve -target=module.ecs_api

# 環境変数チェック付きデプロイ
deploy-api-with-params: verify-ssm-params deploy-api
```

## 5. クラウド環境検出メカニズム

### 5.1 環境変数による検出

アプリケーションがクラウド環境で動作しているかを検出する方法：

```go
isCloudEnv := os.Getenv("APP_ENVIRONMENT") == "production" ||
  os.Getenv("IS_CLOUD_ENV") == "true" ||
  os.Getenv("ECS_CONTAINER_METADATA_URI") != "" ||
  os.Getenv("KUBERNETES_SERVICE_HOST") != ""
```

このロジックにより、以下の環境を自動検出します：
- 本番環境（`APP_ENVIRONMENT=production`）
- クラウド環境（`IS_CLOUD_ENV=true`）
- ECS環境（`ECS_CONTAINER_METADATA_URI`が設定されている場合）
- Kubernetes環境（`KUBERNETES_SERVICE_HOST`が設定されている場合）

### 5.2 検出結果による分岐

検出結果に基づいて設定読み込み方法を分岐させます：

```go
if !isCloudEnv {
  // 設定ファイルプロバイダーを追加（開発環境のみ）
  // ...
} else {
  log.Printf("クラウド環境で実行中のため、設定ファイルは使用しません")
}
```

## 6. ヘルスチェックと運用監視

### 6.1 ヘルスチェック設計

異なるサービスタイプに対応したヘルスチェック：

1. **Dockerfileでのヘルスチェック**:
   ```dockerfile
   HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
     CMD wget -qO- http://localhost:${PORT}/health || exit 1
   ```

2. **ALBターゲットグループのヘルスチェック**:
   ```hcl
   health_check {
     path                = "/health"
     protocol            = "HTTP"
     port                = "traffic-port"
     healthy_threshold   = 3
     unhealthy_threshold = 3
     timeout             = 5
     interval            = 30
   }
   ```

3. **サービス特性に応じたヘルスチェック**:
   - REST API: `/health` エンドポイント
   - GraphQL: `/` または専用エンドポイント
   - gRPC: gRPCヘルスチェックプロトコル

### 6.2 CloudWatch監視

設定関連の問題を監視するための設定：

1. **ログフィルター**:
   ```
   filter = "設定ファイル" | "Database connection" | "Error"
   ```

2. **メトリクスフィルター**:
   ```
   filterPattern = "{ $.message = \"*データベース接続エラー*\" }"
   metricName = "DatabaseConnectionErrors"
   ```

3. **アラーム設定**:
   ```
   alarmName = "DatabaseConnectionFailures"
   comparisonOperator = "GreaterThanThreshold"
   threshold = 0
   evaluationPeriods = 1
   ```

## 7. 発展的な設定管理

### 7.1 AWS AppConfigの活用

動的な設定管理のためのAWS AppConfigの利用：

1. **設定プロファイルの作成**:
   ```bash
   aws appconfig create-configuration-profile \
     --application-id 1234abcd \
     --name "APIConfig" \
     --location-uri "hosted" \
     --type "AWS.Freeform"
   ```

2. **設定データの保存**:
   ```bash
   aws appconfig create-hosted-configuration-version \
     --application-id 1234abcd \
     --configuration-profile-id 5678efgh \
     --content file://config.json \
     --content-type "application/json"
   ```

3. **アプリケーションからの取得**:
   ```go
   // AWS AppConfig SDKを使用して設定を取得
   ```

### 7.2 設定の動的更新

実行中のアプリケーションの設定を動的に更新するための方法：

1. **定期的なポーリング**:
   ```go
   // 定期的にSSMパラメータストアやAppConfigから設定を再読み込み
   ```

2. **トリガーベースの更新**:
   ```go
   // SNSトピックをサブスクライブして設定変更通知を受信
   ```

3. **グレースフル設定更新**:
   ```go
   // 既存接続には古い設定を使用し、新しい接続には新しい設定を使用
   ```

## 8. トラブルシューティング

### 8.1 設定関連の一般的な問題

1. **環境変数が検出されない**:
   - ECSタスク定義の環境変数セクションを確認
   - 大文字/小文字やフォーマットが正しいかを確認
   - ログでどの設定ソースが使用されているかを確認

2. **Parameter Storeからの取得エラー**:
   - タスク実行ロールのIAMポリシーを確認
   - パラメータ名とARNが正確かを確認
   - CloudWatchログでエラーメッセージを確認

3. **データベース接続エラー**:
   - 実際に使用されている接続情報をログから確認
   - セキュリティグループの設定を確認
   - RDSインスタンスの状態を確認

### 8.2 確認コマンド

```bash
# 現在のECSタスク定義を確認
aws ecs describe-task-definition --task-definition your-task-def

# SSMパラメータの内容を確認
aws ssm get-parameter --name "/development/database/password" --with-decryption

# CloudWatchログを確認
aws logs get-log-events --log-group-name /ecs/your-service --log-stream-name your-stream
```

## 9. ベストプラクティスのまとめ

1. **設定の層別化**:
   - 基本設定: デフォルト値（コード内）
   - 環境設定: 環境変数（ECSタスク定義）
   - 機密情報: SSMパラメータストア（セキュア文字列）

2. **透明性の確保**:
   - 設定ソースを明示的にログに出力
   - デバッグモードでより詳細な情報を表示
   - 設定変更履歴を追跡

3. **環境特性に応じた振る舞い**:
   - 開発環境: より柔軟な設定方法（ファイル + 環境変数）
   - クラウド環境: 厳格な設定方法（環境変数のみ）
   - テスト環境: モック可能な設定プロバイダー

4. **セキュリティの確保**:
   - 機密情報は常に暗号化
   - 最小権限の原則を徹底
   - 設定アクセスのログ記録と監査