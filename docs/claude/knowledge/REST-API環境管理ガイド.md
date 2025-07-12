# REST API環境管理ガイド

## 1. 標準デプロイフロー

### 1.1 開発ライフサイクル管理

REST API環境の標準的なライフサイクルは以下の通りです：

1. **開発開始（start-api-dev）**:
   - コア環境（VPC、RDS、ECSクラスター）のデプロイ
   - REST APIサービスのデプロイ
   - ヘルスチェック検証

2. **開発一時停止（pause-api-dev）**:
   - APIサービスとALBの削除
   - コア環境の維持

3. **開発再開（resume-api-dev）**:
   - APIサービスとALBの再デプロイ
   - ヘルスチェック検証

4. **開発完了（stop-api-dev）**:
   - すべてのリソースの削除

### 1.2 シーケンス図

```
┌─────────┐          ┌──────────┐         ┌──────────┐        ┌─────────┐
│  開発者  │          │ workflow │         │ terraform│        │  AWS    │
└────┬────┘          └─────┬────┘         └─────┬────┘        └────┬────┘
     │                     │                    │                   │
     │ start-api-dev       │                    │                   │
     ├────────────────────>│                    │                   │
     │                     │ terraform apply    │                   │
     │                     ├──────────────────>│                    │
     │                     │                    │  リソース作成      │
     │                     │                    ├──────────────────>│
     │                     │                    │                   │
     │                     │                    │  作成完了         │
     │                     │                    │<──────────────────┤
     │                     │ verify-api-health  │                   │
     │                     ├───────────────────>│                   │
     │                     │                    │  ヘルスチェック    │
     │                     │                    ├──────────────────>│
     │                     │                    │                   │
     │                     │                    │  200 OK           │
     │                     │                    │<──────────────────┤
     │ 環境準備完了        │<───────────────────┤                   │
     │<────────────────────┤                    │                   │
     │                     │                    │                   │
```

## 2. コマンド使用ガイド

### 2.1 開発環境管理コマンド

| コマンド | 説明 | 使用例 |
|---------|------|-------|
| `make start-api-dev` | API開発環境を完全にデプロイ | `make start-api-dev TF_ENV=development` |
| `make pause-api-dev` | API環境を一時停止（コア基盤は維持） | `make pause-api-dev TF_ENV=development` |
| `make resume-api-dev` | 一時停止したAPI環境を再開 | `make resume-api-dev TF_ENV=development` |
| `make stop-api-dev` | API環境を完全に削除 | `make stop-api-dev TF_ENV=development` |
| `make test-api-dev` | 一時的なテスト環境をデプロイし検証後に削除 | `make test-api-dev TF_ENV=development` |

### 2.2 検証コマンド

| コマンド | 説明 | 使用例 |
|---------|------|-------|
| `make verify-api-health` | APIのヘルスチェックを検証 | `make verify-api-health TF_ENV=development` |
| `make check-resources` | AWS環境のリソース状態を確認 | `make check-resources TF_ENV=development` |

## 3. よくあるエラーと解決策

### 3.1 ヘルスチェック失敗（404 Not Found）

**症状**:
- ALBのヘルスチェックが失敗し、ECSタスクが安定しない
- curlでヘルスチェックエンドポイントにアクセスすると404が返る

**解決策**:
1. APIのmain.goでヘルスチェックエンドポイントの設定を確認
   ```go
   // 正しい設定:
   router.HandleFunc("/health", healthHandler.Check).Methods(http.MethodGet)
   ```
2. コードを修正したらビルドとECRへのプッシュが必要
   ```bash
   make build SERVICE_TYPE=api
   make prepare-ecr-image SERVICE_TYPE=api TF_ENV=development
   ```
3. ECSサービスを更新
   ```bash
   cd deployments/terraform/environments/development
   terraform apply -target=module.service_api
   ```

### 3.2 503 Service Temporarily Unavailable

**症状**:
- APIエンドポイントにアクセスすると503エラーが返る
- ECSタスクはRunning状態だが、ヘルスチェックが通っていない

**解決策**:
1. ECSサービスのログを確認
   ```bash
   # タスクIDの取得
   TASK_ID=$(aws ecs list-tasks --cluster development-shared-cluster --service-name development-api --query 'taskArns[0]' --output text | awk -F'/' '{print $3}')
   # ログストリーム名の取得
   LATEST_LOG_STREAM=$(aws logs describe-log-streams --log-group-name /ecs/development-api --order-by LastEventTime --descending --limit 1 --query "logStreams[0].logStreamName" --output text)
   # ログの確認
   aws logs get-log-events --log-group-name /ecs/development-api --log-stream-name $LATEST_LOG_STREAM --limit 20
   ```
2. データベース接続設定を確認
3. セキュリティグループの設定を確認

## 4. 高度なシナリオ

### 4.1 環境のリセットと再構築

AWS環境とTerraform状態の不整合が発生した場合のリセットと再構築手順:

```bash
# 1. 状態のバックアップ
make terraform-backup TF_ENV=development

# 2. 状態のリセット
make terraform-reset TF_ENV=development

# 3. 再デプロイ
make start-api-dev TF_ENV=development
```

### 4.2 モジュール間の依存関係

REST API環境のモジュール間には以下の依存関係があります：

```
1. module.networking (VPC、サブネット)
2. module.database (RDSインスタンス)
3. module.shared_ecs_cluster, module.secrets (ECSクラスター、シークレット)
4. module.target_group_api (ターゲットグループ)
5. module.loadbalancer_api (ALB)
6. module.service_api (ECSサービス)
```

この順序でデプロイすることで、依存関係のエラーを防ぎます。

## 5. 実装詳細

### 5.1 Makefileコマンド実装

REST API環境管理コマンドは `makefiles/terraform-workflow.mk` に実装され、以下の機能を提供します：

```makefile
# REST API専用の開発ライフサイクル管理
.PHONY: start-api-dev pause-api-dev resume-api-dev stop-api-dev test-api-dev

# API開発環境のデプロイ（コア環境＋APIサービス）
start-api-dev:
  @echo -e "${BLUE}API開発環境を準備しています...${NC}"
  @cd deployments/terraform/environments/$(TF_ENV) && \
  terraform apply -auto-approve -target=module.networking -target=module.database -target=module.shared_ecs_cluster -target=module.secrets && \
  terraform apply -auto-approve -target=module.target_group_api -target=module.loadbalancer_api -target=module.service_api
  @echo -e "${GREEN}API開発環境が準備されました${NC}"
  @make verify-api-health TF_ENV=$(TF_ENV)

# API開発環境の一時停止
pause-api-dev:
  @echo -e "${BLUE}API開発環境を一時停止しています...${NC}"
  @cd deployments/terraform/environments/$(TF_ENV) && \
  terraform destroy -auto-approve \
  -target=module.service_api -target=module.loadbalancer_api -target=module.target_group_api
  @echo -e "${GREEN}API開発環境は一時停止されました。コア基盤は維持されています。${NC}"

# API開発環境の再開
resume-api-dev:
  @echo -e "${BLUE}API開発環境を再開しています...${NC}"
  @cd deployments/terraform/environments/$(TF_ENV) && \
  terraform apply -auto-approve -target=module.target_group_api -target=module.loadbalancer_api -target=module.service_api
  @echo -e "${GREEN}API開発環境が再開されました${NC}"
  @make verify-api-health TF_ENV=$(TF_ENV)

# API開発環境の完全停止
stop-api-dev:
  @echo -e "${BLUE}API開発環境を完全に停止しています...${NC}"
  @cd deployments/terraform/environments/$(TF_ENV) && \
  terraform destroy -auto-approve
  @echo -e "${GREEN}API開発環境は完全に停止されました。すべてのリソースが削除されています。${NC}"

# API開発環境のクイックテスト
test-api-dev:
  @echo -e "${BLUE}API一時テスト環境をデプロイしています...${NC}"
  @make start-api-dev TF_ENV=$(TF_ENV)
  @echo -e "${YELLOW}テスト環境が準備されました。検証中...${NC}"
  @make verify-api-health TF_ENV=$(TF_ENV)
  @echo -e "${BLUE}テスト完了。環境をクリーンアップしています...${NC}"
  @make stop-api-dev TF_ENV=$(TF_ENV)
```

### 5.2 検証コマンドの実装

APIヘルスチェック検証コマンドは `makefiles/verification.mk` に実装され、対応するスクリプトは `scripts/verification/verify-api-health.sh` に配置されます。
