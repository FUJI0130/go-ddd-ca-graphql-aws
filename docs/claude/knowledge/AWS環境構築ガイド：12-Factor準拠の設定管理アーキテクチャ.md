# AWS環境構築ガイド：12-Factor準拠の設定管理アーキテクチャ

## 1. 背景と目的

このガイドは、AWS環境で運用するアプリケーションにおける設定管理の最適な方法について解説します。12-Factor App原則に準拠し、特にクラウドネイティブな環境で安定した設定管理を実現するためのアーキテクチャ設計と実装方法を提供します。

## 2. 設定管理の基本原則

### 2.1 12-Factor Appにおける設定管理の原則

12-Factor Appの第3原則「設定」では、以下を推奨しています：

- 環境間での変動する設定は環境変数に格納する
- コードと設定を厳密に分離する
- 設定をグループ化したファイルをコードリポジトリにチェックインしない
- 環境変数は、環境ごとに個別に管理・設定する

### 2.2 AWS環境における設定管理のベストプラクティス

1. **環境変数の優先**
   - 環境依存の設定は常に環境変数を優先
   - 設定ファイルは開発環境でのみ使用
   - コンテナ環境では設定ファイルを含めない

2. **環境分離**
   - 開発環境と本番環境で設定管理戦略を分離
   - AWS環境（ECS, Lambda等）では環境変数のみを使用
   - ECSタスク定義やLambda関数定義に直接環境変数を定義

3. **シークレット管理**
   - 機密情報はAWS Secrets ManagerまたはSSM Parameter Storeで管理
   - パラメータストアの値をECSタスク定義の`secrets`セクションで参照
   - IAMロールによるアクセス制御

## 3. 設定管理アーキテクチャの設計

### 3.1 設定プロバイダーパターン

```
ConfigProvider (インターフェース)
  ├── EnvConfigProvider - 環境変数からの設定読み込み（優先度1）
  ├── FileConfigProvider - ファイルからの設定読み込み（優先度2、開発環境のみ）
  └── DefaultConfigProvider - デフォルト値の提供（優先度3）
```

### 3.2 環境検出メカニズム

AWS環境を自動検出するロジック：

```go
// クラウド/AWS環境の検出
isCloudEnv := os.Getenv("APP_ENVIRONMENT") == "production" || 
             os.Getenv("IS_CLOUD_ENV") == "true" || 
             os.Getenv("ECS_CONTAINER_METADATA_URI") != "" ||
             os.Getenv("KUBERNETES_SERVICE_HOST") != ""
```

### 3.3 優先順位付け

```go
// 環境に応じたプロバイダー選択
var providers []ConfigProvider
providers = append(providers, NewEnvConfigProvider("")) // 環境変数は常に最優先

// 本番/クラウド環境では設定ファイルは使用しない
if !isCloudEnv {
    fileProvider, _ := NewFileConfigProvider("./configs", env, "yml")
    if fileProvider != nil {
        providers = append(providers, fileProvider)
    }
}

// デフォルト値は常に最後
providers = append(providers, NewDefaultConfigProvider())

// チェーンプロバイダーの作成
chainedProvider := NewChainedConfigProvider(providers...)
```

## 4. AWS環境での実装パターン

### 4.1 ECSタスク定義での環境変数設定

```json
{
  "containerDefinitions": [
    {
      "name": "app",
      "image": "123456789012.dkr.ecr.region.amazonaws.com/my-app:latest",
      "environment": [
        { "name": "DATABASE_HOST", "value": "db.example.com" },
        { "name": "DATABASE_PORT", "value": "5432" },
        { "name": "APP_ENVIRONMENT", "value": "production" }
      ],
      "secrets": [
        { 
          "name": "DATABASE_PASSWORD", 
          "valueFrom": "arn:aws:ssm:region:123456789012:parameter/app/db/password" 
        }
      ]
    }
  ]
}
```

### 4.2 SSM Parameter Storeとの連携

```bash
# パラメータの作成
aws ssm put-parameter \
  --name "/app/db/password" \
  --type "SecureString" \
  --value "mySecurePassword"

# IAMポリシーでのアクセス制御
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameters"
      ],
      "Resource": "arn:aws:ssm:region:account-id:parameter/app/*"
    }
  ]
}
```

### 4.3 Dockerfileでの設定ファイル除外

```dockerfile
# ビルドステージ
FROM golang:1.19 as builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/api/main.go

# 実行ステージ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server /app/server

# 設定ファイルを含めない
# COPY --from=builder /app/configs /app/configs

# 既存の設定ファイルが残っている可能性があるため明示的に削除
RUN rm -rf /app/configs

ENTRYPOINT ["/app/server"]
```

## 5. トラブルシューティング

### 5.1 一般的な問題と解決策

| 問題 | 原因 | 解決策 |
|------|------|-------|
| 環境変数が反映されない | 優先順位の問題 | 設定プロバイダーの順序確認、ログ出力での診断 |
| コンテナで設定ファイルが優先される | コンテナにファイルが含まれている | Dockerfileの見直し、`rm -rf /app/configs`の追加 |
| シークレットへのアクセスエラー | IAMアクセス権限の問題 | IAMロールとポリシーの確認、権限付与 |
| 本番と開発で動作が異なる | 環境判別ロジックの不備 | 環境検出メカニズムの強化、ログ出力の改善 |

### 5.2 設定のデバッグ方法

```go
// 設定値とそのソースを出力
log.Printf("Database.Host: %s (source: %s)", 
          config.Database.Host, 
          getSettingSource(chainedProvider, "database.host"))
```

デバッグモードを有効化：
```bash
# 開発環境
export DEBUG=true

# ECS環境
# タスク定義に環境変数を追加
{ "name": "DEBUG", "value": "true" }
```

### 5.3 環境検出のテスト

異なる環境を模擬するためのテクニック：

```bash
# ECS環境を模擬
export ECS_CONTAINER_METADATA_URI="http://169.254.170.2/v3"
export IS_CLOUD_ENV="true"

# 開発環境（デフォルト）
unset ECS_CONTAINER_METADATA_URI
unset IS_CLOUD_ENV
```

## 6. ベストプラクティス

1. **明示的な設定値ソースのログ出力**
   - 設定値がどこから来たのか追跡可能にする
   - 環境変数、設定ファイル、デフォルト値のいずれかを明記

2. **包括的なユニットテスト**
   - 異なる環境での設定読み込みをテスト
   - 優先順位が正しく機能することを検証

3. **デフォルト値の慎重な設定**
   - ローカル開発に適した値をデフォルトに
   - 本番環境では必ず環境変数で上書き

4. **設定キーの標準化**
   - 一貫した命名規則（ドット区切りまたはアンダースコア区切り）
   - 環境変数への明確なマッピング

5. **サンプル設定ファイルの提供**
   - `configs/sample.yml`などでテンプレートを提供
   - 実際の設定値は含めない、プレースホルダーのみ

## 7. まとめ

12-Factor原則に基づく設定管理は、AWS環境でのアプリケーション運用において特に重要です。環境変数を優先し、開発環境と本番環境を明確に分離することで、柔軟かつ安全なシステムを構築できます。設定プロバイダーパターンと環境検出メカニズムを組み合わせることで、様々な環境で一貫した動作を保証できます。