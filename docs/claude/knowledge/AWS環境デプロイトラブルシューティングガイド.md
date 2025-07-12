# AWS環境デプロイトラブルシューティングガイド

## 1. 問題の概要

テストケース管理システムのAWS ECSデプロイで、環境変数が設定ファイルよりも優先されず、データベース接続が失敗している問題が発生しています。

**現在のエラー:**
```
設定ファイル /app/configs/development.yml を読み込みました
Database settings: user=testuser, password=testpass, host=localhost, port=5432, dbname=test_management
Failed to initialize database: dial tcp 127.0.0.1:5432: connect: connection refused
```

**原因:**
Dockerfileで設定ファイルがイメージにコピーされており、環境変数より設定ファイルが優先されている。

## 2. 解決手順

### 2.1 Dockerfileの修正

Dockerfileから設定ファイルのコピー行を削除するだけでなく、明示的に設定ファイルディレクトリを削除します：

```bash
# Dockerfileの設定ファイルコピー行をコメントアウト
# COPY --from=builder /app/configs /app/configs

# 設定ファイルディレクトリを明示的に削除
RUN rm -rf /app/configs

# 変更を確認
cat Dockerfile | grep -A 3 -B 3 "バイナリコピー"

# 変更をコミット
git add Dockerfile
git commit -m "fix: Remove config files from Docker image to prioritize environment variables"
```

### 2.2 新しいイメージのビルドとプッシュ

```bash
# Makefileのコマンドを使用して新しいイメージをビルドしてプッシュ
make prepare-ecr-image SERVICE_TYPE=api TF_ENV=development

# ECRのイメージを確認
aws ecr describe-images --repository-name development-test-management-api --query 'imageDetails[*].{Tags:imageTags,PushedAt:imagePushedAt}' --output table
```

### 2.3 サービスの再デプロイ

APIサービスを強制再デプロイで更新します：

```bash
# サービスの強制再デプロイ
aws ecs update-service --cluster development-shared-cluster --service development-api --force-new-deployment

# デプロイ状況の確認
aws ecs describe-services --cluster development-shared-cluster --services development-api --query 'services[0].{Status:status,RunningCount:runningCount,DesiredCount:desiredCount,TaskDefinition:taskDefinition}'
```

### 2.4 ログの確認

```bash
# サービスが起動するまで待機（60秒）
sleep 60

# 最新のログストリームを取得
LOG_GROUP="/ecs/development-api"
LOG_STREAM=$(aws logs describe-log-streams --log-group-name $LOG_GROUP --order-by LastEventTime --descending --limit 1 --query 'logStreams[0].logStreamName' --output text)
echo "最新のログストリーム: $LOG_STREAM"

# ログを確認
aws logs get-log-events --log-group-name $LOG_GROUP --log-stream-name $LOG_STREAM --limit 30
```

## 3. 確認ポイント

ログを確認し、以下を確認します：

1. **設定ファイル読み込みメッセージの変化**: 
   - 設定ファイルが見つからないという警告メッセージが表示されるはず
   ```
   警告: development.yml 設定ファイルが見つかりません。環境変数またはデフォルト値を使用します
   ```

2. **正しいDB接続情報の使用**: 
   - 環境変数から取得した正しいRDSのホスト名が表示されるはず
   ```
   Database connection settings from chain(environment > config_struct): host=development-postgres....
   ```

3. **データベース接続の成功**: 
   - 接続成功メッセージが表示されるはず
   ```
   Successfully connected to database
   ```

## 4. 問題の根本原因

この問題の根本原因は、設定管理アーキテクチャと12-Factor原則の適用に関連しています。

1. **設定ファイル探索ロジック**:
   `config.go` 内のコードが複数の場所から設定ファイルを探索します：
   ```go
   fileConfigPaths := []string{configPath, "./configs", "."}
   ```
   これにより、コンテナ内に設定ファイルが存在すると、環境変数よりも優先される場合があります。

2. **Dockerイメージ構成**:
   マルチステージビルドでは、最初のステージでコードの全体をコピーしますが、最終イメージでは必要なファイルのみが含まれるべきです。設定ファイルのような環境依存のファイルは除外すべきです。

3. **環境変数と設定ファイルの優先順位**:
   12-Factor原則では環境変数を優先するべきですが、実装によっては設定ファイルが優先される場合があります。

## 5. 予防策と推奨事項

### 5.1 Dockerfileのベストプラクティス

1. **環境依存ファイルを除外**:
   ```dockerfile
   # 必要なファイルだけをコピー
   COPY --from=builder /app/server /app/server
   
   # 設定ファイルを明示的に除外
   RUN rm -rf /app/configs
   ```

2. **ビルド引数と環境変数の明確な区別**:
   ```dockerfile
   # ビルド時の引数
   ARG SERVICE_TYPE=api
   
   # 実行時の環境変数
   ENV APP_ENVIRONMENT=production
   ```

### 5.2 設定管理のベストプラクティス

1. **環境変数が確実に優先されるようにする**:
   ```go
   // 環境変数が存在する場合は必ず優先する
   if envValue, exists := os.LookupEnv("DB_HOST"); exists {
       return envValue
   }
   ```

2. **クラウド環境と開発環境で異なる動作**:
   ```go
   // クラウド環境では設定ファイルを使用しない
   if os.Getenv("CLOUD_ENV") == "true" {
       return NewEnvOnlyConfigProvider()
   }
   ```

3. **設定ソースの透明性**:
   ```go
   // 設定値のソースを常にログ出力
   log.Printf("Using configuration from %s: %s=%s", source, key, value)
   ```

### 5.3 ECS環境構成のベストプラクティス

1. **タスク定義での環境変数セット**:
   - 本番環境固有の値をタスク定義に設定
   - 機密情報はSSMパラメータストアや環境変数から提供

2. **ヘルスチェックの設定**:
   - アプリケーションの起動が完了したことを確認するヘルスチェック
   - 設定問題を早期に検出できるチェックポイント

3. **ログ監視の強化**:
   - 設定読み込みに関するログの収集と監視
   - 環境変数と設定ファイルの両方が使用された場合の警告

## 6. トラブルシューティングチェックリスト

設定関連の問題が発生した場合のチェックリスト：

1. **ログの確認**:
   - 設定ファイルの読み込みに関するメッセージを確認
   - どの設定ソースが使用されているかを確認

2. **コンテナ内の確認**:
   - コンテナ内に設定ファイルが存在するか確認
   ```bash
   aws ecs execute-command --cluster your-cluster --task your-task-id --container your-container --command "ls -la /app/configs"
   ```

3. **環境変数の確認**:
   - タスク定義で正しい環境変数が設定されているか確認
   ```bash
   aws ecs describe-task-definition --task-definition your-task-def --query 'taskDefinition.containerDefinitions[0].environment'
   ```

4. **設定読み込みロジックの確認**:
   - 設定管理コードで優先順位が正しく設定されているか確認
   - 環境変数が正しく評価されているか確認

5. **イメージビルドの確認**:
   - 最新のDockerfileでイメージがビルドされているか確認
   - ECRのイメージタグとタイムスタンプを確認