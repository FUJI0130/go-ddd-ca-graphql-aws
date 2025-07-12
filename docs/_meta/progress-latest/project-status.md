# プロジェクト状況サマリー (2025-2-25)

## 1. プロジェクトフェーズと進捗
### 1.1 完了フェーズ
- フェーズ1：基盤構築（完了）
  - データベース設計
    - ER図設計とテーブル構造の確定
    - ENUMタイプの定義と移行
    - インデックスの設定
    - 更新日時トリガーの実装
  - Docker環境構築（RaspberryPi）
    - PostgreSQL 14.13の構築
    - データの永続化設定（Dockerボリューム）
    - システムPostgreSQLの無効化対応
  - マイグレーション実装
    - マイグレーションによるデータベース構造の作成
    - ENUMタイプの定義と移行
    - テーブル構造の作成
  - テストデータ投入
    - テストスイートのサンプルデータ作成
    - テストケースデータの投入
    - 工数記録の投入
    - 状態履歴の投入完了

- フェーズ2：アプリケーション実装（リポジトリ層完了）
  - リポジトリ層の実装
    - ベースリポジトリインターフェース定義
      - 基本的なCRUD操作を定義
      - ジェネリクスを使用した型安全な設計
    - TestSuiteリポジトリインターフェース定義
      - ベースリポジトリの継承
      - テストスイート固有の操作を追加
    - PostgreSQLリポジトリの実装
      - CRUD操作の実装
      - ステータス関連操作の実装
  - テスト環境の整備
    - VSCode Remote Development環境の構築
    - テストコードの実装
      - 値オブジェクト（SuiteStatus）のテスト実装
      - エンティティ（TestSuite）のテスト実装
      - PostgreSQLリポジトリの統合テスト実装
    - Docker環境の整備
      - テスト用DBコンテナの構築（ポート5433）
      - テスト用マイグレーション環境の整備
    - CI/CD基盤の整備
      - Makefileの基本コマンド整備
      - 統合テスト実行環境の整備

### 1.3 次のフェーズ予定
- フェーズ2の続き：アプリケーション層実装
  1. Protocol Buffers導入フェーズの完了（優先度：高）
     - ✅ protoファイルの作成
     - ✅ gRPCサービスの基本実装
     - ✅ ListTestSuites基本機能の実装完了
     - ✅ WatchTestSuite機能の実装
     - 実装優先順位：
       1. ✅ 基本的なRPC定義
       2. ✅ ListTestSuitesの実装
       3. ✅ ストリーミング機能
       4. [ ] エラーハンドリングの標準化
  
  2. gRPCサーバーの機能拡張（優先度：高）
      - ✅ リスト機能の完全実装
      - ✅ ストリーミング機能の実装
      - ✅ エラーハンドリングの強化
      - [ ] テストコードの拡充

  3. GraphQL導入フェーズ（優先度：中）
     - ✅ スキーマ定義
     - ✅ 基本リゾルバーの実装
     - ✅ GraphQL Playgroundの動作確認
     - ✅ TestSuiteのgroupsフィールドリゾルバーの実装
     - ✅ TestGroupのcasesフィールドリゾルバーの実装
     - ✅ DataLoader実装によるN+1問題対策
     - [ ] サブスクリプション機能の検討

### 1.4 学習目標
   - ✅ Protocol Buffersによる型安全なAPI設計（優先度：高）
     - ✅ protoファイルの設計手法
     - ✅ gRPCサービスのベストプラクティス
     - ✅ ストリーミングの実装方法
   - gRPCサービスの実装と運用（優先度：高）
     - ✅ サーバー実装
     - [ ] クライアント実装
     - ✅ エラーハンドリング
     - [ ] テスト手法
   - GraphQLによる柔軟なAPI設計（優先度：中）
     - ✅ スキーマ設計
     - ✅ 基本リゾルバー実装
     - ✅ N+1問題対策
   - マイクロサービスアーキテクチャの実践（優先度：中）
     - サービス間通信
     - エラー処理戦略
     - 監視と運用

### 1.5 実装アクションプラン
   1. ユースケース層のエラーハンドリング改善（優先度：高）
      - ✅ ドメインエラーへの変換処理の実装
      - ✅ エラーメッセージの改善
      - ✅ テストケースの追加
   
   2. リポジトリ層のエラーハンドリング整備（優先度：高）
      - ✅ DBエラーからドメインエラーへの変換処理の実装
      - ✅ PostgreSQLエラーコード（一意性制約違反、外部キー制約違反など）の定数化
      - ✅ テストスイートリポジトリの全メソッドのエラーハンドリング改善
      - ✅ トランザクション関連エラーの処理改善
      - ✅ テストケースの追加
   
   3. REST APIエラーハンドリング更新（優先度：中）
      - ✅ ドメインエラーからAPIエラーへの変換関数の実装
      - ✅ 全ハンドラーへの適用
      - ✅ エラーレスポンスの標準化

   4. GraphQL導入フェーズの基盤整備（優先度：中）
      - ✅ GraphQLスキーマの設計と定義
      - ✅ 基本的なプロジェクト構造の整備
      - ✅ gqlgenによるコード生成環境の構築
      - ✅ 基本クエリ・ミューテーションリゾルバーの実装
      - [ ] TestSuiteのgroupsフィールドリゾルバーの実装
      - [ ] TestGroupのcasesフィールドリゾルバーの実装
      - [ ] DataLoader実装によるN+1問題対策

### 直近の実装計画（2025-03-27更新）
1. ✅ IDジェネレーターのインターフェース化とモック対応
   - ✅ TestSuiteIDGenerator, TestGroupIDGenerator, TestCaseIDGeneratorインターフェース定義
   - ✅ TestSuiteインタラクターテストの修正
   - ✅ TestGroup/TestCaseインタラクターのテストコード作成

2. ✅ テスト環境の安定化
   - ✅ TestMain関数の改善（絶対パスの使用、エラー処理強化）
   - ✅ テストデータクリーンアップの改善
   - ✅ テスト環境IDジェネレーターの実装

3. ✅ GraphQL導入フェーズの続き
   - ✅ 基本的なGraphQLサーバーの動作確認
   - ✅ 開発環境セットアップドキュメントの整備
   - ✅ TestSuiteのgroupsフィールドリゾルバーの実装
   - ✅ TestGroupのcasesフィールドリゾルバーの実装
   - ✅ リゾルバー全体のNULLチェック対応
   - ✅ DataLoader実装によるN+1問題対策
   - ✅ GraphQLサーバー設定の最新化（非推奨API対応）

4. 新しい優先実装計画
   - [ ] GraphQLサブスクリプション機能の実装
   - [ ] 認証・認可の統合
   - [ ] フロントエンド連携のためのAPI改善
   - [ ] マイクロサービスアーキテクチャの検討

5. プロジェクト管理の改善
   - ✅ 開発環境セットアップガイドの作成
   - ✅ ブランチ戦略の見直しと標準化
   - ✅ コミット方針の確立
   - ✅ 作業フローの明確化
   - ✅ Claudeプロジェクトインストラクションの整理

6. 保留中の検討事項
   - より詳細な進捗率計算ロジックの実装（テストケースの状態を考慮）
   - エラーメッセージの国際化対応
   - API応答のバージョニング戦略
   - pkg/errorsディレクトリ構造のリファクタリング

## 2. 環境状況
### 2.1 開発環境
- RaspberryPi環境
  - Docker 27.3.1
  - Docker Compose v2.29.7
  - PostgreSQL 14.13（コンテナ）
  - Protocol Buffers compiler（protoc）v3.21.12-3
- 開発ツール
  - DBeaver 24.2.4（接続設定済み）
  - DB接続情報：192.168.10.5:5432
- VSCode Remote Development
  - SSH経由でのリモート開発環境
  - Go言語開発環境の設定
  - テスト実行環境の構築
  - Protocol Buffers/gRPC開発環境の設定

### 2.1.2 導入済み開発ツール
- Protocol Buffers関連
  - ✅ Protocol Buffers compiler (protoc) v3.21.12
  - ✅ Go言語用プラグイン（protoc-gen-go, protoc-gen-go-grpc）
  - ✅ Makefileコマンドの設定

### 2.1.3 追加予定の開発ツール
- Protocol Buffers関連（優先導入）
  - ✅ Protocol Buffers compiler (protoc) v3.17.3以上
  - ✅ gRPC-Go v1.56.1
  - ✅ protoc-gen-go v1.28.1
  - ✅ protoc-gen-go-grpc v1.2.0
  - [ ] buf（オプション：スキーマ管理用）
- GraphQL関連（次フェーズ）
  - [ ] gqlgen v0.17.38
  - [ ] graphql-playground（開発用UIツール）
  - [ ] dataloaden（N+1対策用）

### 2.2 データベース状況
- スキーマ構造
  - テーブル構造確定済み
    - test_suites
    - test_groups
    - test_cases
    - effort_records
    - status_history
  - ENUM定義完了
    - priority_enum: Critical, High, Medium, Low
    - suite_status_enum: 準備中, 実行中, 完了, 中断
    - test_status_enum: 作成, テスト, 修正, レビュー待ち, レビュー中, 完了, 再テスト
- データ状況
  - テストデータ投入完了
    - テストスイート：TS001-202411（商品管理システム）
    - テストケース：基本機能の検証データ
    - 工数・状態履歴：初期サンプルデータ


### 2.3 プロジェクト構造の更新状況
- /api
- /proto          # Protocol Buffers定義 ✅
  /testsuite
    /v1
      - test_suite.proto ✅
      - test_suite_service.proto ✅
      - test_suite.pb.go ✅
      - test_suite_service_grpc.pb.go ✅
- /graphql        # GraphQLスキーマ（予定）
- /internal
  /domain
    /repository
      - base_repository.go ✅
      - test_suite_repository.go ✅
      - test_group_repository.go ✅
      - test_case_repository.go ✅
      - id_generator.go ✅ (IDジェネレーターインターフェース追加)
  /usecase/interactor
  - ✅ test_suite_interactor_test.go (IDジェネレーター対応修正)
  - ✅ test_group_interactor_test.go (新規作成)
  - ✅ test_case_interactor_test.go (新規作成)
  /interface
    /grpc         # gRPCサーバー実装 ✅
      /handler
        - test_suite_server.go ✅
      /server
        - server.go ✅
    /graphql        # GraphQL実装 ✅
      /schema       # GraphQLスキーマ定義 ✅
      /resolver     # リゾルバー実装 ✅
        - resolver.go ✅
        - schema.resolvers.go ✅ (TestSuiteとTestGroupのリレーションリゾルバー実装済み)
      /model        # GraphQLモデル定義 ✅
      /dataloader   # DataLoader実装（予定）
    /api
      /handler
        - test_suite_handler.go ✅
        - test_suite_handler_test.go ✅
- /cmd
    /api           # REST APIエントリーポイント
    /grpc          # gRPCサーバーエントリーポイント ✅
      - main.go ✅
    /graphql          # GraphQLサーバーエントリーポイント ✅
    - main.go ✅
- /pkg
  /errors           # エラーハンドリング ✅
    - api_error.go  # 既存のREST APIエラー ✅
    - domain_error.go # ドメインエラーインターフェース ✅
    - validation_error.go # 検証エラー ✅
    - not_found_error.go # 未検出エラー ✅
    - conflict_error.go # 競合エラー ✅
    - permission_error.go # 権限エラー ✅
    - system_error.go # システムエラー ✅
    - grpc_error.go # gRPCエラー変換 ✅
    - domain_error_test.go # ドメインエラーテスト ✅
    - grpc_error_test.go # gRPCエラー変換テスト ✅
- test
  /integration       # 統合テスト
    /postgres
      - docker-compose.test.yml ✅


### 2.4 ブランチ戦略 （2025-03-27更新）

現在のブランチ状況：
```bash
fuji0130@raspberrypi:~/workspace/go-DDD-CA/go-ddd-ca $ git branch
  DBdesign
  domain-base
* feat/dataloader-implementation
  feature/graphql-resolver-refactoring
  feature/grpc-implementation
  main
```

- 実装フェーズごとのフィーチャーブランチ
  1. Protocol Buffers実装（feature/grpc-implementation）
     - ✅ feat/proto-definitions
     - ✅ feat/grpc-service
     - [ ] feat/grpc-client
  
  2. GraphQL実装
    - ✅ feature/graphql-resolver-refactoring (NULLチェック対応)
    - ✅ feat/dataloader-implementation (DataLoader実装、GraphQLサーバー設定)
    - [ ] feat/graphql-subscription

- ブランチ改善計画

- 新しいブランチ命名規則の採用:

- feat/: 新機能開発
- fix/: バグ修正
- refactor/: リファクタリング
- docs/: ドキュメント更新
- test/: テスト追加


- 移行手順:

- 不要な古いブランチを整理
- コミットメッセージに接頭辞を追加（feat:, fix:など）
- ブランチごとに明確な目的を設定



### 2.5 実装状況（2025-02-25現在）
### 2.5 実装状況（2025-03-27現在）
- internal/usecase/dto
 - ✅ TestSuiteCreateDTO
 - ✅ TestSuiteResponseDTO
 - ✅ TestSuiteUpdateDTO
 - ✅ TestSuiteStatusUpdateDTO
 - ✅ TestSuiteQueryParamDTO
 - ✅ バリデーションルール設計
- internal/usecase/interactor
 - ✅ CreateTestSuiteメソッド
 - ✅ GetTestSuiteメソッド
 - ✅ ListTestSuitesメソッド
 - ✅ UpdateTestSuiteメソッド
 - ✅ UpdateTestSuiteStatusメソッド
 - ✅ インタラクターのテストコード
- internal/interface/api/handler
 - TestSuiteHandler
   - ✅ 全メソッド実装完了
   - ✅ テストコード実装完了
   - ✅ エラーハンドリング実装完了
   - ✅ バリデーション実装完了
 - pkg/errors
  - ✅ ドメインエラーインターフェース実装
  - ✅ 各種エラー型の実装
  - ✅ APIエラー変換関数の実装
  - ✅ gRPCエラー変換関数の実装
  - ✅ テストコードの実装

- Protocol Buffers/gRPC実装
  - ✅ proto/testsuite/v1/test_suite.proto
  - ✅ proto/testsuite/v1/test_suite_service.proto
  - ✅ internal/interface/grpc/handler/test_suite_server.go
    - ✅ CreateTestSuite
    - ✅ GetTestSuite
    - ✅ UpdateTestSuite
    - ✅ UpdateTestSuiteStatus
    - ✅ ListTestSuites
    - ✅ WatchTestSuite (ストリーミング実装)
  - ✅ internal/interface/grpc/server/server.go
  - ✅ cmd/grpc/main.go

- GraphQL実装
  - ✅ internal/interface/graphql/schema/schema.graphqls
  - ✅ internal/interface/graphql/resolver/schema.resolvers.go
    - ✅ CreateTestSuite
    - ✅ UpdateTestSuite
    - ✅ UpdateTestSuiteStatus
    - ✅ TestSuite (単一取得)
    - ✅ TestSuites (一覧取得、ページネーション対応)
    - ✅ TestSuiteのgroupsフィールドリゾルバー
    - ✅ TestGroupのcasesフィールドリゾルバー
    - ✅ TestSuiteStatusChanged (スケルトンのみ)
  - ✅ internal/interface/graphql/dataloader
    - ✅ dataloader.go (ミドルウェア実装)
    - ✅ test_group_loader.go (グループローダー実装)
    - ✅ test_case_loader.go (ケースローダー実装)  
  - ✅ internal/interface/graphql/integration_test.go (改善済み)
  - ✅ cmd/graphql/main.go (最新のAPIで実装)- internal/usecase/dto
 - ✅ TestSuiteCreateDTO
 - ✅ TestSuiteResponseDTO
 - ✅ TestSuiteUpdateDTO
 - ✅ TestSuiteStatusUpdateDTO
 - ✅ TestSuiteQueryParamDTO
 - ✅ バリデーションルール設計
- internal/usecase/interactor
 - ✅ CreateTestSuiteメソッド
 - ✅ GetTestSuiteメソッド
 - ✅ ListTestSuitesメソッド
 - ✅ UpdateTestSuiteメソッド
 - ✅ UpdateTestSuiteStatusメソッド
 - ✅ インタラクターのテストコード
- internal/interface/api/handler
 - TestSuiteHandler
   - ✅ 全メソッド実装完了
   - ✅ テストコード実装完了
   - ✅ エラーハンドリング実装完了
   - ✅ バリデーション実装完了
 - pkg/errors
  - ✅ ドメインエラーインターフェース実装
  - ✅ 各種エラー型の実装
  - ✅ APIエラー変換関数の実装
  - ✅ gRPCエラー変換関数の実装
  - ✅ テストコードの実装

- Protocol Buffers/gRPC実装
  - ✅ proto/testsuite/v1/test_suite.proto
  - ✅ proto/testsuite/v1/test_suite_service.proto
  - ✅ internal/interface/grpc/handler/test_suite_server.go
    - ✅ CreateTestSuite
    - ✅ GetTestSuite
    - ✅ UpdateTestSuite
    - ✅ UpdateTestSuiteStatus
    - ✅ ListTestSuites
    - ✅ WatchTestSuite (ストリーミング実装)
  - ✅ internal/interface/grpc/server/server.go
  - ✅ cmd/grpc/main.go

- GraphQL実装
  - ✅ internal/interface/graphql/schema/schema.graphqls
  - ✅ internal/interface/graphql/resolver/schema.resolvers.go
    - ✅ CreateTestSuite
    - ✅ UpdateTestSuite
    - ✅ UpdateTestSuiteStatus
    - ✅ TestSuite (単一取得)
    - ✅ TestSuites (一覧取得、ページネーション対応)
    - ✅ TestSuiteのgroupsフィールドリゾルバー
    - ✅ TestGroupのcasesフィールドリゾルバー
    - ❌ TestSuiteStatusChanged (スケルトンのみ)
  - ✅ internal/interface/graphql/integration_test.go (改善済み)
  - ✅ cmd/graphql/main.go

- GraphQLテスト環境の改善
  - ✅ GraphQL統合テスト用TestMain関数の改善
    - ✅ 絶対パスを使用したマイグレーションファイルの読み込み
    - ✅ 詳細なエラーメッセージの追加
    - ✅ ファイルの存在確認機能の追加
  - ✅ GraphQLテストデータ管理の改善
    - ✅ 外部キー制約を一時的に無効化する方法の実装
    - ✅ IDリストのクリア機能
  - ✅ GraphQLテスト用IDジェネレーターの実装
    - ✅ TestEnvironmentSuiteIDGenerator
    - ✅ TestEnvironmentGroupIDGenerator
    - ✅ TestEnvironmentCaseIDGenerator

## 3. 保留タスク
### 3.1 インフラ関連
- Protocol Buffers関連の環境整備
  - [x] Protocol Buffersコンパイラ導入
  - [x] Go言語用プラグインのインストール
  - [ ] gRPCゲートウェイの検討
  - [ ] 負荷テスト環境の構築
  - [ ] 監視設定の更新
- AWS環境への移行計画
  - RDSへの移行計画（変更なし）
  - コンテナ環境の構築計画（変更なし）
- 監視・メンテナンス計画
  - gRPCサービスの監視追加
  - パフォーマンスメトリクスの追加
  - サービス単位のアラート設定

### 3.2 アプリケーション関連
- Protocol Buffers実装に向けた検討
  - ✅ メッセージ定義の基本実装
  - ✅ バージョニング戦略の基本実装
  - ✅ 設計方針の文書化と標準化
  - ✅ 後方互換性の維持方針の明確化

- API間の整合性維持
  - ✅ gRPCの基本実装完了
  - ✅ エラーハンドリングの統一
  - [ ] レスポンス形式の標準化

- ドキュメント管理の方針
  - ✅ Protocol Buffersの設計文書作成
  - ✅ エラーハンドリング方針の文書化
  - [ ] API仕様書の統合管理方針の策定


## 4. 参考情報

### 4.1 データベース構造
```mermaid
erDiagram
    test_suites ||--o{ test_groups : "contains"
    test_groups ||--o{ test_cases : "contains"
    test_cases ||--o{ effort_records : "tracks"
    test_cases ||--o{ status_history : "logs"

    test_suites {
        varchar(50) id PK
        varchar(100) name
        text description
        date estimated_start_date
        date estimated_end_date
        boolean require_effort_comment
        timestamp created_at
        timestamp updated_at
        suite_status_enum status
    }

    test_groups {
        varchar(50) id PK
        varchar(50) suite_id FK
        varchar(100) name
        text description
        integer display_order
        timestamp created_at
        timestamp updated_at
        suite_status_enum status
    }

    test_cases {
        varchar(50) id PK
        varchar(50) group_id FK
        varchar(200) title
        text description
        test_status_enum status
        priority_enum priority
        float planned_effort
        float actual_effort
        boolean is_delayed
        integer delay_days
        varchar(100) current_editor
        boolean is_locked
        timestamp created_at
        timestamp updated_at
    }

    effort_records {
        integer id PK
        varchar(50) test_case_id FK
        date record_date
        float effort_amount
        boolean is_additional
        text comment
        varchar(100) recorded_by
        timestamp created_at
    }

    status_history {
        integer id PK
        varchar(50) test_case_id FK
        test_status_enum old_status
        test_status_enum new_status
        timestamp changed_at
        varchar(100) changed_by
        text reason
    }
```

### 4.2 ENUM定義
```mermaid
classDiagram
    class priority_enum {
        <<enumeration>>
        Critical
        High
        Medium
        Low
    }

    class suite_status_enum {
        <<enumeration>>
        準備中
        実行中
        完了
        中断
    }

    class test_status_enum {
        <<enumeration>>
        作成
        テスト
        修正
        レビュー待ち
        レビュー中
        完了
        再テスト
    }
```

### 4.3 インデックス定義
```
PrimaryKeys                    SecondaryIndexes
------------------            ------------------------------------------
effort_records_pkey(id)       idx_effort_records_date(record_date)
status_history_pkey(id)       idx_test_cases_priority(priority)
test_cases_pkey(id)          idx_test_cases_status(status)
test_groups_pkey(id)         idx_test_groups_order(display_order, suite_id)
test_suites_pkey(id)
```

### 4.4 gRPC開発環境

```protobuf
// テストスイート定義例
syntax = "proto3";

package testsuite.v1;

import "google/protobuf/timestamp.proto";

message TestSuite {
  string id = 1;
  string name = 2;
  string description = 3;
  SuiteStatus status = 4;
  google.protobuf.Timestamp estimated_start_date = 5;
  google.protobuf.Timestamp estimated_end_date = 6;
  bool require_effort_comment = 7;
  float progress = 8;
  google.protobuf.Timestamp created_at = 9;
  google.protobuf.Timestamp updated_at = 10;
}

// スニペット - 実際の定義はprotoディレクトリを参照
```
# Protocol Buffers コマンド
protoc --proto_path=. \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/testsuite/v1/*.proto


## 4.2 開発環境構成

### 4.2.1 Docker環境
```yaml
services:
  postgres:
    image: postgres:14.13
    container_name: test_management_db
    environment:
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
      POSTGRES_DB: test_management
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app_network

networks:
  app_network:
    driver: bridge

volumes:
  postgres_data:
```

### 4.2.2 マイグレーション構成
```
scripts/migrations/
├── 000001_create_enums.up.sql
├── 000001_create_enums.down.sql
├── 000002_create_tables.up.sql
├── 000002_create_tables.down.sql
├── 000003_create_indexes.up.sql
├── 000003_create_indexes.down.sql
├── 000004_create_triggers.up.sql
└── 000004_create_triggers.down.sql
```

### 4.2.3 運用コマンド
```bash
# Docker操作
docker compose up -d
docker compose down
docker compose down -v  # ボリューム削除

# マイグレーション
migrate -path scripts/migrations -database "postgresql://testuser:testpass@localhost:5432/test_management?sslmode=disable" up
migrate -path scripts/migrations -database "postgresql://testuser:testpass@localhost:5432/test_management?sslmode=disable" down
```


### 5.1 ID体系
- テストスイート: `TS{001-999}-{YYYYMM}`
- テストグループ: `TS{000}TG{01-99}-{YYYYMM}`
- テストケース: `TS{000}TG{00}TC{001-999}-{YYYYMM}`

### 5.2 各種制約
- effort_records: `check_positive_effort (effort_amount > 0)`
- 更新日時自動更新トリガー（test_suites, test_groups, test_cases）
- 外部キー制約（suite_id, group_id, test_case_id）

### 5.3 使用中のツール
- データベース: PostgreSQL 14.13（Dockerコンテナ）
- コンテナ環境: Docker 27.3.1, Docker Compose v2.29.7
- 開発環境: RaspberryPi
- DB管理ツール: DBeaver 24.2.4（接続設定済み）
