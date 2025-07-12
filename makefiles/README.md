# Makefile 構造ガイド

## 概要

このプロジェクトのMakefileは、機能別に複数のファイルに分割されています。これにより、保守性の向上と機能の追加・変更が容易になります。

## ディレクトリ構造

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
    ├── instructions.mk      # インストラクション生成用コマンド
    ├── integration.mk       # インストラクション管理システム統合
    └── README.md            # このファイル
```

## 機能別Makefileの内容

### base.mk
- 共通変数の定義（TF_ENV, SERVICE_TYPE, AWS_REGION）
- カラー設定
- ヘルプコマンド
- 基本開発コマンド（run, build, test）
- Protocol Buffers関連

### test.mk
- 統合テスト関連コマンド
- GraphQL関連テスト
- リポジトリ層のテスト

### db.mk
- データベースの起動・停止
- マイグレーション関連コマンド
- テスト用DB操作

### docker.mk
- Dockerイメージのビルド・実行
- サービス別Dockerコマンド
- Dockerテスト
- Docker Compose関連コマンド

### aws.mk
- Terraformの基本コマンド
- SSMパラメータ管理
- terraform.tfvars更新
- ECRイメージ準備

### deploy.mk
- インフラストラクチャコンポーネントデプロイ
- サービスデプロイコマンド
- 補助コマンド
- アプリケーションデプロイワークフロー

### cost.mk
- AWS環境のコスト見積もり
- クリーンアップコマンド（最小限、標準、完全）
- 一時デプロイと検証

### instructions.mk
- インストラクション生成コマンド
- モジュール組み合わせ機能

### integration.mk
- instructions.mkのターゲットをメインMakefileから呼び出す統合コマンド

## 使用方法

通常のMakeコマンドと同様に使用できます。例：

```bash
make help                    # ヘルプを表示
make run                     # APIサーバーを実行
make deploy-api TF_ENV=development # APIをデプロイ
make instructions-aws        # AWS作業用インストラクションを生成
```

すべてのコマンドはメインのMakefileからアクセスできるため、個別のMakefileを直接呼び出す必要はありません。

## カスタマイズと拡張

新しい機能グループを追加する場合：

1. `makefiles/` ディレクトリに新しいMakefileを追加（例：`ci.mk`）
2. メインの `Makefile` に `include makefiles/ci.mk` を追加

## 変数の共有

各Makefileは基本的に独立していますが、`base.mk` で定義された変数（TF_ENV, SERVICE_TYPEなど）はすべてのMakefileで利用可能です。新しい共通変数を追加する場合は `base.mk` に定義してください。