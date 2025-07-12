# Makefile解説ドキュメント（更新版）

## 概要

このドキュメントでは、テストケース管理システムのMakefileについて解説します。Makefileは様々な開発タスク、テスト、デプロイタスク、AWSリソース管理を自動化するためのコマンド集です。整理された構造で、関連する操作をグループ化し、開発からAWS環境へのデプロイまでをカバーしています。

## 構造と特徴

### Makefileの分割構造

プロジェクトのMakefileは機能別に分割され、保守性と拡張性を向上させています：

```
./
├── Makefile                 # メインMakefile（他のMakefileをインクルード）
└── makefiles/
    ├── base.mk              # 基本変数と開発コマンド
    ├── test.mk              # テスト関連コマンド
    ├── db.mk                # データベース関連コマンド
    ├── docker.mk            # Docker関連コマンド
    ├── aws.mk               # AWS/Terraform関連コマンド
    ├── deploy.mk            # デプロイ関連コマンド
    ├── cost.mk              # コスト最適化関連コマンド
    ├── terraform-cleanup.mk # Terraformベースのクリーンアップ
    ├── workflow.mk          # 開発ライフサイクル管理 (新規)
    ├── instructions.mk      # インストラクション生成用コマンド
    ├── integration.mk       # インストラクション管理システム統合
    └── README.md            # Makefile構造ガイド
```

### 主要セクション

Makefileの機能は以下のセクションに整理されています：

1. **基本開発コマンド** - アプリケーションの実行、ビルド、テスト
2. **データベース操作** - DBコンテナの起動/停止、マイグレーション
3. **テスト関連コマンド** - 各種統合テストの実行
4. **Docker関連コマンド** - イメージのビルド、コンテナ実行、テスト
5. **AWS/Terraform関連コマンド** - インフラのデプロイと管理
6. **サービスデプロイコマンド** - 各サービスのデプロイと検証
7. **開発環境ライフサイクル管理** - AWS環境の開始・一時停止・停止・再開
8. **コスト最適化コマンド** - AWS環境のコスト管理とクリーンアップ
9. **インストラクション管理** - Claudeへの指示生成と管理

### 特徴

- **環境変数の活用** - `TF_ENV`, `SERVICE_TYPE`などの変数を使用
- **カラー出力** - ターミナルでの視認性向上のためのANSIカラーコード
- **ヘルプ機能** - 詳細な使用方法を表示する`help`ターゲット
- **依存関係の自動化** - 依存するタスクの自動実行
- **エラーハンドリング** - コマンド失敗時の適切なフィードバック
- **モジュラー設計** - 機能別のMakefileによる保守性向上

## 主要コマンドグループ

### 基本開発コマンド

```bash
# アプリケーション起動
make run                # APIサーバー起動
make run-graphql        # GraphQLサーバー起動
make run-grpc           # gRPCサーバー起動

# ビルド
make build              # アプリケーションバイナリをビルド
make proto              # Protocol Buffersコードを生成
```

### データベース操作

```bash
# データベース管理
make db-up              # DBコンテナを起動
make db-down            # DBコンテナを停止
make migrate            # マイグレーションを実行
make migrate-down       # マイグレーションをロールバック

# テスト用データベース
make test-db-up         # テスト用DBを起動 (ポート5433)
make test-db-down       # テスト用DBを停止
```

### テスト関連コマンド

```bash
# 基本テスト
make test               # 全単体テストを実行

# 統合テスト
make test-integration   # リポジトリ層の統合テスト
make test-graphql       # GraphQL統合テスト
make test-all-integration  # すべての統合テスト
make test-graphql-resolver # GraphQLリゾルバーのテスト
```

### Docker関連コマンド

```bash
# Dockerイメージビルド
make docker-build-api      # APIサービスのイメージビルド
make docker-build-graphql  # GraphQLサービスのイメージビルド
make docker-build-grpc     # gRPCサービスのイメージビルド

# Dockerコンテナ実行
make docker-run-api        # APIコンテナを起動
make docker-run-graphql    # GraphQLコンテナを起動
make docker-run-grpc       # gRPCコンテナを起動

# Dockerテスト
make test-docker-api       # APIサービスのテスト
make test-docker-graphql   # GraphQLサービスのテスト
make test-docker-grpc      # gRPCサービスのテスト
make test-docker-all       # すべてのサービスのテスト
```

### AWS/Terraform関連コマンド

```bash
# Terraformの基本コマンド
make tf-status          # AWS環境の状態を確認
make tf-init            # Terraformを初期化
make tf-plan MODULE=xxx # 指定モジュールのプラン作成
make tf-apply           # プランを適用
make tf-destroy MODULE=xxx # リソースを破棄

# ECRイメージ準備
make prepare-ecr-image SERVICE_TYPE=api      # APIイメージをECRに準備
make prepare-all-ecr-images                  # 全サービスのイメージを準備
```

### サービスデプロイコマンド

```bash
# コンポーネントデプロイ
make deploy-network     # ネットワークリソースをデプロイ
make deploy-database    # データベースをデプロイ
make deploy-ecs-cluster # ECSクラスターをデプロイ

# サービスデプロイ
make deploy-api         # APIサービスをデプロイ
make deploy-graphql     # GraphQLサービスをデプロイ
make deploy-grpc        # gRPCサービスをデプロイ
make deploy-all-services # すべてのサービスをデプロイ
```

### 開発環境ライフサイクル管理（新規）

```bash
# 開発ライフサイクル
make start-dev          # 開発環境の初期化（コア基盤のみ）
                         # コスト: ~$0.80/日、時間: ~15分
                         # 用途: 開発環境の初期セットアップ

make pause-dev          # 開発の一時停止（ECSサービスのみ削除）
                         # 削減額: ~$1.00/日、再開時間: ~5分
                         # 用途: 数時間〜数日の開発休止

make resume-dev         # 一時停止状態からの再開
                         # 追加コスト: ~$1.00/日、時間: ~5分
                         # 用途: pause-dev後の開発再開

make stop-dev           # 開発の完全停止（すべてのリソースを削除）
                         # 削減額: ~$2.00/日、再構築時間: ~30分
                         # 用途: 数日以上の長期開発休止

make quick-test         # 一時的なテスト実行（デプロイ→検証→クリーンアップ）
                         # 総コスト: ~$0.10（約1時間の使用を想定）
                         # 用途: 短時間の検証作業
```

この新しいコマンド体系は、開発環境のライフサイクル全体をカバーし、AWS環境とTerraform状態の一貫性を保ちます。古いクリーンアップコマンドからの移行については「コマンド移行ガイド」セクションを参照してください。

### Terraformステート管理コマンド

```bash
# Terraformステート管理
make terraform-backup   # Terraformステートをバックアップ
make terraform-reset    # ステートをリセット（バックアップ後）
make terraform-verify   # ステートとAWS環境の一致を検証

# Terraformベースのクリーンアップ
make terraform-cleanup-minimal   # ECSサービスとALBのみ削除
make terraform-cleanup-standard  # 最小限 + RDSを削除
make terraform-cleanup-complete  # すべてのリソースを削除
make terraform-safe-cleanup      # バックアップと検証を含む安全なクリーンアップ
```

### コスト最適化コマンド

```bash
# リソース状態とコスト確認
make cost-estimate      # AWS環境のコスト見積もりを取得
make check-resources    # 現在のAWSリソース状態を確認
```

### インストラクション管理コマンド

```bash
# 基本インストラクション生成
make instructions-aws         # AWS作業用インストラクション
make instructions-backend     # バックエンド開発用インストラクション
make instructions-frontend    # フロントエンド開発用インストラクション
make instructions-problem     # 問題解決用インストラクション

# カスタム組み合わせとユーティリティ
make instructions-custom MODULES="aws backend"  # カスタムインストラクション
```

## 主要なワークフローの解説

### 1. 新しい開発ライフサイクルワークフロー

```bash
# 1. 開発開始: コア環境のみデプロイ（VPC、RDS、ECSクラスター）
make start-dev TF_ENV=development

# 2. サービスデプロイ（必要に応じて）
make deploy-api TF_ENV=development

# 3. 開発・テスト作業...

# 4. 一時的な開発休止（数時間〜数日）: ECSサービスとALBのみ削除
make pause-dev TF_ENV=development

# 5. 開発再開: ECSサービスとALBを再デプロイ
make resume-dev TF_ENV=development

# 6. 長期的な開発休止（数日以上）: すべてのリソースを削除
make stop-dev TF_ENV=development

# 7. 短時間の検証のみ: 一時的なデプロイ→検証→クリーンアップ
make quick-test TF_ENV=development
```

### 2. ローカル開発ワークフロー

```bash
# データベース起動
make db-up

# マイグレーション実行
make migrate

# アプリケーション起動
make run  # または make run-graphql, make run-grpc
```

### 3. Dockerテストワークフロー

```bash
# すべてのサービスをテスト（DBの自動起動・停止を含む）
make test-docker-all

# または個別サービスのテスト
make test-docker-api
make test-docker-graphql
make test-docker-grpc
```

### 4. AWS完全デプロイワークフロー

```bash
# 事前にAWS認証情報を設定しておく
aws configure

# 完全デプロイ（ECRイメージ準備→tfvars更新→デプロイ→検証）
make deploy-app-workflow

# または、SSMパラメータ確認付きの完全デプロイ
make deploy-app-workflow-secure
```

## コマンド移行ガイド（非推奨→推奨）

以下の非推奨コマンドは、より堅牢で一貫性のある新コマンドに置き換えられています：

| 非推奨コマンド | 推奨コマンド | 主な改善点 |
|--------------|------------|---------|
| `cleanup-minimal` | `pause-dev` | Terraformステートの更新、直感的な名前 |
| `cleanup-standard` | `terraform-cleanup-standard` | Terraformステートの更新 |
| `cleanup-complete` | `stop-dev` | Terraformステートの更新、直感的な名前 |
| `verify-and-cleanup-api` | `quick-test` | 一貫したライフサイクル管理、状態更新 |
| `temporary-deploy-api` | `quick-test` | 自動クリーンアップ、より堅牢な実装 |

非推奨コマンドは2025年7月に完全に削除される予定です。それまでの間、警告メッセージが表示されます。

## 環境変数の指定方法

Makefileのコマンドは、環境変数を指定して実行できます：

```bash
# 開発環境（デフォルト）
make deploy-network

# 本番環境
TF_ENV=production make deploy-network

# GraphQLサービス
SERVICE_TYPE=graphql make prepare-ecr-image
```

## AWSコスト管理とベストプラクティス

### コスト管理の基本原則

1. **必要最小限のリソース利用**
   - 開発に必要なリソースのみをデプロイ
   - 不要になったリソースは速やかに削除

2. **ライフサイクルに応じたリソース管理**
   - 短期休止: `pause-dev`（ECSサービスのみ削除）
   - 長期休止: `stop-dev`（すべてのリソースを削除）

3. **定期的なコスト確認**
   - `make cost-estimate` で現在のコストを確認
   - 想定外のコストが発生していないか監視

### リソースのコスト影響度

| リソース | 概算コスト | 影響度 | 削除タイミング |
|--------|---------|-------|------------|
| ECSサービス/ALB | ~$1.00/日 | 中 | 数時間の休止時 |
| RDSインスタンス | ~$0.80/日 | 高 | 数日の休止時 |
| VPC/NAT Gateway | ~$0.20/日 | 低 | 長期休止時 |

詳細は「AWS環境のリソース管理とコスト最適化ガイド」を参照してください。

## 注意点と推奨事項

1. **環境変数の管理**
   - データベース認証情報は`.env.terraform`ファイルで管理することを推奨

2. **AWS環境とTerraform状態の一貫性**
   - AWS環境の変更は常にTerraformを通して行う
   - 状態の不一致が疑われる場合は `make terraform-verify` を実行

3. **コスト管理**
   - 長時間のテスト/検証は避け、検証後に速やかにリソースを削除
   - 定期的にコスト見積もりを確認（`make cost-estimate`）
   - 「デプロイ → 作業 → クリーンアップ → 再開」のサイクルを意識する

4. **インストラクション生成**
   - 新しいスレッドを開始する際は、適切なインストラクションを生成
   - 特に複雑な作業や重要な決定を伴う場合は確認ステップを徹底

5. **ターゲット名の選択**
   - 新しいターゲットを追加する場合は、既存の命名パターンに従う
   - 例: `deploy-xxx`, `verify-xxx`, `test-xxx`など
