# 開発環境セットアップガイド（品質改善版）

## はじめに：このガイドの目的と使い方

このガイドは、**Go-DDD-CA プロジェクト**の開発環境を構築するための実用的な手順を提供します。ただし、単なる「コマンドの羅列」ではなく、**なぜそのコマンドが必要なのか**、**どのような効果があるのか**を理解しながら進められるよう設計されています。

### なぜこのガイドが必要なのか？

現代のWebアプリケーション開発では、多くの技術要素が絡み合います。本プロジェクトでは以下の技術を統合しています：

```mermaid
graph TB
    A[開発者] --> B[開発環境]
    B --> C[Go言語アプリケーション]
    B --> D[React+TypeScript フロントエンド]
    B --> E[PostgreSQL データベース]
    B --> F[AWS環境]
    
    C --> G[REST API]
    C --> H[GraphQL API]
    C --> I[gRPC API]
    
    F --> J[ECS コンテナ]
    F --> K[RDS データベース]
    F --> L[CloudFront CDN]
    
    style A fill:#e1f5fe
    style B fill:#f3e5f5
    style F fill:#fff3e0
```

このように複雑な構成だからこそ、**確実に動作する環境**を構築する必要があります。

### 実際の開発実績について

本プロジェクトは単なる学習用ではありません。**実際にラズベリーパイ環境で完全開発を行い、AWS本番環境で継続稼働している**実用レベルのWebアプリケーションです。

つまり、このガイドに従うことで、**実際に動作する開発環境**を構築できることが実証されています。

## 推奨環境：なぜこの環境を選ぶのか？

### 環境選択の背景

開発環境選択では、以下の要素を考慮する必要があります：

1. **開発効率**：コードの編集からテストまでの快適さ
2. **本番環境との親和性**：AWS環境（Linux）との類似性
3. **チーム協調**：他の開発者との環境共有のしやすさ
4. **長期保守性**：環境の維持・更新のしやすさ

```mermaid
graph LR
    A[環境選択] --> B[開発効率]
    A --> C[本番親和性]
    A --> D[チーム協調]
    A --> E[長期保守性]
    
    B --> F[WSL]
    C --> F
    D --> F
    E --> F
    
    B --> G[macOS]
    C --> G
    D --> G
    E --> G
    
    B --> H[Ubuntu]
    C --> H
    D --> H
    E --> H
    
    style F fill:#e8f5e8
    style G fill:#e8f5e8
    style H fill:#e8f5e8
```

### 主要推奨環境

#### 1. WSL（Windows Subsystem for Linux）
**対象**：Windows 10/11 ユーザー
**推奨理由**：
- **本番環境と同じLinux環境**で開発できる
- **Windows環境との統合**により、既存の開発ツールも使える
- **VS Code統合**により、シームレスな開発体験

#### 2. macOS
**対象**：macOS 12.0 (Monterey) 以降
**推奨理由**：
- **Unix系OS**により、Linux環境に近い開発体験
- **Homebrew**による依存関係管理が優秀
- **Docker統合**が安定している

#### 3. Ubuntu Desktop
**対象**：Ubuntu 20.04 LTS または 22.04 LTS
**推奨理由**：
- **本番環境と同じOS**での開発
- **最も直接的な開発環境**
- **高いパフォーマンス**

### 参考：実際の開発成功例

**ラズベリーパイでの開発実績**：
- 仕様：Raspberry Pi 4B 8GB、Ubuntu Server 22.04
- 成果：全機能実装・動作確認完了
- 教訓：ARM64環境でも完全動作することを実証

ただし、**初心者にはラズベリーパイは推奨しません**。上記の推奨環境を選択してください。

## システム要件：なぜこのスペックが必要なのか？

### 要件の背景

本プロジェクトは以下の要素を同時に動作させる必要があります：

```mermaid
graph TB
    A[開発環境] --> B[Go アプリケーション]
    A --> C[Node.js 開発サーバー]
    A --> D[PostgreSQL データベース]
    A --> E[Docker コンテナ]
    A --> F[開発ツール群]
    
    B --> G[REST API サーバー]
    B --> H[GraphQL サーバー]
    B --> I[gRPC サーバー]
    
    C --> J[React 開発サーバー]
    C --> K[TypeScript コンパイラ]
    C --> L[ESLint/Prettier]
    
    style A fill:#ffebee
    style B fill:#e8f5e8
    style C fill:#e3f2fd
    style D fill:#fff3e0
```

### 最小要件
- **RAM 4GB以上**：複数サービスの同時動作に必要
- **ディスク容量 10GB以上**：依存関係・データベース・ビルドキャッシュに必要
- **CPU 2コア以上**：並行ビルド・テスト実行に必要

### 推奨要件
- **RAM 8GB以上**：快適な開発体験のため
- **ディスク容量 20GB以上**：本番環境デプロイ用イメージ作成に必要
- **CPU 4コア以上**：並行処理による開発効率向上

## 必要なソフトウェア：各ツールの役割

### なぜこれらのツールが必要なのか？

各ツールは、システム全体の中で重要な役割を果たします：

```mermaid
graph TB
    A[開発者] --> B[エディタ/IDE]
    A --> C[Git]
    A --> D[コンパイラ/ランタイム]
    A --> E[データベース]
    A --> F[コンテナ]
    
    B --> G[コード編集]
    C --> H[バージョン管理]
    D --> I[アプリケーション実行]
    E --> J[データ永続化]
    F --> K[環境再現性]
    
    G --> L[実際のアプリケーション]
    H --> L
    I --> L
    J --> L
    K --> L
    
    style A fill:#e1f5fe
    style L fill:#c8e6c9
```

### コア開発環境

#### Go 1.21以上
**役割**：バックエンドアプリケーションの実行環境
**なぜ必要**：
- REST API、GraphQL、gRPCサーバーの実行
- 高性能・並行処理能力
- 静的型付けによる安全性

#### Node.js 18以上
**役割**：フロントエンド開発環境
**なぜ必要**：
- React アプリケーションの開発サーバー
- TypeScript コンパイル
- パッケージ管理（npm/yarn）

#### PostgreSQL 13以上
**役割**：データベース管理システム
**なぜ必要**：
- アプリケーションデータの永続化
- 本番環境と同じデータベース
- ACID特性による信頼性

#### Docker & Docker Compose
**役割**：コンテナ化・環境再現
**なぜ必要**：
- 本番環境と同じコンテナ実行
- 環境の一貫性確保
- デプロイ前の動作確認

### 開発支援ツール

#### Git 2.30以上
**役割**：バージョン管理
**なぜ必要**：
- コードの履歴管理
- チーム開発での協調
- 本番環境へのデプロイ

#### Make
**役割**：タスク自動化
**なぜ必要**：
- 複雑なビルド・テストコマンドの簡素化
- 開発ワークフローの標準化
- 人的エラーの削減

#### curl & jq
**役割**：API テスト・JSON処理
**なぜ必要**：
- REST API の動作確認
- GraphQL クエリのテスト
- レスポンスの解析

## セットアップ手順：段階的な環境構築

### なぜこの順番で進めるのか？

環境構築は依存関係を考慮して段階的に進める必要があります：

```mermaid
graph TB
    A[Step 1: 基本環境] --> B[Step 2: プロジェクト取得]
    B --> C[Step 3: データベース設定]
    C --> D[Step 4: 環境変数設定]
    D --> E[Step 5: AWS設定]
    E --> F[Step 6: 動作確認]
    F --> G[Step 7: フロントエンド]
    
    A --> H[OS・基本ツール]
    B --> I[ソースコード・依存関係]
    C --> J[データ永続化基盤]
    D --> K[アプリケーション設定]
    E --> L[クラウド連携]
    F --> M[統合テスト]
    G --> N[UI開発環境]
    
    style A fill:#ffebee
    style B fill:#e8f5e8
    style C fill:#e3f2fd
    style D fill:#fff3e0
    style E fill:#fce4ec
    style F fill:#f3e5f5
    style G fill:#e0f2f1
```

### Step 1: 基本環境構築

#### WSL環境での構築

**なぜWSLを選ぶのか？**
- Windows環境でLinux開発環境を使用可能
- 本番環境（AWS Linux）との親和性
- VS Code統合による快適な開発体験

```bash
# WSLの有効化（Windows PowerShell管理者権限で実行）
# なぜ必要：Windows上でLinux環境を使用するため
wsl --install -d Ubuntu-22.04

# WSL Ubuntu内での作業開始
# 以下のコマンドはWSL Ubuntu内で実行

# システム更新（重要：セキュリティ・安定性確保）
sudo apt update && sudo apt upgrade -y

# 基本開発ツールのインストール
# なぜ必要：コンパイル・ビルドに必要な基本ツール
sudo apt install -y build-essential git curl wget jq make

# Go 1.21のインストール
# なぜ1.21以上：プロジェクトで使用している言語機能のため
wget https://golang.org/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Go環境変数の設定
# なぜ必要：Goコマンドをシステム全体で使用可能にするため
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 設定確認
go version
# 期待される出力：go version go1.21.6 linux/amd64
```

**何が起こっているのか？**
1. WSLでLinux環境を作成
2. 基本的な開発ツールをインストール
3. Go言語の実行環境を構築
4. コマンドラインからGoを使用可能にする

#### macOS環境での構築

**なぜHomebrewを使うのか？**
- macOS向けの最も信頼性が高いパッケージマネージャー
- 依存関係の自動解決
- 簡単なアップデート管理

```bash
# Homebrewのインストール（未インストールの場合）
# なぜ必要：macOS用パッケージマネージャーとして最も信頼性が高い
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 基本開発ツールのインストール
# なぜ必要：Git（バージョン管理）、curl（API テスト）、jq（JSON処理）
brew install git curl wget jq make

# Go 1.21のインストール
# なぜ@1.21：特定バージョンを指定して安定性を確保
brew install go@1.21

# Node.js 18のインストール
# なぜ18以上：React 18の最新機能を使用するため
brew install node@18

# PostgreSQLのインストール
# なぜ@13：本番環境と同じバージョンを使用
brew install postgresql@13
brew services start postgresql@13

# Docker Desktopのインストール
# なぜ必要：本番環境と同じコンテナ実行環境
brew install --cask docker
```

**何が起こっているのか？**
1. Homebrewで統一的なパッケージ管理
2. 必要なツールを適切なバージョンでインストール
3. サービスを自動起動設定
4. 開発環境の基盤を構築

### Step 2: プロジェクトのクローンと初期設定

**なぜこの順番で進めるのか？**
- プロジェクトファイルの取得
- 依存関係の解決
- 動作確認の準備

```bash
# プロジェクトのクローン
# なぜ必要：ソースコードと設定ファイルを取得
git clone https://github.com/your-username/go-ddd-ca.git
cd go-ddd-ca

# Go依存関係のインストール
# なぜ必要：プロジェクトで使用している外部ライブラリを取得
go mod download

# 何が起こっているのか確認
go mod verify
# 期待される出力：すべてのモジュールが検証済み

# フロントエンド依存関係のインストール
cd frontend
# なぜ必要：React、TypeScript、その他フロントエンドライブラリを取得
npm install
cd ..

# Makefileの動作確認
# なぜ重要：プロジェクトの標準的なタスクを確認
make help
```

**プロジェクト構造の理解**
```mermaid
graph TB
    A[go-ddd-ca/] --> B[cmd/]
    A --> C[internal/]
    A --> D[frontend/]
    A --> E[infrastructure/]
    A --> F[Makefile]
    
    B --> G[api/]
    B --> H[graphql/]
    B --> I[grpc/]
    
    C --> J[domain/]
    C --> K[usecase/]
    C --> L[infrastructure/]
    
    D --> M[src/]
    D --> N[package.json]
    
    E --> O[terraform/]
    E --> P[docker/]
    
    style A fill:#e1f5fe
    style B fill:#e8f5e8
    style C fill:#fff3e0
    style D fill:#e3f2fd
    style E fill:#fce4ec
```

### Step 3: データベースの初期化

**なぜPostgreSQLを選ぶのか？**
- 本番環境（AWS RDS）と同じデータベース
- ACID特性による信頼性
- 豊富な機能とパフォーマンス

#### PostgreSQL設定の詳細

```bash
# PostgreSQLユーザーの作成
# なぜ必要：アプリケーション専用のデータベースユーザーを作成
sudo -u postgres psql

-- PostgreSQL内での操作
-- なぜtestuserを作成：本番環境と同様の権限分離のため
CREATE USER testuser WITH PASSWORD 'password123';
CREATE DATABASE testdb OWNER testuser;
GRANT ALL PRIVILEGES ON DATABASE testdb TO testuser;

-- 設定確認
\l
-- 期待される出力：testdbがtestuserの所有で表示される
\q
```

**何が起こっているのか？**
1. PostgreSQLに専用ユーザーを作成
2. アプリケーション専用のデータベースを作成
3. 適切な権限を付与
4. セキュリティ原則（最小権限）を実践

#### データベース接続の確認

```bash
# データベース接続テスト
# なぜ重要：アプリケーションからの接続を事前確認
psql -h localhost -U testuser -d testdb -c "SELECT version();"

# 期待される出力例：
# PostgreSQL 13.x on x86_64-pc-linux-gnu...
```

#### スキーマの作成

```bash
# マイグレーションの実行
# なぜ必要：アプリケーションが使用するテーブル構造を作成
make migrate-up

# 何が起こっているのか確認
psql -h localhost -U testuser -d testdb -c "\dt"
# 期待される出力：test_suites, test_cases, test_groups テーブルが表示

# テストデータの投入（オプション）
# なぜ有用：開発時の動作確認用データ
make seed-data

# データベース接続確認
make test-db-connection
```

### Step 4: 環境変数の設定

**なぜ環境変数が重要なのか？**
- セキュリティ：機密情報をコードから分離
- 柔軟性：環境ごとの設定を動的に変更
- 12-Factor App原則：設定の外部化

#### システム設計での環境変数の役割

```mermaid
graph TB
    A[環境変数] --> B[データベース設定]
    A --> C[JWT設定]
    A --> D[AWS設定]
    A --> E[サーバー設定]
    
    B --> F[DB_HOST, DB_PORT, DB_NAME]
    C --> G[JWT_SECRET, JWT_EXPIRATION]
    D --> H[AWS_ACCESS_KEY_ID, AWS_REGION]
    E --> I[HTTP_PORT, GRPC_PORT]
    
    F --> J[アプリケーション]
    G --> J
    H --> J
    I --> J
    
    style A fill:#e1f5fe
    style J fill:#c8e6c9
```

#### 基本環境設定ファイルの作成

```bash
# 環境変数ファイルの作成
# なぜ.env.example：テンプレートファイルを基に安全に作成
cp .env.example .env.local

# .env.localの編集
# なぜ.local：ローカル開発環境専用であることを明示
nano .env.local

# 設定内容の例と説明
cat << 'EOF' > .env.local
# データベース設定
# なぜ必要：アプリケーションがデータベースに接続するため
DB_HOST=localhost
DB_PORT=5432
DB_NAME=testdb
DB_USER=testuser
DB_PASSWORD=password123
DB_SSL_MODE=disable

# サーバー設定
# なぜ必要：各サービスが使用するポート番号を指定
HTTP_PORT=8080
GRPC_PORT=50051
GRAPHQL_PATH=/graphql

# JWT設定
# なぜ必要：ユーザー認証のためのトークン生成・検証
JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRATION=24h

# ログ設定
# なぜ必要：開発時の問題特定・デバッグ情報
LOG_LEVEL=debug
LOG_FORMAT=json
EOF
```

**各設定の意味**
- `DB_*`：データベース接続情報
- `HTTP_PORT`：REST API・GraphQLサーバーのポート
- `GRPC_PORT`：gRPCサーバーのポート
- `JWT_SECRET`：認証トークンの暗号化キー
- `LOG_LEVEL`：ログの詳細度（debug=最詳細）

#### 環境変数の読み込み確認

```bash
# 環境変数の確認
# なぜ重要：設定が正しく読み込まれているか確認
make check-env

# 期待される出力例：
# DB_HOST: localhost
# DB_PORT: 5432
# JWT_SECRET: [設定済み]

# 設定ファイルの検証
make validate-config
```

### Step 5: AWS環境変数の設定（最重要）

**なぜAWS環境変数が重要なのか？**
- 本番環境（AWS）への デプロイに必須
- AWS CLIが自動的に参照する固定の環境変数名
- セキュリティ：アクセスキーの適切な管理

#### AWS CLI環境変数の仕組み

AWS CLIは以下の**固定の環境変数名**を自動参照します：

```mermaid
graph TB
    A[AWS CLI] --> B[環境変数自動参照]
    B --> C[AWS_ACCESS_KEY_ID]
    B --> D[AWS_SECRET_ACCESS_KEY]
    B --> E[AWS_REGION]
    
    C --> F[AWSアクセスキーID]
    D --> G[AWSシークレットアクセスキー]
    E --> H[AWSリージョン]
    
    F --> I[AWS API認証]
    G --> I
    H --> I
    
    I --> J[terraform apply]
    I --> K[aws sts get-caller-identity]
    I --> L[ECR push]
    
    style A fill:#ff9800
    style I fill:#4caf50
    style J fill:#2196f3
    style K fill:#2196f3
    style L fill:#2196f3
```

#### ~/.env.terraform設定ファイルの作成

**なぜこのファイルが必要なのか？**
- AWS認証情報を一元管理
- 新しいシェルセッションでも自動的に利用可能
- セキュリティ設定（権限600）による保護

```bash
# ~/.env.terraform ファイルの作成
touch ~/.env.terraform

# ファイル権限の設定（重要：機密情報保護）
# なぜ600：所有者のみ読み書き可能、他者はアクセス不可
chmod 600 ~/.env.terraform

# 設定内容の例
cat << 'EOF' > ~/.env.terraform
# AWS認証情報（必須）
# なぜ必要：AWS APIへの認証に使用
export AWS_ACCESS_KEY_ID=AKIAXXXXXXXXXXXXXXXX
export AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
export AWS_REGION=ap-northeast-1

# Terraform変数
# なぜ TF_VAR_：Terraformが自動的に認識する変数プレフィックス
export TF_VAR_db_username="testadmin"
export TF_VAR_db_password="SecurePassword123"

# アプリケーション設定
export JWT_SECRET="your-jwt-secret-key-change-in-production"
export GRAPHQL_URL="https://your-graphql-endpoint.com/graphql"

# 環境識別
export APP_ENVIRONMENT="development"
export TF_VAR_environment="development"

# CloudFront設定
export CLOUDFRONT_DISTRIBUTION_ID="EXXXXXXXXXXXXX"
export S3_BUCKET_NAME="your-frontend-bucket"

# RDS設定
export TF_VAR_db_instance_class="db.t3.micro"
export TF_VAR_db_allocated_storage="20"
export TF_VAR_db_backup_retention_period="7"

# ECS設定
export TF_VAR_ecs_desired_count="1"
export TF_VAR_ecs_cpu="256"
export TF_VAR_ecs_memory="512"
EOF
```

**各設定の意味**
- `AWS_ACCESS_KEY_ID`：AWSアカウントのアクセスキーID
- `AWS_SECRET_ACCESS_KEY`：対応するシークレットキー
- `AWS_REGION`：使用するAWSリージョン
- `TF_VAR_*`：Terraformが使用する変数

#### 環境変数の永続化設定

**なぜ永続化が必要なのか？**
- 新しいターミナルセッションで自動的に利用可能
- デプロイコマンド実行時に手動設定が不要
- 開発効率の向上

```bash
# ~/.bashrc に自動読み込み設定を追加
echo '# AWS Terraform環境変数の自動読み込み' >> ~/.bashrc
echo '[ -f ~/.env.terraform ] && source ~/.env.terraform' >> ~/.bashrc

# 設定の反映
source ~/.bashrc

# 永続化確認
echo "AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID"
echo "AWS_REGION: $AWS_REGION"
echo "TF_VAR_db_username: $TF_VAR_db_username"

# 新しいシェルでの確認
bash -c 'echo "新しいシェル - AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID"'
```

#### デプロイワークフローと環境変数の関係

**なぜこの流れを理解する必要があるのか？**
- エラー発生時の原因特定
- 環境変数設定の重要性理解
- デプロイプロセスの透明性確保

```mermaid
graph TD
    A[make deploy-*] --> B[prepare-ecr-image.sh]
    B --> C[aws sts get-caller-identity]
    C --> D{環境変数チェック}
    D -->|設定済み| E[AWS認証成功]
    D -->|未設定| F[エラー: AWS認証情報が無効]
    
    E --> G[Docker イメージビルド]
    F --> G1[スクリプト停止]
    G --> H[ECR プッシュ]
    H --> I[Terraform デプロイ]
    I --> J[update-tfvars.sh]
    J --> K[terraform apply]
    K --> L[AWS環境構築完了]
    
    style D fill:#f9f,stroke:#333,stroke-width:2px
    style F fill:#f96,stroke:#333,stroke-width:2px
    style E fill:#9f9,stroke:#333,stroke-width:2px
    style G1 fill:#f96,stroke:#333,stroke-width:2px
    style L fill:#9f9,stroke:#333,stroke-width:2px
```

**プロセスの詳細**
1. `make deploy-*`：デプロイコマンド実行
2. `prepare-ecr-image.sh`：ECR準備スクリプト起動
3. `aws sts get-caller-identity`：AWS認証確認
4. 環境変数チェック：設定済みかどうか判定
5. 成功時：Docker→ECR→Terraform→AWS環境構築
6. 失敗時：エラーメッセージ表示・スクリプト停止

#### 環境変数設定の検証

```bash
# AWS認証の確認
aws sts get-caller-identity

# 期待される出力例
{
    "UserId": "AIDAXXXXXXXXXXXXXXXX",
    "Account": "123456789012",
    "Arn": "arn:aws:iam::123456789012:user/terraform-user"
}

# エラーの場合の対処
# エラー例：Unable to locate credentials
# 解決方法：~/.env.terraform の設定を確認

# Terraform変数の確認
echo "DB Username: $TF_VAR_db_username"
echo "Environment: $TF_VAR_environment"

# 新しいシェルでの確認
bash -c 'echo "新しいシェル - AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID"'
```

#### セキュリティ設定の重要事項

**なぜセキュリティが重要なのか？**
- AWS認証情報の漏洩防止
- 不正アクセスの防止
- Gitリポジトリへの機密情報混入防止

```bash
# 1. ファイル権限の確認
ls -la ~/.env.terraform
# 期待される出力: -rw------- 1 username username

# 2. Git除外設定の確認
echo "/.env.terraform" >> ~/.gitignore
echo "/.env.*" >> ~/.gitignore

# 3. 権限確認
ls -la ~/.env.terraform
# 600以外の場合は修正
chmod 600 ~/.env.terraform

# 4. 機密情報の分離
# 本番環境では直接環境変数を設定し、~/.env.terraformファイルは使用しない
```

### Step 6: 各サービスの起動と動作確認

**なぜ段階的に起動するのか？**
- 各サービスの動作を個別に確認
- 問題発生時の原因特定を容易にする
- システム全体の理解を深める

#### サービス起動の全体像

```mermaid
graph TB
    A[開発者] --> B[make run-api]
    A --> C[make run-graphql]
    A --> D[make run-grpc]
    
    B --> E[REST API Server]
    C --> F[GraphQL Server]
    D --> G[gRPC Server]
    
    E --> H[localhost:8080]
    F --> I[localhost:8080/graphql]
    G --> J[localhost:50051]
    
    H --> K[PostgreSQL]
    I --> K
    J --> K
    
    style E fill:#4caf50
    style F fill:#2196f3
    style G fill:#ff9800
    style K fill:#9c27b0
```

#### 個別サービスの起動

```bash
# REST APIサービスの起動
# なぜ最初：最もシンプルなサービスから確認
make run-api

# 期待される出力：
# Starting REST API server on port 8080...
# Database connection established
# Server started successfully

# 別のターミナルで動作確認
curl http://localhost:8080/health
# 期待される応答: {"status":"UP","timestamp":"2024-01-15T10:30:00Z"}
```

**何が起こっているのか？**
1. REST APIサーバーが起動
2. データベースへの接続確立
3. ヘルスチェックエンドポイントが応答

```bash
# 別のターミナルでGraphQLサービスの起動
make run-graphql

# 期待される出力：
# Starting GraphQL server on port 8080...
# GraphQL Playground available at http://localhost:8080/graphql
# Server started successfully

# GraphQLの確認
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __schema { queryType { name } } }"}'

# 期待される応答：GraphQLスキーマ情報が返される
```

**GraphQL Playgroundの使用**
- ブラウザで`http://localhost:8080/graphql`にアクセス
- インタラクティブなクエリツールが表示
- スキーマ情報・ドキュメントが利用可能

```bash
# 別のターミナルでgRPCサービスの起動
make run-grpc

# 期待される出力：
# Starting gRPC server on port 50051...
# gRPC reflection enabled
# Server started successfully

# gRPCの確認（grpcurlが必要）
grpcurl -plaintext localhost:50051 list
# 期待される応答：利用可能なサービス一覧が表示
```

#### 統合テストの実行

**なぜ統合テストが重要なのか？**
- 各サービスの連携動作確認
- データベースとの統合動作確認
- 本番環境と同様の動作確認

```bash
# すべてのサービスの統合テスト
make test-integration

# 期待される出力：
# Running integration tests...
# ✓ Database connection test
# ✓ REST API endpoints test
# ✓ GraphQL queries test
# ✓ gRPC service test
# All tests passed!

# 個別サービスのテスト
make test-api
make test-graphql
make test-grpc
```

**テスト内容の理解**
- データベース接続：CRUD操作の確認
- API エンドポイント：各種リクエストの確認
- 認証機能：JWT トークンの生成・検証
- エラーハンドリング：異常系の動作確認

### Step 7: フロントエンドの起動

**なぜフロントエンドが必要なのか？**
- バックエンドAPIの動作確認
- ユーザーインターフェースの提供
- 実際のWebアプリケーション体験

#### フロントエンド技術スタック

```mermaid
graph TB
    A[React 19.1.0] --> B[TypeScript 5.8.3]
    A --> C[Apollo Client 3.13.8]
    A --> D[Material UI 7.1.1]
    
    B --> E[型安全性]
    C --> F[GraphQL統合]
    D --> G[UIコンポーネント]
    
    E --> H[フロントエンドアプリケーション]
    F --> H
    G --> H
    
    H --> I[localhost:3000]
    
    style A fill:#61dafb
    style B fill:#3178c6
    style C fill:#311c87
    style D fill:#0081cb
    style H fill:#4caf50
```

#### 開発サーバーの起動

```bash
# フロントエンドディレクトリに移動
cd frontend

# 依存関係の確認
npm list --depth=0
# 期待される出力：主要パッケージのバージョン一覧

# 開発サーバーの起動
npm run dev

# 期待される出力：
# Local:   http://localhost:3000
# Network: http://192.168.1.100:3000
# Ready in 1.2s
```

**何が起こっているのか？**
1. React開発サーバーが起動
2. TypeScriptコンパイラーが監視開始
3. Hot Module Replacement（HMR）が有効化
4. GraphQL Code Generatorが型生成

#### フロントエンドの動作確認

```bash
# TypeScriptの型チェック
# なぜ重要：コンパイル時エラーの早期発見
npm run type-check

# 期待される出力：
# Type checking...
# No TypeScript errors found

# ビルドテスト
# なぜ重要：本番環境でのビルド可能性確認
npm run build

# 期待される出力：
# Building for production...
# ✓ 45 modules transformed
# Build completed successfully

# ESLint実行
# なぜ重要：コード品質・一貫性の確保
npm run lint

# テスト実行
# なぜ重要：フロントエンドロジックの動作確認
npm test
```

#### ブラウザでのアクセス確認

**アクセスURL**：`http://localhost:3000`

**期待される画面**：
- ログイン画面が表示
- demo_user/demo_password でログイン可能
- テストスイート管理画面が表示
- GraphQL APIとの連携動作確認

## 開発ワークフロー：日常的な開発作業

### なぜワークフローが重要なのか？

効率的な開発には、標準化された作業手順が必要です：

```mermaid
graph TB
    A[開発開始] --> B[環境起動]
    B --> C[コード編集]
    C --> D[テスト実行]
    D --> E[動作確認]
    E --> F[コミット]
    F --> G[プッシュ]
    
    C --> H[フォーマット]
    D --> I[静的解析]
    E --> J[統合テスト]
    
    style A fill:#e1f5fe
    style F fill:#4caf50
    style G fill:#4caf50
```

### 日常的な開発作業

#### 1. 開発環境の起動

```bash
# バックエンドサービスの起動
# なぜ必要：API サーバーの起動
make dev-backend

# 期待される出力：
# Starting development environment...
# ✓ Database connection established
# ✓ REST API server started (port 8080)
# ✓ GraphQL server started (port 8080)
# ✓ gRPC server started (port 50051)

# フロントエンドの起動（別ターミナル）
make dev-frontend

# 期待される出力：
# Starting frontend development server...
# ✓ TypeScript compilation successful
# ✓ Development server started (port 3000)

# データベースの起動確認
make check-db-status
```

#### 2. 開発中の確認作業

```bash
# バックエンドテストの実行
# なぜ頻繁に：早期のバグ発見・修正
make test

# 期待される出力：
# Running Go tests...
# ✓ domain layer tests
# ✓ usecase layer tests
# ✓ infrastructure layer tests
# All tests passed!

# フロントエンドテストの実行
cd frontend && npm test

# コードフォーマットの実行
# なぜ重要：コードの一貫性確保
make format

# 静的解析の実行
# なぜ重要：潜在的な問題の早期発見
make lint
```

#### 3. 機能開発後の確認

```bash
# 統合テストの実行
# なぜ重要：システム全体の動作確認
make test-integration

# フロントエンド・バックエンド統合確認
make test-e2e

# パフォーマンステスト
make test-performance
```

### コミット前のチェックリスト

**なぜチェックリストが必要なのか？**
- 品質の一貫性確保
- 他の開発者への影響最小化
- 本番環境での問題防止

```bash
# 1. テストの実行
make test-all
# 期待される結果：すべてのテストが成功

# 2. コードフォーマットの確認
make format-check
# 期待される結果：フォーマットエラーなし

# 3. 静的解析の実行
make lint-all
# 期待される結果：lintエラーなし

# 4. ビルドの確認
make build-all
# 期待される結果：ビルドエラーなし

# 5. 統合テストの実行
make test-integration
# 期待される結果：統合テストが成功
```

## トラブルシューティング：問題発生時の対処方法

### なぜトラブルシューティングが重要なのか？

開発中には様々な問題が発生します。重要なのは：
- 問題の原因を理解すること
- 解決方法を学ぶこと
- 同じ問題を回避する方法を身につけること

### よくある問題と解決方法

#### 1. PostgreSQL接続エラー

**エラー例**：
```
connection refused
authentication failed
```

**原因と解決方法**：

```bash
# PostgreSQLサービスの確認
# なぜ重要：サービスが起動しているか確認
sudo systemctl status postgresql

# 期待される出力：
# ● postgresql.service - PostgreSQL RDBMS
#    Loaded: loaded
#    Active: active (running)

# サービスが停止している場合の対処
sudo systemctl restart postgresql

# 接続テスト
psql -h localhost -U testuser -d testdb -c "SELECT 1;"

# パスワードエラーの場合
sudo -u postgres psql -c "ALTER USER testuser PASSWORD 'password123';"
```

**根本原因の理解**：
- PostgreSQLサービスの停止
- ユーザー認証情報の不一致
- データベースの存在確認

#### 2. ポート使用エラー

**エラー例**：
```
port already in use
bind: address already in use
```

**原因と解決方法**：

```bash
# ポート使用状況の確認
# なぜ重要：どのプロセスがポートを使用しているか特定
sudo netstat -tlnp | grep :8080

# 期待される出力例：
# tcp6  0  0 :::8080  :::*  LISTEN  12345/main

# プロセスの終了
kill -9 12345
# または
kill -9 $(lsof -t -i:8080)

# またはmakeコマンドでクリーンアップ
make clean-ports
```

**根本原因の理解**：
- 前回の実行プロセスが残っている
- 異なるサービスが同じポートを使用
- システムの再起動後のポート衝突

#### 3. Go依存関係エラー

**エラー例**：
```
go: module not found
checksum mismatch
```

**原因と解決方法**：

```bash
# Go モジュールキャッシュのクリア
# なぜ必要：破損したキャッシュの削除
go clean -modcache

# 依存関係の再取得
go mod download
go mod tidy

# Goバージョンの確認
go version
# 期待される出力：go version go1.21.x

# プロキシ設定の確認
go env GOPROXY
# 期待される出力：https://proxy.golang.org,direct
```

**根本原因の理解**：
- モジュールキャッシュの破損
- ネットワーク接続の問題
- Goバージョンの不一致

#### 4. Node.js依存関係エラー

**エラー例**：
```
npm ERR! Cannot resolve dependency
npm ERR! peer dependency warning
```

**原因と解決方法**：

```bash
# Node.jsバージョンの確認
node --version
npm --version

# 期待される出力：
# v18.x.x
# 9.x.x

# npm キャッシュのクリア
npm cache clean --force

# node_modules の削除と再インストール
rm -rf node_modules package-lock.json
npm install
```

**根本原因の理解**：
- Node.jsバージョンの不一致
- 依存関係の競合
- キャッシュの破損

#### 5. Docker関連エラー

**エラー例**：
```
Cannot connect to the Docker daemon
permission denied
```

**原因と解決方法**：

```bash
# Dockerサービスの確認
sudo systemctl status docker

# Dockerの再起動
sudo systemctl restart docker

# ユーザーのDockerグループ追加確認
groups $USER
# docker グループが含まれているか確認

# 含まれていない場合
sudo usermod -aG docker $USER
# ログアウト・ログインまたは再起動が必要
```

**根本原因の理解**：
- Dockerデーモンの停止
- ユーザー権限の不足
- グループ設定の問題

#### 6. AWS認証エラー

**エラー例**：
```
AWS認証情報が無効または設定されていません
Unable to locate credentials
```

**原因**：環境変数（AWS_ACCESS_KEY_ID、AWS_SECRET_ACCESS_KEY、AWS_REGION）が未設定

**解決方法**：

```bash
# 環境変数の確認
echo "AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID"
echo "AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY"
echo "AWS_REGION: $AWS_REGION"

# 空の場合の対処
# ~/.env.terraform ファイルの確認
ls -la ~/.env.terraform
cat ~/.env.terraform

# 永続化設定の確認
grep -n "env.terraform" ~/.bashrc

# ~/.env.terraform が存在しない場合
# Step 5の「AWS環境変数の設定」を実行

# 新しいシェルでの確認
bash -c 'echo "新しいシェル - AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID"'

# AWS認証の検証
aws sts get-caller-identity
```

**根本原因の理解**：
- 環境変数の設定忘れ
- 永続化設定の不備
- ファイルパス・権限の問題

#### 7. Terraform関連エラー

**エラー例**：
```
terraform init failed
Error: Backend initialization required
```

**原因と解決方法**：

```bash
# Terraform状態の確認
terraform show

# 環境変数の確認
env | grep TF_VAR

# Terraform初期化
terraform init

# プラン確認
terraform plan

# 状態管理の問題の場合
terraform refresh
```

**根本原因の理解**：
- Terraform状態ファイルの不整合
- 環境変数の設定不備
- リモートステートの問題

## 環境別の注意事項

### WSL環境

**特有の問題と対処**：

```bash
# WSL2の使用を推奨
# なぜ重要：パフォーマンスと互換性の向上
wsl --set-version Ubuntu-22.04 2

# Windows側からのファイルアクセス
# 避けるべき：/mnt/c/Users/...
# 推奨：/home/... を使用

# VS Codeの連携
# Remote - WSL拡張機能を使用
code .
```

### macOS環境

**特有の問題と対処**：

```bash
# Homebrew の定期更新
brew update && brew upgrade

# macOS Ventura以降でのDockerパフォーマンス
# Settings > Advanced > VirtioFS を有効化
```

### Ubuntu環境

**特有の問題と対処**：

```bash
# ファイアウォールの設定
sudo ufw allow 8080
sudo ufw allow 50051
sudo ufw allow 3000

# 権限の設定
sudo chown -R $USER:$USER /home/$USER/go
```

## パフォーマンス最適化

### 開発環境のパフォーマンス向上

#### 1. Go開発の最適化

```bash
# Go モジュールキャッシュの活用
export GOCACHE=/tmp/go-cache
export GOMODCACHE=/tmp/go-mod-cache

# ビルドキャッシュの確認
go env GOCACHE
go env GOMODCACHE

# 並列ビルドの最適化
export GOMAXPROCS=4
```

#### 2. Node.js開発の最適化

```bash
# npm レジストリの最適化
npm config set registry https://registry.npmjs.org/

# 並列インストールの最適化
npm config set maxsockets 50

# キャッシュの最適化
npm config set cache /tmp/npm-cache
```

#### 3. PostgreSQL の最適化

```bash
# postgresql.conf の調整
sudo nano /etc/postgresql/13/main/postgresql.conf

# 推奨設定
shared_buffers = 256MB
effective_cache_size = 1GB
max_connections = 100
```

## セキュリティ考慮事項

### 機密情報の管理

#### 1. 環境変数ファイルの保護

```bash
# ファイル権限の設定
chmod 600 ~/.env.terraform

# 権限の確認
ls -la ~/.env.terraform
# 期待される出力: -rw------- 1 username username

# Git除外設定
echo "/.env.terraform" >> ~/.gitignore
echo "/.env.*" >> ~/.gitignore
echo "/config/development.yml" >> ~/.gitignore
```

#### 2. 本番環境との設定分離

```bash
# 開発環境専用の設定
export APP_ENVIRONMENT=development
export DB_SSL_MODE=disable
export LOG_LEVEL=debug

# 本番環境では以下の設定を使用
# export APP_ENVIRONMENT=production
# export DB_SSL_MODE=require
# export LOG_LEVEL=info
```

#### 3. 認証情報のローテーション

```bash
# 定期的な認証情報の更新
# 1. AWSアクセスキーの更新
# 2. データベースパスワードの更新
# 3. JWT シークレットの更新

# 更新後の確認
aws sts get-caller-identity
psql -h localhost -U testuser -d testdb -c "SELECT 1;"
```

## AWS デプロイ準備

### AWS 環境へのデプロイ

#### 1. 環境変数の確認

```bash
# 必要な環境変数の確認
make check-aws-env

# 期待される出力例
AWS_ACCESS_KEY_ID: AKIAXXXXXXXXXXXXXXXX
AWS_SECRET_ACCESS_KEY: [設定済み]
AWS_REGION: ap-northeast-1
TF_VAR_db_username: testadmin
TF_VAR_environment: development
```

#### 2. Terraform初期化

```bash
# Terraform初期化
make terraform-init

# プラン確認
make terraform-plan

# 実際のデプロイ
make deploy-dev
```

#### 3. デプロイ確認

```bash
# デプロイされたサービスの確認
make check-deployment

# GraphQLエンドポイントの確認
curl -X POST https://your-graphql-endpoint.com/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __schema { queryType { name } } }"}'
```

## まとめ

このガイドは、単なる「コマンドの羅列」ではなく、**なぜそのコマンドが必要なのか**、**どのような効果があるのか**を理解しながら進められるよう設計しました。

### 重要なポイント

1. **実証済みの手順**：実際の開発成功例に基づく確実な手順
2. **理解重視**：コマンドの背景・理由を併記
3. **視覚的理解**：Mermaidダイアグラムによる構造の可視化
4. **親切な説明**：初心者でも理解できる詳細な解説
5. **実用性**：他の開発者が実際に使用できる内容

### 環境変数設定の重要性

特に、AWS環境変数の設定は、デプロイ成功の鍵となります：

- **AWS CLI固定環境変数**：`AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY`、`AWS_REGION`
- **永続化設定**：`~/.bashrc`での`~/.env.terraform`自動読み込み
- **セキュリティ**：ファイル権限600、Git除外設定、本番環境分離

### 学習価値

このガイドを通じて、以下の能力を身につけることができます：

1. **システム理解**：各コンポーネントの役割と関係性
2. **問題解決力**：トラブル発生時の原因特定と対処
3. **セキュリティ意識**：機密情報の適切な管理
4. **開発効率**：標準化されたワークフロー

### 次のステップ

環境構築が完了したら、以下のドキュメントを参照して開発を進めてください：

- **AWS環境構築**：本番環境へのデプロイ手順
- **アーキテクチャ解説**：プロジェクトの技術的詳細
- **セキュリティ移行ガイド**：本番環境での適切なセキュリティ設定

### サポート

環境構築で問題が発生した場合は、以下を確認してください：

1. **トラブルシューティング**セクションの該当する解決方法
2. **環境別の注意事項**での特有の対応
3. **GitHub Issues**での既知の問題と解決方法

このガイドが、Go-DDD-CA プロジェクトでの開発体験を向上させることを願っています。開発を楽しんでください！