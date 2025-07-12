# AWS環境構築ガイド：ネットワークインフラのデプロイ

## 概要

このガイドは、テストケース管理システムのAWS環境構築における、ネットワークインフラストラクチャのデプロイに関する手順と知見をまとめています。Terraformを使用してVPC、サブネット、ゲートウェイなどの基本ネットワークリソースをデプロイする方法を説明します。

## 前提条件

- AWS IAMユーザーが作成済みであること
- AWS CLIがインストールされ、適切に設定されていること
- Terraformがインストールされていること
- リモートステート環境（S3バケットとDynamoDBテーブル）が構築済みであること

## 1. 環境変数の設定

### 1.1 認証情報の管理

データベース認証情報を安全に管理するための`.env.terraform`ファイルを作成します：

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

## 2. ネットワークインフラのデプロイ

### 2.1 プランの作成と確認

まずはデプロイプランを作成して内容を確認します：

```bash
# 環境変数を読み込み
source ~/.env.terraform

# プランの作成
make tf-plan MODULE=network
```

このコマンドにより、以下のリソースを含むプランが生成されます：
- VPC (CIDR: 10.0.0.0/16)
- パブリックサブネット (2つ、異なるAZ)
- プライベートサブネット (2つ、異なるAZ)
- インターネットゲートウェイ
- NATゲートウェイ
- ルートテーブル (パブリック用とプライベート用)
- セキュリティグループ

### 2.2 デプロイの実行

プランの内容を確認した後、デプロイを実行します：

```bash
make deploy-network
```

このコマンドは内部的に以下のTerraformコマンドを実行します：
```bash
terraform -chdir=${TF_DIR} apply tfplan
```

### 2.3 デプロイ時の注意点

- デプロイには数分かかる場合があります（特にNATゲートウェイの作成）
- コンソール出力でエラーや警告がないか確認してください
- 途中でエラーが発生した場合は、`terraform destroy`を実行して一度リソースを削除し、問題を修正してから再度デプロイすることをお勧めします

## 3. デプロイ結果の検証

### 3.1 AWS Management Consoleでの確認

AWS Management Consoleにログインし、以下の項目を確認します：
- VPCサービスを開き、作成されたVPCの存在と構成
- サブネットの設定（パブリック/プライベート、CIDRブロック、AZ）
- ゲートウェイの接続状態
- ルートテーブルの構成

### 3.2 AWS CLIでの確認

AWS CLIを使用して作成されたリソースを確認します：

```bash
# VPCの確認
aws ec2 describe-vpcs --filters "Name=tag:Name,Values=development-vpc" --query "Vpcs[*].{VpcId:VpcId,CidrBlock:CidrBlock}"

# サブネットの確認
aws ec2 describe-subnets --filters "Name=vpc-id,Values=$(aws ec2 describe-vpcs --filters "Name=tag:Name,Values=development-vpc" --query "Vpcs[0].VpcId" --output text)" --query "Subnets[*].{SubnetId:SubnetId,CidrBlock:CidrBlock,AZ:AvailabilityZone,Name:Tags[?Key=='Name'].Value|[0]}"

# NATゲートウェイの確認
aws ec2 describe-nat-gateways --filter "Name=vpc-id,Values=$(aws ec2 describe-vpcs --filters "Name=tag:Name,Values=development-vpc" --query "Vpcs[0].VpcId" --output text)" --query "NatGateways[*]"
```

## 4. トラブルシューティング

### 4.1 一般的な問題と解決策

#### NAT Gateway作成の失敗
問題: Elastic IPの制限またはサブネット設定の問題でNAT Gatewayの作成が失敗
```
解決策:
1. AWS Management Consoleで作成されたリソースを確認
2. Elastic IPの制限に達していないか確認
3. 必要に応じて不要なElastic IPを解放
4. 再度デプロイを試みる
```

#### リソース制限エラー
問題: AWSアカウントのリソース制限に達してデプロイに失敗
```
解決策:
1. エラーメッセージから制限に達したリソースを確認
2. AWS Support Centerから制限引き上げリクエストを作成
3. 制限が引き上げられた後に再度デプロイ
```

### 4.2 Terraformステートの確認

デプロイ状態を確認するには：

```bash
# テラフォームステータスの確認
make tf-status
```

## 5. コスト管理

VPCとサブネット自体はコストが発生しませんが、以下のコンポーネントには課金があります：

1. **NATゲートウェイ**：時間単位の料金 + データ処理量
2. **Elastic IP**：EC2インスタンスにアタッチされていない場合に課金

開発環境での不要なコストを抑えるためのヒント：
- 使用しない時間帯はNATゲートウェイとElastic IPを削除（`make tf-destroy MODULE=network`）
- 本番環境では高可用性のために複数AZにNATゲートウェイを配置するが、開発環境では単一のNATゲートウェイで十分

## 6. 次のステップ

ネットワークインフラのデプロイが完了したら、次のステップに進みます：

1. データベースモジュールの非推奨属性を修正
   ```bash
   # outputs.tfファイルでnameをdb_nameに変更
   ```

2. データベースモジュールをデプロイ
   ```bash
   make tf-plan MODULE=database
   make deploy-database
   ```

3. ECSクラスターとロードバランサーの準備
   - ECRリポジトリの作成
   - Dockerイメージのビルドとプッシュ
   - ECSタスク定義の確認と調整