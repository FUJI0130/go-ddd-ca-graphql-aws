# AWS環境構築ガイド：RDSデータベースのデプロイ

## 概要

このドキュメントでは、テストケース管理システムのAWS環境におけるRDSデータベースのデプロイ手順と知見を記録しています。VPC環境が構築済みであることを前提に、Terraformを使用してPostgreSQLデータベースをデプロイする方法を解説します。

## 前提条件

- AWSアカウントが作成済みであること
- AWS CLIがインストールされ、適切に設定されていること
- Terraformがインストールされていること
- VPC、サブネット、ルートテーブルなどのネットワークリソースがデプロイ済みであること
- データベース認証情報（ユーザー名・パスワード）が環境変数として設定されていること

## 1. データベースモジュールの構成

### 1.1 ディレクトリ構造

```
deployments/terraform/
├── environments/
│   └── development/
│       ├── main.tf           # 環境固有の設定
│       ├── terraform.tfvars  # 変数値の設定
│       └── variables.tf      # 変数定義
└── modules/
    └── database/
        ├── main.tf           # RDSリソース定義
        ├── outputs.tf        # 出力変数
        └── variables.tf      # モジュール変数
```

### 1.2 主要設定項目

RDSインスタンスの主な設定項目：

- **エンジン**: PostgreSQL 14.13
- **インスタンスクラス**: db.t3.small
- **ストレージ**: 20GB (gp3)
- **自動スケーリング上限**: 100GB
- **バックアップ保持期間**: 7日間
- **メンテナンスウィンドウ**: 月曜日 04:00-05:00
- **バックアップウィンドウ**: 03:00-04:00
- **ストレージ暗号化**: 有効
- **パブリックアクセス**: 無効
- **マルチAZ**: 無効（開発環境）
- **最終スナップショット**: 不要（開発環境）

## 2. デプロイ前の準備

### 2.1 環境変数の設定

データベース認証情報を安全に管理するための環境変数設定：

```bash
# ~/.env.terraform ファイルを作成
cat > ~/.env.terraform << EOF
export TF_VAR_db_username=testadmin
export TF_VAR_db_password=SecurePassword123!
EOF

# ファイルのパーミッションを制限
chmod 600 ~/.env.terraform

# 環境変数の読み込み
source ~/.env.terraform
```

### 2.2 デプロイ前の環境確認

デプロイ前にVPC環境が正しく構築されているか確認：

```bash
make tf-status
```

出力例：
```
development環境のAWSインフラストラクチャの状態を確認しています...
✓ 1個のVPCがデプロイされています
  - VPC: vpc-0d2b65172956387de (development-vpc)
    サブネット: 4個
✗ RDSインスタンス (development-postgres) は存在しません
```

## 3. 統合デプロイタスクの実装

複数のコマンドを一括で実行するための統合デプロイタスクを実装：

```makefile
# データベースの完全デプロイ（環境変数読み込み→状態確認→プラン→デプロイ→検証）
deploy-db-complete:
	@echo "データベースの完全デプロイを開始します..."
	@if [ -f ~/.env.terraform ]; then \
		echo "環境変数を読み込んでいます..."; \
		. ~/.env.terraform; \
	else \
		echo "警告: ~/.env.terraform が見つかりません。データベース認証情報が設定されているか確認してください。"; \
	fi
	@echo "現在のインフラ状況を確認しています..."
	@make tf-status
	@echo "データベースデプロイのプランを作成しています..."
	@make tf-plan MODULE=database
	@echo "データベースをデプロイしています..."
	@make deploy-database
	@echo "デプロイ結果を確認しています..."
	@make tf-status
	@echo "データベースデプロイプロセスが完了しました"
```

## 4. デプロイの実行

統合デプロイタスクを使用したデプロイの実行：

```bash
make deploy-db-complete
```

このコマンドは以下のステップを自動的に実行します：
1. 環境変数の読み込み
2. 現在のインフラ状態確認
3. データベースデプロイのプラン作成
4. データベースのデプロイ
5. デプロイ結果の確認

## 5. デプロイ完了の確認

デプロイが完了すると、以下のようなメッセージが表示されます：

```
module.database.aws_db_instance.main: Creation complete after 7m59s [id=development-postgres]
```

また、`tf-status`コマンドで確認すると、RDSインスタンスの存在が確認できます：

```
✓ RDSインスタンス (development-postgres) は存在します
  - ステータス: available
  - エンジン: postgres 14.13
```

## 6. RDS接続情報の取得

デプロイしたRDSインスタンスの接続情報を取得するには：

```bash
aws rds describe-db-instances \
  --db-instance-identifier development-postgres \
  --query 'DBInstances[0].{Endpoint:Endpoint.Address,Port:Endpoint.Port,DBName:DBName}'
```

出力例：
```json
{
    "Endpoint": "development-postgres.xxxxxxxxxx.ap-northeast-1.rds.amazonaws.com",
    "Port": 5432,
    "DBName": "test_management_dev"
}
```

## 7. トラブルシューティング

### 7.1 デプロイ時間について

RDSインスタンスの作成には通常5〜15分程度かかります。これはAWSが以下の作業を行うためです：

- 物理ストレージのプロビジョニング
- 仮想マシンインスタンスの作成と初期化
- データベースエンジンのインストールと設定
- セキュリティ設定の適用
- バックアップ機能の設定
- データベースパラメータの適用
- ネットワーク接続の確立

デプロイ中は `Still creating...` メッセージが表示され続けますが、これは正常な動作です。

### 7.2 接続エラーの対処

RDSインスタンスに接続できない場合のチェックポイント：

1. セキュリティグループのルール確認
   ```bash
   aws ec2 describe-security-groups --group-id sg-0cea11654f45fb277
   ```

2. サブネットグループの確認
   ```bash
   aws rds describe-db-subnet-groups --db-subnet-group-name development-db-subnet-group
   ```

3. ルートテーブルとNATゲートウェイの確認
   ```bash
   aws ec2 describe-route-tables --filters "Name=association.subnet-id,Values=subnet-02725de2915aba322"
   ```

## 8. コスト管理

RDSインスタンスは以下の要素で課金されます：

1. **インスタンス時間**: db.t3.small で約$0.034/時間
2. **ストレージ**: gp3 20GB で約$0.10/GB/月
3. **バックアップストレージ**: バックアップ保持期間が7日で課金あり

開発環境での不要なコストを抑えるためのヒント：
- 使用しない時間帯はRDSインスタンスを停止
- 長期間使用しない場合は削除（最終スナップショットを取得）

```bash
# RDSインスタンスの停止
aws rds stop-db-instance --db-instance-identifier development-postgres

# RDSインスタンスの削除（最終スナップショットなし）
aws rds delete-db-instance --db-instance-identifier development-postgres --skip-final-snapshot
```

## 9. 次のステップ

RDSインスタンスのデプロイが完了したら、次のステップに進みます：

1. アプリケーション設定ファイルの更新
   - RDSエンドポイントとポート設定
   - データベース名とユーザー情報の設定

2. データベースマイグレーションの実行
   - テーブルスキーマの作成
   - 初期データの投入

3. アプリケーションからの接続テスト
   - 接続文字列の検証
   - データアクセス機能の動作確認