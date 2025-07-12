# AWS Infrastructure ガイド

このプロジェクトのAWS環境構築に関する詳細文書は以下の3つに分かれています。

## 📋 ファイル構成

### terraform-architecture-overview.md
詳細: [./7_learning-materials/aws/terraform-architecture-overview.md](./7_learning-materials/aws/terraform-architecture-overview.md)
- **内容**: システム全体のアーキテクチャ設計・モジュール構造
- **こんな時に**: 全体構成を理解したい・設計思想を知りたい・依存関係を把握したい
- **規模**: 約10,000行・Mermaidダイアグラム豊富・設計原則詳細
- **特徴**: 3サービス共存設計・VPC/RDS/ECS全体俯瞰

### terraform-component-design.md
詳細: [./7_learning-materials/aws/terraform-component-design.md](./7_learning-materials/aws/terraform-component-design.md)
- **内容**: 各コンポーネントの詳細実装設計・基底モジュール設計
- **こんな時に**: 具体的な実装方法を知りたい・コード例を見たい・設計パターンを学びたい
- **規模**: 約8,000行・実装詳細・設計意思決定背景
- **特徴**: サービス固有実装・比較分析・コード例豊富

### terraform-workflow-guide.md
詳細: [./7_learning-materials/aws/terraform-workflow-guide.md](./7_learning-materials/aws/terraform-workflow-guide.md)
- **内容**: 実際のデプロイ・運用・検証・クリーンアップ手順
- **こんな時に**: 実際にデプロイしたい・Makefileコマンドを知りたい・トラブルシューティングしたい
- **規模**: 約6,000行・実践的運用手順・ベストプラクティス
- **特徴**: 実際の運用フロー・検証スクリプト・運用ノウハウ

## 🚀 Quick Start

### 基本環境構築（Thread 67-70で確立）
AWS環境構築を始める前に、以下の開発環境が必要です：

```bash
# 必要なツール（バージョン確認）
go version        # go1.23.11 推奨
gqlgen version    # v0.17.76 推奨
terraform version # v1.12.2 推奨
aws --version     # aws-cli/2.27.50 推奨
```

### 基本デプロイフロー
```bash
# 基本的なデプロイ手順
make terraform-init
make deploy-dev
make verify-deployment
```

### よくあるつまずきポイント
- **terraform.tfvars未作成**: `terraform.tfvars.example`をコピーして作成
- **AWS CLI未設定**: 認証設定前に課金防止策を確認
- **依存関係エラー**: 上記4つのツールのバージョン確認

## 📚 学習推奨順序

1. **初回理解**: terraform-architecture-overview.md（全体像把握）
2. **実装学習**: terraform-component-design.md（詳細理解）
3. **実践運用**: terraform-workflow-guide.md（実際のデプロイ）

## ⚠️ 重要な注意事項

- **課金防止**: AWS認証設定前に必ず課金防止策を確認
- **リソース管理**: デプロイ後は適切なクリーンアップを実施
- **環境分離**: development環境での検証を推奨

---

詳細な技術情報・実装手順・運用ノウハウは上記3つの専門文書をご参照ください。