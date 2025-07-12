# AWS環境構築ガイド：設定管理と環境変数の扱い

## 概要

このドキュメントでは、テストケース管理システムのAWS環境での設定管理と環境変数の扱いについて説明します。ECSタスク定義での環境変数設定から、アプリケーション内での設定読み込みまでの流れを解説します。

## 1. 12-Factor原則に基づく設定管理

### 1.1 設定管理の原則

テストケース管理システムでは、以下の原則に基づいて設定を管理しています：

1. **環境変数による設定**: 環境間での動作の違いを環境変数で管理
2. **優先順位**: 環境変数 > 設定ファイル > デフォルト値
3. **透明性**: 設定値のソースを明示的にログ出力
4. **抽象化**: 設定プロバイダーパターンによる一貫したインターフェース

### 1.2 設定プロバイダーの階層

```
ChainedConfigProvider
  ├── EnvConfigProvider (優先度1)
  ├── FileConfigProvider (優先度2)
  └── DefaultConfigProvider (優先度3)
```

## 2. AWS ECSでの環境変数設定

### 2.1 タスク定義での環境変数設定

```json
{
  "family": "development-test-management-api",
  "containerDefinitions": [
    {
      "name": "api",
      "image": "xxxxxxxxxxxx.dkr.ecr.ap-northeast-1.amazonaws.com/development-test-management-api:latest",
      "environment": [
        {
          "name": "DB_HOST",
          "value": "development-postgres.xxxxxxxxxxxx.ap-northeast-1.rds.amazonaws.com"
        },
        {
          "name": "DB_PORT",
          "value": "5432"
        },
        {
          "name": "DB_NAME",
          "value": "test_management_dev"
        },
        {
          "name": "DB_USERNAME",
          "value": "db_user"
        },
        {
          "name": "DB_SSLMODE",
          "value": "require"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:ssm:ap-northeast-1:xxxxxxxxxxxx:parameter/development/database/password"
        }
      ]
    }
  ]
}
```

### 2.2 AWS Systems Manager Parameter Storeの活用

機密情報はSSM Parameter Storeで管理し、タスク定義の`secrets`セクションで参照します：

```bash
# パラメータの作成
aws ssm put-parameter \
  --name "/development/database/password" \
  --type SecureString \
  --value "your-secure-password"
```

```bash
# タスク実行ロールにパラメータ読み取り権限を付与
aws iam attach-role-policy \
  --role-name ecsTaskExecutionRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess
```

## 3. デプロイとトラブルシューティング

### 3.1 デプロイコマンド

```bash
# APIサービスのデプロイ（環境変数チェック付き）
make deploy-api-with-params TF_ENV=development

# GraphQLサービスのデプロイ（環境変数チェック付き）
make deploy-graphql-with-params TF_ENV=development

# gRPCサービスのデプロイ（環境変数チェック付き）
make deploy-grpc-with-params TF_ENV=development
```

### 3.2 環境変数とパラメータの検証

```bash
# SSMパラメータの存在確認
aws ssm get-parameter --name "/development/database/password" --with-decryption

# タスク定義の環境変数確認
aws ecs describe-task-definition --task-definition development-test-management-api
```

### 3.3 よくある問題と解決策

1. **データベース接続エラー**

   現象: `SYSTEM_ERROR: データベース操作中にエラーが発生しました`
   
   確認ポイント:
   - RDSインスタンスの状態確認
   ```bash
   aws rds describe-db-instances --db-instance-identifier development-postgres
   ```
   - セキュリティグループの設定確認
   ```bash
   aws ec2 describe-security-groups --group-ids <sg-id>
   ```
   - ログの確認
   ```bash
   aws logs get-log-events --log-group-name /aws/ecs/development-test-management-api --log-stream-name <stream>
   ```

2. **環境変数の優先順位問題**

   現象: 期待した設定値が使用されない
   
   解決策:
   - アプリケーションログで設定ソースを確認
   - 環境変数の名前が正確か確認
   - チェーンプロバイダーの優先順位を確認

3. **SSMパラメータアクセス権限問題**

   現象: `Secrets Manager cannot access the specified parameter`
   
   解決策:
   - タスク実行ロールのポリシー確認
   ```bash
   aws iam list-attached-role-policies --role-name ecsTaskExecutionRole
   ```
   - パラメータのARNが正確か確認
   - パラメータのリージョンがタスクと同じか確認

## 4. ベストプラクティス

### 4.1 環境変数命名規則

- データベース接続情報: `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- アプリケーション設定: `APP_ENV`, `LOG_LEVEL`, `API_PORT`, `GRAPHQL_PORT`, `GRPC_PORT`
- セキュリティ設定: `JWT_SECRET`, `JWT_EXPIRATION`

### 4.2 機密情報の管理

1. SSM Parameter Storeの階層構造を活用
   ```
   /environment/category/name
   ```
   例: `/development/database/password`

2. パラメータへのアクセスを最小権限原則に基づいて制限

3. パラメータの上書き保護と監査ログの有効化

### 4.3 マルチ環境対応

各環境（development, staging, production）で統一された環境変数名を使い、値のみを環境に応じて変更します：

```
/development/database/password
/staging/database/password
/production/database/password
```

## 5. 設定管理アーキテクチャの拡張

### 5.1 機能拡張ポイント

1. **AWS Secrets Managerとの統合**
   - より高度なシークレット管理
   - 自動ローテーション機能の活用

2. **環境変数の動的更新**
   - AppConfig/Parameter Storeの変更検知
   - アプリケーションの再起動なしでの設定更新

3. **高度な設定検証**
   - 設定値の整合性検証
   - 依存関係を持つ設定の検証

### 5.2 監視と分析

1. CloudWatchメトリクスとアラートの設定
   - RDS接続エラー率の監視
   - 設定読み込みエラーのアラート

2. X-Rayによるトレース
   - データベース接続の遅延分析
   - 設定読み込みのボトルネック検出

## 6. 結論

設定管理は、クラウド環境でのアプリケーション運用において非常に重要な要素です。12-Factor原則に基づき、環境変数を優先的に使用し、設定ソースを明示的にログ出力することで、トラブルシューティングが容易になります。また、SSM Parameter Storeを活用することで、機密情報を安全に管理できます。

テストケース管理システムでは、設定プロバイダーパターンを採用することで、設定ソースの抽象化と優先順位の明確化を実現しています。この設計により、異なる環境（開発、テスト、本番）で一貫した設定管理が可能になっています。