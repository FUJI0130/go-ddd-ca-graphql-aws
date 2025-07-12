# AWS環境構築ガイド：認証設定とTerraform基盤整備

## はじめに

このドキュメントは、テストケース管理システムのAWS環境構築に関する手順と知見をまとめたものです。AWS環境の初期セットアップからTerraformによるインフラストラクチャのコード化までの流れを記録し、将来の参照用のガイドとして活用できるよう構成しています。

## 1. AWS認証環境の準備

### 1.1 IAMユーザーの作成

AWSのベストプラクティスとして、ルートユーザーではなく適切な権限を持つIAMユーザーを作成して使用します。

1. AWSマネジメントコンソールにルートユーザーでログイン
2. サービス一覧から「IAM」を選択
3. 左側のナビゲーションから「ユーザー」→「ユーザーを作成」をクリック
4. 以下の情報を入力：
   - ユーザー名: `terraform-admin`
   - AWS Management Console へのユーザーアクセスを有効化
   - カスタムパスワードを設定
   - パスワードのリセットを不要に設定（任意）
5. 「次へ」をクリックし、権限設定画面で「既存のポリシーを直接アタッチ」を選択
6. `AdministratorAccess`ポリシーを選択（Terraform用の管理者権限）
7. 「次へ」をクリックし、タグ設定（任意）を行う
8. 「次へ」→「ユーザーの作成」でIAMユーザーを作成

### 1.2 IAMユーザーの確認とログイン

IAMユーザーを使ってAWSマネジメントコンソールにログインします。

1. IAMユーザーのサインインURL形式: 
   ```
   https://[ACCOUNT_ID].signin.aws.amazon.com/console
   ```
   または、
   ```
   https://console.aws.amazon.com/iam/
   ```
   にアクセスし、AWSアカウントID（12桁の数字）を入力

2. IAMユーザー名とパスワードを入力してログイン
3. IAMダッシュボードでユーザーの権限と設定を確認

### 1.3 アクセスキーの作成

TerraformやAWS CLIがAWSリソースにアクセスするためのプログラマティックアクセスキーを作成します。

1. IAMコンソールで「ユーザー」→「terraform-admin」を選択
2. 「セキュリティ認証情報」タブを選択
3. 「アクセスキーを作成」ボタンをクリック
4. 用途として「コマンドライン インターフェイス (CLI)」を選択
5. 必要に応じてタグを追加（任意）
6. 「アクセスキーを作成」をクリック
7. 表示されたアクセスキーIDとシークレットアクセスキーを安全に保存
   - **.csvファイルのダウンロードが推奨**
   - **重要**: シークレットアクセスキーはこの画面でしか表示されないため、必ず保存すること

### 1.4 AWS CLI設定

AWS CLIをローカル環境に設定し、作成したIAMユーザーの認証情報を使用できるようにします。

1. AWS CLIのインストール（システムレベル）が完了していることを確認
2. ターミナルで以下のコマンドを実行：
   ```bash
   aws configure
   ```
3. プロンプトに従って以下の情報を入力：
   - AWS Access Key ID: [作成したアクセスキーID]
   - AWS Secret Access Key: [作成したシークレットアクセスキー]
   - Default region name: ap-northeast-1 （東京リージョンの場合）
   - Default output format: json

4. 設定が正しいことを確認：
   ```bash
   aws sts get-caller-identity
   ```
   以下のような出力が表示されれば成功：
   ```json
   {
       "UserId": "AIDA...",
       "Account": "123456789012",
       "Arn": "arn:aws:iam::123456789012:user/terraform-admin"
   }
   ```

### 1.5 トラブルシューティング

**IAMユーザーログインの問題**
- アカウントIDまたはエイリアスがわからない場合は、ルートアカウントでログインし、右上のアカウント名をクリックするとアカウントIDが表示されます
- IAMユーザーのログインURLに正しいアカウントIDを入力しているか確認してください

**AWS CLI設定の問題**
- アクセスキーとシークレットアクセスキーの入力順序を間違えないよう注意してください
- アクセスキーIDは通常「AKIA」で始まります
- 設定ファイルを直接編集する場合：
  ```bash
  nano ~/.aws/credentials  # 認証情報
  nano ~/.aws/config      # リージョンなどの設定
  ```

## 2. Terraformリモートステート環境の構築

（次回のスレッドで続行予定）

## 3. Terraformプロジェクト構造

現在のプロジェクトには以下のTerraform構造が存在しています：

```
deployments/terraform/
├── environments/
│   ├── development/
│   │   ├── main.tf
│   │   ├── terraform.tfvars
│   │   └── variables.tf
│   └── production/
│       ├── main.tf
│       ├── terraform.tfvars
│       └── variables.tf
└── modules/
    ├── networking/
    │   ├── main.tf
    │   ├── outputs.tf
    │   └── variables.tf
    ├── database/
    │   ├── main.tf
    │   ├── outputs.tf
    │   └── variables.tf
    ├── ecs/
    │   ├── main.tf
    │   ├── outputs.tf
    │   └── variables.tf
    └── loadbalancer/
        ├── main.tf
        ├── outputs.tf
        └── variables.tf
```

この構造に基づいて、リモートステート環境の構築と各モジュールの実装を進めていきます。

## 4. 次のステップ

次回のスレッドでは以下の作業を予定しています：

1. S3バケットとDynamoDBテーブルの作成（Terraformリモートステート用）
2. Terraformバックエンド設定の更新
3. 基本ネットワークインフラのデプロイ

## 参考資料

- [AWS IAMベストプラクティス](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [AWS CLIコマンドリファレンス](https://awscli.amazonaws.com/v2/documentation/api/latest/index.html)
- [Terraformバックエンド設定ドキュメント](https://www.terraform.io/docs/language/settings/backends/s3.html)
- [AWS S3バケット命名規則](https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html)