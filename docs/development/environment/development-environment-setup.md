# テストケース管理システム - 開発環境セットアップガイド

このドキュメントでは、テストケース管理システムの開発環境のセットアップ方法および各種コマンドの使用方法について説明します。

## 目次

1. [必要条件](#必要条件)
2. [環境セットアップ](#環境セットアップ)
3. [データベース操作](#データベース操作)
4. [APIサーバー起動方法](#APIサーバー起動方法)
5. [テスト実行方法](#テスト実行方法)
6. [コード生成](#コード生成)
7. [開発ワークフロー](#開発ワークフロー)
8. [トラブルシューティング](#トラブルシューティング)

## 必要条件

以下のツールがインストールされていることを確認してください：

- Go 1.21以上
- Docker 27.3.1以上
- Docker Compose v2.29.7以上
- Protocol Buffers compiler (protoc) v3.21.12以上
- PostgreSQL 14.13（Docker内で実行）
- Git

## 環境セットアップ

### 1. リポジトリのクローン

```bash
git clone https://github.com/FUJI0130/go-ddd-ca.git
cd go-ddd-ca
```

### 2. 依存関係のインストール

```bash
go mod download
go mod tidy
```

### 3. Protocol Buffersコンパイラプラグインのインストール

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
```

### 4. GraphQL関連ツールのインストール

```bash
go get github.com/99designs/gqlgen@v0.17.67
go get golang.org/x/tools/go/packages@latest
go get golang.org/x/tools/go/ast/astutil@latest
go get golang.org/x/tools/imports@latest
go get github.com/urfave/cli/v2@latest
go get github.com/lib/pq
```

## データベース操作

### データベースコンテナの起動

```bash
make db-up
```

これにより、以下の設定でPostgreSQLコンテナが起動します：
- ホスト: localhost
- ポート: 5432
- ユーザー: testuser
- パスワード: testpass
- データベース: test_management

### データベースコンテナの停止

```bash
make db-down
```

### マイグレーションの実行（スキーマ作成）

```bash
make migrate
```

### マイグレーションのロールバック

```bash
make migrate-down
```

### テスト用データベースの起動

```bash
make test-integration
```

## APIサーバー起動方法

### REST APIサーバーの起動

```bash
make run
# または
go run cmd/api/main.go
```

デフォルトではポート8080で起動します。

### GraphQLサーバーの起動

```bash
go run cmd/graphql/main.go
```

デフォルトではポート8080で起動します。
GraphQL Playgroundには http://localhost:8080 でアクセスできます。

### gRPCサーバーの起動

```bash
go run cmd/grpc/main.go
```

デフォルトではポート50051で起動します。

## テスト実行方法

### 単体テストの実行

```bash
make test
# または
go test ./...
```

### 統合テストの実行

```bash
make test-integration
```

これにより、テスト用PostgreSQLコンテナが起動し、テスト実行後にコンテナが停止します。

## コード生成

### Protocol Buffersコードの生成

```bash
make proto
# または
protoc --proto_path=. \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/testsuite/v1/*.proto
```

### GraphQLコードの生成

```bash
go run github.com/99designs/gqlgen generate
```

## 開発ワークフロー

### 1. データベース準備

開発前にデータベースが起動していることを確認します：

```bash
# データベースの状態確認
docker ps | grep postgres

# もし起動していなければ起動
make db-up
```

### 2. 変更の実装

1. 新機能のブランチを作成
```bash
git checkout -b feature/新機能名
```

2. コードの変更実装

3. テストの実行
```bash
go test ./...
```

4. 統合テストの実行（必要に応じて）
```bash
make test-integration
```

### 3. API動作確認

#### REST APIの動作確認

```bash
make run
# curlなどでエンドポイントにリクエスト
```

#### GraphQLの動作確認

```bash
go run cmd/graphql/main.go
# ブラウザでGraphQL Playgroundにアクセス: http://localhost:8080
```

#### gRPCの動作確認

```bash
go run cmd/grpc/main.go
# grpcurlなどでgRPCサービスにリクエスト
```

## トラブルシューティング

### データベース接続エラー

エラーメッセージ: `SYSTEM_ERROR: データベース操作中にエラーが発生しました`

確認事項:
1. データベースコンテナが起動しているか確認
```bash
docker ps | grep postgres
```

2. PostgreSQLドライバが正しくインポートされているか確認
```go
import (
    _ "github.com/lib/pq" // PostgreSQLドライバ
)
```

3. 接続情報が正しいか確認
```bash
# デフォルト設定
- ホスト: localhost
- ポート: 5432
- ユーザー: testuser
- パスワード: testpass
- データベース: test_management
```

### Protocol Buffersコード生成エラー

エラーメッセージ: `protoc-gen-go: program not found or is not executable`

解決方法:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
```

### GraphQLコード生成エラー

エラーメッセージ: `cannot find package "github.com/99designs/gqlgen"`

解決方法:
```bash
go get github.com/99designs/gqlgen@v0.17.67
```
