# GraphQLテスト手順ガイド（2025年3月25日更新）

このドキュメントではテストケース管理システムのGraphQLインターフェースをテストするための手順と環境構築について説明します。テストの実行方法、自動化、一般的な問題のトラブルシューティングを含みます。

## 1. テスト環境の概要

テストケース管理システムのGraphQLインターフェースには以下の種類のテストが用意されています：

- **ユニットテスト**: リゾルバーの個別機能をテスト
- **統合テスト**: データベースを含む実際のシステム全体をテスト

## 2. 手動テスト環境のセットアップ

### 2.1 データベースの起動

まず、PostgreSQLデータベースコンテナを起動します：

```bash
make db-up
```

これにより、以下の設定でデータベースが起動します：
- ホスト: localhost
- ポート: 5432
- ユーザー: testuser
- パスワード: testpass
- データベース名: test_management

### 2.2 マイグレーションの実行

次に、必要なテーブルとシーケンスを作成するためにマイグレーションを実行します：

```bash
make migrate
```

これにより、スキーマ、テーブル、シーケンスなどのデータベースオブジェクトが作成されます。

### 2.3 GraphQLサーバーの起動

次に、GraphQLサーバーを起動します：

```bash
make run-graphql
# または
go run cmd/graphql/main.go
```

サーバーが起動したら、ブラウザで以下のURLにアクセスしてGraphQL Playgroundを開きます：

```
http://localhost:8080
```

## 3. 自動テストの実行

### 3.1 リゾルバーのユニットテスト

リゾルバーの基本的な機能をテストする場合は、以下のコマンドを実行します：

```bash
make test-graphql-resolver
```

このテストではモック化されたユースケースを使用し、データベースとの接続は行いません。

### 3.2 GraphQL統合テスト

データベースを含むエンドツーエンドの統合テストを実行する場合は、以下のコマンドを実行します：

```bash
make test-graphql
```

このコマンドは以下の処理を自動的に実行します：
1. テスト用PostgreSQLコンテナの起動（ポート5433）
2. データベースマイグレーションの実行
3. シーケンスのリセット
4. 統合テストの実行
5. テスト終了後のコンテナ停止

> **注意**: 統合テストを実行する前に、他のテストやアプリケーションがテスト用ポート（5433）を使用していないことを確認してください。

## 4. 主要なGraphQLクエリとミューテーション

### 4.1 テストスイート一覧の取得

```graphql
query GetTestSuites {
  testSuites {
    edges {
      node {
        id
        name
        status
        progress
        estimatedStartDate
        estimatedEndDate
      }
    }
    totalCount
    pageInfo {
      hasNextPage
      hasPreviousPage
      startCursor
      endCursor
    }
  }
}
```

### 4.2 ステータスによるフィルタリング

```graphql
query GetTestSuitesByStatus($status: SuiteStatus) {
  testSuites(status: $status) {
    edges {
      node {
        id
        name
        status
      }
    }
    totalCount
  }
}
```

変数:
```json
{
  "status": "IN_PROGRESS"
}
```

### 4.3 ページネーション

```graphql
query GetTestSuitesPaginated($page: Int, $pageSize: Int) {
  testSuites(page: $page, pageSize: $pageSize) {
    edges {
      node {
        id
        name
      }
    }
    totalCount
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
  }
}
```

変数:
```json
{
  "page": 1,
  "pageSize": 2
}
```

### 4.4 単一テストスイートの取得

```graphql
query GetTestSuite($id: ID!) {
  testSuite(id: $id) {
    id
    name
    description
    status
    estimatedStartDate
    estimatedEndDate
    requireEffortComment
    progress
    createdAt
    updatedAt
  }
}
```

変数:
```json
{
  "id": "TS001-202501"
}
```

### 4.5 リレーションを含むテストスイートの取得

```graphql
query GetTestSuiteWithRelations($id: ID!) {
  testSuite(id: $id) {
    id
    name
    status
    groups {
      id
      name
      displayOrder
      cases {
        id
        title
        status
        priority
      }
    }
  }
}
```

変数:
```json
{
  "id": "TS001-202501"
}
```

### 4.6 テストスイートの作成

```graphql
mutation CreateTestSuite($input: CreateTestSuiteInput!) {
  createTestSuite(input: $input) {
    id
    name
    description
    status
    estimatedStartDate
    estimatedEndDate
    requireEffortComment
  }
}
```

変数:
```json
{
  "input": {
    "name": "新規テストスイート",
    "description": "GraphQLテスト用スイート",
    "estimatedStartDate": "2025-04-01T00:00:00Z",
    "estimatedEndDate": "2025-04-30T00:00:00Z",
    "requireEffortComment": true
  }
}
```

### 4.7 テストスイートの更新

```graphql
mutation UpdateTestSuite($id: ID!, $input: UpdateTestSuiteInput!) {
  updateTestSuite(id: $id, input: $input) {
    id
    name
    description
    estimatedEndDate
  }
}
```

変数:
```json
{
  "id": "作成時に取得したID",
  "input": {
    "description": "更新後の説明文",
    "estimatedEndDate": "2025-05-15T00:00:00Z"
  }
}
```

### 4.8 ステータスの更新

```graphql
mutation UpdateStatus($id: ID!, $status: SuiteStatus!) {
  updateTestSuiteStatus(id: $id, status: $status) {
    id
    status
  }
}
```

変数:
```json
{
  "id": "作成時に取得したID",
  "status": "IN_PROGRESS"
}
```

## 5. テスト観点とチェックポイント

### 5.1 基本機能のテスト

- ✅ テストスイート一覧が正しく取得できるか
- ✅ 単一テストスイートが正しく取得できるか
- ✅ テストスイートの作成が成功するか
- ✅ テストスイートの更新が正しく反映されるか
- ✅ ステータス更新が正しく機能するか

### 5.2 リレーションのテスト

- ✅ TestSuiteからGroupsが正しく取得できるか
- ✅ TestGroupからCasesが正しく取得できるか
- ✅ リレーションデータの件数や内容は正確か

### 5.3 フィルタリングとページネーション

- ✅ ステータスによるフィルタリングが正しく機能するか
- ✅ ページネーション（page, pageSize）が正しく機能するか
- ✅ pageInfoの値（hasNextPage, hasPreviousPage）が正確か

### 5.4 エラーハンドリング

- ✅ 存在しないIDを指定した場合のエラー
- ✅ バリデーションエラー（日付の前後関係など）
- ✅ 不正なステータス値を指定した場合のエラー

## 6. 自動テストの設定と拡張

### 6.1 テストコードの構成

GraphQLテストコードは以下のディレクトリに配置されています：

```
internal/interface/graphql/
├── integration_test.go    // 統合テスト
├── resolver/
│   ├── resolver_test.go   // リゾルバーのユニットテスト
```

### 6.2 新しいテストケースの追加

新しいテストケースを追加する場合は、以下の手順で行います：

1. ユニットテストの場合：`resolver_test.go`に追加
   - モックを使用した軽量なテスト
   - 入力バリデーションや変換ロジックのテストに最適

2. 統合テストの場合：`integration_test.go`に追加
   - リレーションや実際のデータベースの動作を含むテスト
   - エンドツーエンドの動作確認に最適

### 6.3 テスト用DBの構成とマイグレーション

統合テストでは、テスト専用のPostgreSQLコンテナが以下の設定で使用されます：

- ホスト: localhost
- ポート: 5433 （通常の開発DB 5432と区別）
- ユーザー: test_user
- パスワード: test_pass
- データベース名: test_db

テストデータベースに必要なスキーマとシーケンスは、起動時に自動的にマイグレーションされます。

## 7. トラブルシューティング

### 7.1 データベース接続エラー

エラーメッセージ: `データベース操作中にエラーが発生しました`

確認事項:
- テスト用DBが起動しているか確認（`docker ps | grep test_db`）
- 接続設定が正しいか確認（ポート、ユーザー名、パスワード）
- ファイアウォールやネットワーク設定で接続がブロックされていないか確認

解決方法:
```bash
# コンテナの状態確認
docker ps | grep test_db

# コンテナの再起動
make test-graphql
```

### 7.2 シーケンスエラー

エラーメッセージ: `failed to generate sequence number: pq: relation "test_group_seq" does not exist`

確認事項:
- マイグレーションが正しく実行されているか確認
- シーケンスが存在するか確認

解決方法:
```bash
# マイグレーションの明示的な実行
migrate -path scripts/migrations -database "postgresql://test_user:test_pass@localhost:5433/test_db?sslmode=disable" up

# シーケンス作成スクリプトの直接実行
PGPASSWORD=test_pass psql -h localhost -p 5433 -U test_user -d test_db -f scripts/setup/create_test_sequences.sql
```

### 7.3 ID競合エラー

エラーメッセージ: `CONFLICT: TestSuite (ID: TSxxx-yyyymm) は既に存在しています`

確認事項:
- テスト実行前にデータベースがクリーンアップされているか確認
- 同じ日に複数回テストを実行している場合、IDが重複する可能性がある

解決方法:
```bash
# テスト用DBの完全リセット
docker compose -f test/integration/postgres/docker-compose.test.yml down -v
docker compose -f test/integration/postgres/docker-compose.test.yml up -d
migrate -path scripts/migrations -database "postgresql://test_user:test_pass@localhost:5433/test_db?sslmode=disable" up
```

### 7.4 NILポインタ参照エラー

エラーメッセージ: `runtime error: invalid memory address or nil pointer dereference`

確認事項:
- PageとPageSizeパラメータがNULLの場合の処理
- 返される配列やオブジェクトがNULLでないか確認

解決方法:
- リゾルバーコードにNULLチェックを追加
- テストコードでNULLになりうる変数の初期化を確認

```go
// NULLチェックを行う例
if page != nil {
    params.Page = page
} else {
    defaultPage := 1
    params.Page = &defaultPage
}
```

### 7.5 テスト間でのID共有の問題

エラーメッセージ: `TEST_SUITE_NOT_FOUND: ID  のテストスイートが見つかりません`

確認事項:
- 子テストへのID引き渡しが正しく行われているか
- 前のテストのエラーが後続のテストに影響していないか

解決方法:
- テスト間の依存関係を減らす
- 空IDチェックを追加

```go
// 空IDチェックの例
if newID != "" {
    // テストを実行
} else {
    t.Skip("テストスイートIDが取得できなかったためテストをスキップします")
}
```

## 8. ベストプラクティス

### 8.1 テストデータの独立性

各テストケースは独立して実行できるように設計されるべきです：

- テスト間でデータが干渉しないようにする
- 各テストで必要なデータを明示的に作成する
- テスト終了後にデータをクリーンアップする

```go
// テストデータの準備
createdID, err := seedTestData(ctx)
require.NoError(t, err)

// テスト終了後のクリーンアップ
defer cleanupTestData(ctx, createdID)
```

### 8.2 テスト用シーケンスの管理

テスト間でIDが重複しないようにするためのアプローチ：

1. 各テスト実行前にシーケンスをリセットする
2. テスト専用のID生成ロジックを使用する
3. タイムスタンプベースの一意識別子を使用する（例：`TEST-{timestamp}-TS001`）

```go
// タイムスタンプベースの一意識別子の例
var testPrefix = fmt.Sprintf("TEST-%d-", time.Now().UnixNano())

// テストデータ作成時に使用
testName := fmt.Sprintf("%s統合テスト用スイート", testPrefix)
```

### 8.3 効率的なテスト実行

テストの実行時間を短縮するためのヒント：

- テスト用DBコンテナの再利用（毎回作り直さない）
- ユニットテストと統合テストを分離して実行
- テスト専用の軽量なDBセットアップの使用

### 8.4 テストコードの保守性

テストコードの保守性を高めるためのヒント：

- デバッグ用のログ出力を追加する
- テスト用のヘルパー関数を作成する
- テスト終了後のクリーンアップを確実に行う
- エラーメッセージを具体的にする

```go
// デバッグログの例
fmt.Printf("作成したテストスイートID: %s\n", createdTestSuiteID)

// エラーメッセージを具体的にする例
assert.NoError(t, err, "テストスイート作成中にエラーが発生しました: %v", err)
```

## 9. 今後の拡張予定

1. **テスト環境の改善**
   - IDジェネレーターをテスト用に拡張（ランダム成分や環境検出機能の追加）
   - テストデータクリーンアップ処理の強化
   - テスト間の依存関係軽減

2. **DataLoaderの実装と検証**
   - N+1問題対策のDataLoader実装
   - パフォーマンステスト追加

3. **サブスクリプションテスト**
   - リアルタイム更新の検証
   - WebSocketベースのテスト

4. **認証・認可のテスト**
   - ユーザーロールごとの動作検証
   - アクセス制限のテスト

5. **スキーマのバージョニングテスト**
   - 後方互換性の検証
   - マイグレーションテスト