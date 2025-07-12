# Terraform状態管理ガイド

## 1. 概要と問題点

Terraformでは、インフラストラクチャの現在の状態を「状態ファイル（state file）」として管理しています。AWS環境とTerraformステートの一貫性を確保することは非常に重要です。

### 1.1 従来の問題点

当プロジェクトでは、以下の問題が発生していました：

1. AWS CLIベースのクリーンアップコマンド（`cleanup-*`）がTerraformステートを更新しない
2. その結果、AWS環境とTerraformステートの間に不一致が発生
3. 後続のTerraform操作で予期しない動作やエラーの原因となる

### 1.2 新たなアプローチ

これらの問題を解決するため、以下の改善が実装されました：

1. Terraformを使用した一貫性のあるリソース管理
2. 開発ライフサイクルに基づいたコマンド体系の導入
3. ステート管理用の専用コマンドの提供

## 2. 正しいリソース管理の原則

### 2.1 基本原則

- **Terraformで作成したリソースはTerraformで削除する**
- AWSコンソールやAWS CLIでの直接操作は避ける
- 常にTerraformステートとAWS環境の一致を確保する

### 2.2 正しいクリーンアップ方法

```bash
# 特定のモジュールを削除する場合
terraform destroy -target=module.loadbalancer_api

# 環境全体を削除する場合
terraform destroy
```

### 2.3 ライフサイクルに基づいた推奨コマンド

```bash
# 開発環境の一時停止（ECSサービスとALBのみ削除）
make pause-dev TF_ENV=development

# 開発環境の完全停止（すべてのリソースを削除）
make stop-dev TF_ENV=development
```

## 3. Terraform状態管理コマンド

### 3.1 状態管理のためのユーティリティ

```bash
# Terraformステートのバックアップ
make terraform-backup TF_ENV=development

# Terraformステートのリセット（バックアップ後）
make terraform-reset TF_ENV=development

# Terraformステートと実際のAWS環境の一致を検証
make terraform-verify TF_ENV=development
```

### 3.2 Terraformベースのクリーンアップ

```bash
# 最小限クリーンアップ（ECSサービスとALB）
make terraform-cleanup-minimal TF_ENV=development

# 標準クリーンアップ（最小限 + RDS）
make terraform-cleanup-standard TF_ENV=development

# 完全クリーンアップ（すべてのリソース）
make terraform-cleanup-complete TF_ENV=development

# 安全なクリーンアップ（バックアップと検証を含む）
make terraform-safe-cleanup TF_ENV=development
```

## 4. 状態の不一致発生時の対処法

### 4.1 状態の診断

```bash
# 状態の検証
make terraform-verify TF_ENV=development

# 詳細な状態確認
cd deployments/terraform/environments/development
terraform plan -detailed-exitcode
```

### 4.2 不一致の修復手順

1. 現在の状態をバックアップ
```bash
make terraform-backup TF_ENV=development
```

2. 状態をリセット
```bash
make terraform-reset TF_ENV=development
```

3. AWS環境を反映するよう状態を更新
```bash
cd deployments/terraform/environments/development
terraform import module.networking.aws_vpc.main <vpc-id>
terraform import module.database.aws_db_instance.postgres <db-instance-id>
# 必要なリソースをすべてインポート
```

4. 再検証
```bash
make terraform-verify TF_ENV=development
```

## 5. モジュール間の依存関係管理

### 5.1 循環依存の問題と解決方法

プロジェクトでは、ECSサービスとロードバランサー間の循環依存関係が問題となりました。これは、ターゲットグループを独立したモジュールとして実装することで解決しました：

```
module.service_api → module.target_group_api ← module.loadbalancer_api
```

このパターンにより、循環依存を避けつつ、リソース間の連携を維持できます。

### 5.2 推奨されるモジュール構成

```
deployments/terraform/
└── modules/
    ├── networking/         # VPC、サブネット、ゲートウェイ
    ├── database/           # RDSインスタンス
    ├── service/
    │   ├── target-group/   # ALBターゲットグループ
    │   ├── load-balancer/  # ALBとリスナー
    │   └── ecs-service/    # ECSサービスとタスク定義
    └── shared/             # 共有リソース（ECSクラスター、IAMロールなど）
```

## 6. ベストプラクティス

### 6.1 デプロイにおけるベストプラクティス

- ターゲットを指定したデプロイよりも、環境全体のデプロイを優先する
- 依存関係のあるリソースは同時に更新する
- 常に最新のステートファイルを使用する（`terraform init`を適切に実行）

### 6.2 状態管理におけるベストプラクティス

- 操作前に常にステートをバックアップする
- 共有環境ではリモートステートを使用する
- 状態ファイルをバージョン管理システムに保存しない

### 6.3 チーム作業におけるベストプラクティス

- 状態ロックを尊重する（同時編集を避ける）
- 変更前に`terraform plan`で影響を確認する
- 大きな変更は小さなステップに分割する

## 7. トラブルシューティング

### 7.1 よくある問題と解決策

1. **問題**: `Error: resource already exists`
   **解決策**: 既存リソースをインポートするか、既存リソースを削除してから再作成

2. **問題**: `Error: Cycle: module.X (expand), module.Y (expand)`
   **解決策**: モジュール間の依存関係を見直し、循環を解消する

3. **問題**: リソースは存在するが、Terraformが認識しない
   **解決策**: `terraform import`コマンドでリソースをステートに追加

### 7.2 デバッグのヒント

- `-trace`フラグを使用して詳細ログを取得: `terraform apply -trace`
- `-var`フラグで変数を上書き: `terraform apply -var="db_instance_class=db.t3.small"`
- ステートファイルのバックアップと履歴を保持する

## 8. 参考資料

- [Terraform 公式ドキュメント](https://www.terraform.io/docs)
- AWS環境のリソース管理とコスト最適化ガイド（プロジェクト内ドキュメント）
- Makefile解説ドキュメント（プロジェクト内ドキュメント）