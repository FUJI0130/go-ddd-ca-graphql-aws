# AWS環境構築ガイド：Terraformリモートステート環境

## 概要

このガイドでは、Terraformを使用したインフラストラクチャの管理に必要なリモートステート環境をAWS上に構築する手順を説明します。S3バケットとDynamoDBテーブルを使用して、Terraformの状態ファイルを安全に保存し、複数人での作業時のロック機能を実現します。

## 1. 前提条件

- AWSアカウントが作成済みであること
- AWS CLIがインストールされ、適切に設定されていること
- 必要な権限を持つIAMユーザーが作成されていること

## 2. リモートステート用S3バケットの作成

### 2.1 S3バケットの作成

```bash
aws s3api create-bucket \
  --bucket terraform-state-testmgmt \
  --region ap-northeast-1 \
  --create-bucket-configuration LocationConstraint=ap-northeast-1
```

### 2.2 バージョニングの有効化

状態ファイルの変更履歴を保持するためにバージョニングを有効化します：

```bash
aws s3api put-bucket-versioning \
  --bucket terraform-state-testmgmt \
  --versioning-configuration Status=Enabled
```

### 2.3 暗号化の設定

セキュリティを向上させるために、バケットのデフォルト暗号化を設定します：

```bash
aws s3api put-bucket-encryption \
  --bucket terraform-state-testmgmt \
  --server-side-encryption-configuration '{"Rules": [{"ApplyServerSideEncryptionByDefault": {"SSEAlgorithm": "AES256"}}]}'
```

## 3. ステートロック用DynamoDBテーブルの作成

複数人が同時にTerraformを実行する際の競合を防ぐためのロック機能を提供します：

```bash
aws dynamodb create-table \
  --table-name terraform-state-lock \
  --billing-mode PAY_PER_REQUEST \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --region ap-northeast-1
```

## 4. Terraformバックエンド設定

プロジェクトのTerraform設定に、作成したS3バケットとDynamoDBテーブルを使用するためのバックエンド設定を追加します：

```hcl
terraform {
  backend "s3" {
    bucket         = "terraform-state-testmgmt"
    key            = "development/terraform.tfstate"
    region         = "ap-northeast-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}
```

## 5. Terraformの初期化

バックエンド設定を行った後、以下のコマンドでTerraformを初期化します：

```bash
terraform init
```

## 6. トラブルシューティング

### 6.1 バケット名の競合

S3バケット名はグローバルに一意である必要があります。名前の競合が発生した場合は、異なるバケット名を試してください。

### 6.2 権限の問題

エラーが発生した場合、IAMユーザーに必要な権限（S3、DynamoDB）が付与されているか確認してください。

### 6.3 Terraformモジュールパスの問題

モジュールのソースパスには変数を使用せず、相対パスで直接指定してください：
```hcl
module "networking" {
  source = "../../modules/networking"
  # 変数や関数を使用しない: "${path.module}/../../modules/networking" は不可
}
```

## 7. ベストプラクティス

- バケット名とテーブル名はプロジェクト名を含む命名規則を使用する
- 本番環境と開発環境で異なるキーを使用する（例：`production/terraform.tfstate`と`development/terraform.tfstate`）
- リモートステート関連のリソースは手動で作成し、通常のTerraform管理対象に含めない
- バケットのライフサイクルポリシーを設定し、古いバージョンの状態ファイルを適切に管理する