# AWS CI/CDパイプライン設計書

## 1. 概要
本書では、テストケース管理システムのための継続的インテグレーション/継続的デリバリー(CI/CD)パイプラインの設計について記述します。GitLabのCI/CD機能とAWSサービスを連携させ、自動テスト、ビルド、デプロイを実現します。

## 2. 全体アーキテクチャ
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  GitLab     │    │  GitLab CI  │    │  AWS ECR    │    │  AWS ECS    │
│ Repository  │───►│   Runner    │───►│ (Container  │───►│  Fargate    │
└─────────────┘    └─────────────┘    │  Registry)  │    └─────────────┘
                          │           └─────────────┘           ▲
                          │                                     │
                          ▼                                     │
                   ┌─────────────┐                      ┌───────────────┐
                   │  Terraform  │                      │  AWS Secrets  │
                   │    State    │─────────────────────►│   Manager     │
                   └─────────────┘                      └───────────────┘
```

## 3. パイプラインステージ

### 3.1 ステージの概要
1. **コード検証** - 静的解析、コードスタイルチェック
2. **ユニットテスト** - コンポーネント単位のテスト実行
3. **統合テスト** - コンポーネント間連携のテスト
4. **イメージビルド** - Dockerイメージの構築とタグ付け
5. **インフラ検証** - Terraformプランの作成と検証
6. **デプロイ** - ターゲット環境へのリリース
7. **検証テスト** - デプロイ後の正常性確認

### 3.2 環境分離
- **Development** - 開発環境（自動デプロイ）
- **Staging** - ステージング環境（手動承認後デプロイ）
- **Production** - 本番環境（手動承認後デプロイ）

## 4. GitLab CI/CD設定

### 4.1 GitLab CI/CDパイプライン定義
```yaml
# .gitlab-ci.yml

stages:
  - validate
  - test
  - build
  - infra
  - deploy
  - verify

variables:
  AWS_DEFAULT_REGION: ap-northeast-1
  ECR_REPOSITORY: test-management-app
  TF_STATE_BUCKET: terraform-state-test-management

# コード検証ステージ
code-lint:
  stage: validate
  image: golangci/golangci-lint:latest
  script:
    - golangci-lint run ./...

# テストステージ
unit-tests:
  stage: test
  image: golang:1.21
  script:
    - make test

integration-tests:
  stage: test
  image: golang:1.21
  services:
    - postgres:14.13
  variables:
    POSTGRES_USER: testuser
    POSTGRES_PASSWORD: testpass
    POSTGRES_DB: test_db
  script:
    - make test-integration

# イメージビルドステージ
build-image:
  stage: build
  image: docker:20.10
  services:
    - docker:20.10-dind
  variables:
    DOCKER_TLS_CERTDIR: "/certs"
  script:
    - echo "$AWS_ECR_PASSWORD" | docker login -u $AWS_ECR_USER --password-stdin $AWS_ECR_REGISTRY
    - docker build -t $AWS_ECR_REGISTRY/$ECR_REPOSITORY:$CI_COMMIT_SHORT_SHA .
    - docker tag $AWS_ECR_REGISTRY/$ECR_REPOSITORY:$CI_COMMIT_SHORT_SHA $AWS_ECR_REGISTRY/$ECR_REPOSITORY:latest
    - docker push $AWS_ECR_REGISTRY/$ECR_REPOSITORY:$CI_COMMIT_SHORT_SHA
    - docker push $AWS_ECR_REGISTRY/$ECR_REPOSITORY:latest

# インフラ検証ステージ
terraform-plan:
  stage: infra
  image: hashicorp/terraform:1.5
  script:
    - cd terraform/environments/${CI_ENVIRONMENT_NAME}
    - terraform init -backend-config="bucket=${TF_STATE_BUCKET}" -backend-config="key=${CI_ENVIRONMENT_NAME}/terraform.tfstate"
    - terraform plan -out=tfplan
  artifacts:
    paths:
      - terraform/environments/${CI_ENVIRONMENT_NAME}/tfplan

# デプロイステージ
deploy:
  stage: deploy
  image: hashicorp/terraform:1.5
  script:
    - cd terraform/environments/${CI_ENVIRONMENT_NAME}
    - terraform init -backend-config="bucket=${TF_STATE_BUCKET}" -backend-config="key=${CI_ENVIRONMENT_NAME}/terraform.tfstate"
    - terraform apply -auto-approve tfplan
    - aws ecs update-service --cluster test-management-cluster --service test-management-service --force-new-deployment
  dependencies:
    - terraform-plan
  environment:
    name: ${CI_ENVIRONMENT_NAME}
  when: manual
  only:
    - main

# デプロイ検証ステージ
verify-deployment:
  stage: verify
  image: alpine:latest
  script:
    - apk add --no-cache curl
    - curl -f https://${CI_ENVIRONMENT_NAME}-api.example.com/health || exit 1
  dependencies:
    - deploy
  environment:
    name: ${CI_ENVIRONMENT_NAME}
```

### 4.2 環境別の設定
GitLab CI/CDの環境変数を使用して、環境ごとの設定を管理します：

| 環境変数 | 開発環境 | ステージング環境 | 本番環境 |
|---------|----------|----------------|---------|
| `CI_ENVIRONMENT_NAME` | `development` | `staging` | `production` |
| `ECS_CLUSTER` | `test-mgmt-dev` | `test-mgmt-stg` | `test-mgmt-prod` |
| `ECS_SERVICE` | `app-service-dev` | `app-service-stg` | `app-service-prod` |

## 5. AWS連携設定

### 5.1 AWS認証設定
GitLab CI/CDからAWSリソースにアクセスするための認証設定:

#### 5.1.1 IAMユーザー方式
GitLab CI/CD変数に以下を設定:
- `AWS_ACCESS_KEY_ID`: IAMユーザーのアクセスキー
- `AWS_SECRET_ACCESS_KEY`: IAMユーザーのシークレットキー
- `AWS_DEFAULT_REGION`: デフォルトリージョン

#### 5.1.2 IAMロール方式（推奨）
GitLabから一時的な認証情報を取得してAWSリソースにアクセス:

```yaml
assume-role:
  script:
    - >
      CREDENTIALS=$(aws sts assume-role 
      --role-arn arn:aws:iam::123456789012:role/GitLabCICDRole 
      --role-session-name GitLabSession)
    - export AWS_ACCESS_KEY_ID=$(echo $CREDENTIALS | jq -r '.Credentials.AccessKeyId')
    - export AWS_SECRET_ACCESS_KEY=$(echo $CREDENTIALS | jq -r '.Credentials.SecretAccessKey')
    - export AWS_SESSION_TOKEN=$(echo $CREDENTIALS | jq -r '.Credentials.SessionToken')
```

### 5.2 ECRへのイメージプッシュ設定
```yaml
build-and-push:
  script:
    - aws ecr get-login-password | docker login --username AWS --password-stdin $AWS_ECR_REGISTRY
    - docker build -t $AWS_ECR_REGISTRY/$ECR_REPOSITORY:$CI_COMMIT_SHORT_SHA .
    - docker push $AWS_ECR_REGISTRY/$ECR_REPOSITORY:$CI_COMMIT_SHORT_SHA
```

### 5.3 ECSサービス更新設定
```yaml
deploy-to-ecs:
  script:
    - aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE --force-new-deployment
```

## 6. Terraform状態管理

### 6.1 S3バックエンド設定
```hcl
terraform {
  backend "s3" {
    bucket = "terraform-state-test-management"
    key    = "env/terraform.tfstate"
    region = "ap-northeast-1"
    encrypt = true
    dynamodb_table = "terraform-state-lock"
  }
}
```

### 6.2 DynamoDBロック設定
```hcl
resource "aws_dynamodb_table" "terraform_locks" {
  name         = "terraform-state-lock"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}
```

## 7. シークレット管理

### 7.1 AWS Secrets Managerの活用
環境ごとの機密情報をAWS Secrets Managerで管理し、アプリケーション起動時に取得:

```yaml
deploy:
  script:
    - >
      aws ecs run-task 
      --cluster $ECS_CLUSTER 
      --task-definition $TASK_DEFINITION 
      --overrides '{
        "containerOverrides": [{
          "name": "app",
          "environment": [
            { "name": "SECRETS_ARN", "value": "arn:aws:secretsmanager:region:account:secret/app/env" }
          ]
        }]
      }'
```

### 7.2 GitLab環境変数
GitLabのCI/CD変数機能を使用して、環境ごとの変数を管理:

- グループレベル変数: 全プロジェクト共通の変数
- プロジェクトレベル変数: プロジェクト固有の変数
- 環境固有の変数: 特定環境専用の変数

## 8. パイプラインのセキュリティと最適化

### 8.1 セキュリティ対策
- GitLab Runnerのセキュリティ対策
  - プライベートランナーの使用
  - イメージの脆弱性スキャン
- 最小権限原則の適用
  - 環境ごとに必要最小限のIAM権限
- シークレットの安全な管理
  - 環境変数のマスキング
  - AWS Secrets Managerの活用

### 8.2 パイプライン最適化
- キャッシュ設定
  ```yaml
  cache:
    key: ${CI_COMMIT_REF_SLUG}
    paths:
      - .go/pkg/mod/
  ```

- 並列実行
  ```yaml
  test-suite1:
    stage: test
    parallel: 3
  ```

- ジョブの条件付き実行
  ```yaml
  deploy:
    rules:
      - if: '$CI_COMMIT_BRANCH == "main"'
        when: manual
  ```

## 9. パイプライン運用とモニタリング

### 9.1 失敗時の通知設定
```yaml
stages:
  - build
  - test
  - deploy

build:
  stage: build
  script:
    - echo "Building..."

test:
  stage: test
  script:
    - echo "Testing..."

deploy:
  stage: deploy
  script:
    - echo "Deploying..."

notify:
  stage: .post
  script:
    - echo "Sending notification..."
  rules:
    - when: on_failure
```

### 9.2 デプロイ履歴とロールバック戦略
- ECSのデプロイメント履歴を追跡
- 問題発生時の迅速なロールバック手順
  ```bash
  aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE --task-definition $PREVIOUS_TASK_DEF
  ```

## 10. 今後の拡張ポイント

- マルチリージョンデプロイメント対応
- ブルー/グリーンデプロイメント戦略の導入
- カナリアリリースの実装
- 自動ロールバックメカニズムの強化
- デプロイメントメトリクスの詳細分析