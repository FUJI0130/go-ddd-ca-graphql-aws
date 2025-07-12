# AWS検証スクリプト改善ガイド：jq依存性の解決と堅牢性向上

## 1. 背景と課題

AWSリソース検証スクリプトでは、JSON形式のレスポンスを処理するために`jq`コマンドが広く使用されています。しかし、`jq`はすべての環境に標準でインストールされているわけではなく、これが環境依存性の問題を引き起こします。本ガイドでは、この依存性を解消しつつ、AWS CLIの機能を最大限に活用する方法について説明します。

### 1.1 主な課題

- `jq`コマンドへの依存性（コマンドが見つからないエラー）
- 異なる環境での一貫した動作の確保
- JSON処理の効率性と正確性の維持
- エラーメッセージの明確化と問題解決の容易さ

## 2. 共通ユーティリティライブラリの作成

### 2.1 ファイル作成

以下のパスに共通ユーティリティライブラリを作成します：

```bash
mkdir -p scripts/common
touch scripts/common/json_utils.sh
```

### 2.2 ユーティリティ関数の実装

`json_utils.sh`に以下の内容を実装します：

```bash
#!/bin/bash

# JSON処理用ユーティリティ関数

# jqの有無を確認
has_jq() {
    command -v jq &> /dev/null
}

# AWS CLI Output JSONから値を取得（AWS CLI --query が使用可能な場合に使用）
get_aws_value() {
    local service="$1"     # 例: 'ecs'
    local command="$2"     # 例: 'describe-services'
    local query="$3"       # 例: 'services[0].status'
    local output="${4:-text}" # デフォルトは 'text'
    shift 4
    
    # 残りの引数をそのまま渡す
    aws "$service" "$command" --query "$query" --output "$output" "$@"
}

# JSON文字列から値を取得
get_json_value() {
    local json="$1"
    local path="$2"
    
    if has_jq; then
        # jqが利用可能な場合
        echo "$json" | jq -r "$path"
    else
        # jqが利用できない場合の代替実装
        case "$path" in
            '.services[0].status')
                echo "$json" | grep -o '"status":"[^"]*"' | head -1 | sed 's/"status":"//;s/"//g'
                ;;
            '.services[0].runningCount')
                echo "$json" | grep -o '"runningCount":[0-9]*' | head -1 | sed 's/"runningCount"://g'
                ;;
            '.services[0].desiredCount')
                echo "$json" | grep -o '"desiredCount":[0-9]*' | head -1 | sed 's/"desiredCount"://g'
                ;;
            # 必要に応じてパターンを追加
            *)
                echo "Unsupported JSON path: $path" >&2
                return 1
                ;;
        esac
    fi
}

# JSON配列のフィルタリング (jq 'select(...)' 相当)
filter_json_array() {
    local json="$1"
    local array_path="$2"  # 例: '.TargetHealthDescriptions[]'
    local filter="$3"      # 例: '.TargetHealth.State=="healthy"'
    
    if has_jq; then
        echo "$json" | jq -r "$array_path | select($filter)"
    else
        # フィルタリングの複雑さにより、基本的なケースのみサポート
        echo "Complex JSON filtering requires jq. Please install jq for full functionality." >&2
        return 1
    fi
}

# サポートされていない場合のフォールバック処理
fallback_with_warning() {
    local message="$1"
    local fallback_value="$2"
    
    echo "WARNING: $message" >&2
    echo "$fallback_value"
}
```

## 3. 検証スクリプトの改善

### 3.1 スクリプトの基本構造

AWS検証スクリプトの基本的な改善パターンは以下の通りです：

1. 共通ユーティリティライブラリをインポート
2. jqコマンドの直接呼び出しを共通関数に置き換え
3. AWS CLI `--query`オプションを最大限に活用
4. 存在確認から始めるよう処理順序を最適化

### 3.2 基本的な更新パターン

```bash
# 変更前:
SERVICE_DETAILS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME})
SERVICE_STATUS=$(echo ${SERVICE_DETAILS} | jq -r '.services[0].status')

# 変更後:
# サービスの存在確認
SERVICE_COUNT=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'length(services)' --output text)

if [ "${SERVICE_COUNT}" = "0" ]; then
  echo -e "${RED}エラー: サービス ${SERVICE_NAME} が存在しません${NC}"
  exit 1
fi

# サービスが存在する場合のみステータスを取得
SERVICE_STATUS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].status' --output text)
```

### 3.3 条件分岐パターン

複雑なJSON処理が必要な場合は、jqの有無によって処理を分岐します：

```bash
if has_jq; then
  # jqを使用した処理
  # ...
else
  # 代替処理（AWS CLI --queryとシェルコマンド）
  # ...
fi
```

## 4. 実装のベストプラクティス

### 4.1 AWS CLIの--queryオプション活用のコツ

1. **サービスやリソースの存在確認**:
   ```bash
   # 配列の長さをチェック
   aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'length(services)' --output text
   ```

2. **特定の属性値の取得**:
   ```bash
   # 直接パスを指定して値を取得
   aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].status' --output text
   ```

3. **条件によるフィルタリング**:
   ```bash
   # 条件に一致する要素の抽出
   aws elbv2 describe-target-groups --load-balancer-arn ${ALB_ARN} --query 'TargetGroups[?contains(TargetGroupName, `api`)].TargetGroupArn' --output text
   ```

4. **複数の値の数のカウント**:
   ```bash
   # 条件に一致する要素数のカウント
   aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN} --query 'length(TargetHealthDescriptions[?TargetHealth.State==`healthy`])' --output text
   ```

### 4.2 エラーメッセージの明確化

1. **コンテキスト情報の提供**:
   ```bash
   echo -e "${RED}エラー: サービス ${SERVICE_NAME} が存在しません${NC}"
   ```

2. **問題と解決策の提示**:
   ```bash
   echo -e "${RED}エラー: ALB ${LOAD_BALANCER_NAME} が見つかりません。ロードバランサーが正しくデプロイされているか確認してください。${NC}"
   ```

3. **代替情報の表示**:
   ```bash
   echo "不健全なターゲットの詳細を取得するにはjqが必要です。"
   echo "代わりに基本情報を表示します:"
   ```

## 5. 実装例

### 5.1 サービスの存在確認と状態チェック

```bash
# ECSサービスの存在確認
echo "ECSサービスの存在を確認しています..."
SERVICE_COUNT=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'length(services)' --output text)

if [ "${SERVICE_COUNT}" = "0" ]; then
  echo -e "${RED}エラー: サービス ${SERVICE_NAME} が存在しません${NC}"
  exit 1
fi

# ECSサービスの状態確認
echo "ECSサービスのステータスを確認しています..."
SERVICE_STATUS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].status' --output text)

if [ "${SERVICE_STATUS}" != "ACTIVE" ]; then
  echo -e "${RED}エラー: サービスがアクティブではありません (Status: ${SERVICE_STATUS})${NC}"
  exit 1
fi
```

### 5.2 タスク数の確認

```bash
# 実行中のタスク数の確認
RUNNING_TASKS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].runningCount' --output text)
DESIRED_TASKS=$(aws ecs describe-services --cluster ${CLUSTER_NAME} --services ${SERVICE_NAME} --query 'services[0].desiredCount' --output text)

if [ ${RUNNING_TASKS} -lt ${DESIRED_TASKS} ]; then
  echo -e "${YELLOW}警告: 実行中のタスク数が期待値よりも少ないです (実行中: ${RUNNING_TASKS}, 期待値: ${DESIRED_TASKS})${NC}"
  # ...追加の診断...
else
  echo -e "${GREEN}✓ タスク数は期待通りです (実行中: ${RUNNING_TASKS}, 期待値: ${DESIRED_TASKS})${NC}"
fi
```

### 5.3 ヘルスチェック状態の監視

```bash
# jqの有無を検出し、適切な実装を選択
if has_jq; then
  # jqが利用可能な場合
  HEALTH_STATUS=$(aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN})
  HEALTHY_COUNT=$(echo ${HEALTH_STATUS} | jq -r '.TargetHealthDescriptions[] | select(.TargetHealth.State=="healthy") | .TargetHealth.State' | wc -l)
  UNHEALTHY_COUNT=$(echo ${HEALTH_STATUS} | jq -r '.TargetHealthDescriptions[] | select(.TargetHealth.State!="healthy") | .TargetHealth.State' | wc -l)
else
  # jqが利用できない場合 - AWS CLIのクエリを使用
  HEALTHY_TARGETS=$(aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN} \
                   --query 'length(TargetHealthDescriptions[?TargetHealth.State==`healthy`])' --output text)
  HEALTHY_COUNT=${HEALTHY_TARGETS:-0}
  
  UNHEALTHY_TARGETS=$(aws elbv2 describe-target-health --target-group-arn ${TARGET_GROUP_ARN} \
                    --query 'length(TargetHealthDescriptions[?TargetHealth.State!=`healthy`])' --output text)
  UNHEALTHY_COUNT=${UNHEALTHY_TARGETS:-0}
fi
```

## 6. 検証スクリプト適用ステップ

### 6.1 修正手順

1. 共通ユーティリティライブラリを作成・配置
2. 検証スクリプトが共通ライブラリをインポートするように修正
3. jqコマンドの直接呼び出しを共通関数や`--query`オプションに置き換え
4. エラーメッセージを明確化
5. 動作確認とテスト

### 6.2 修正後の確認ポイント

- スクリプトがjqなしでも動作するか
- エラーメッセージが明確で有用な情報を提供しているか
- サービスの状態に応じて適切な動作を行うか
- AWS CLIの`--query`オプションが適切に使用されているか

## 7. 注意点と推奨事項

### 7.1 環境依存性の対応

- 検証スクリプトは様々な環境（開発環境、CI/CD環境、本番環境）で実行される可能性がある
- すべての環境でjqがインストールされているとは限らない
- AWS CLIは標準で存在することを前提としている
- 環境変数や設定ファイルによる追加のカスタマイズも考慮する

### 7.2 エラーハンドリングのベストプラクティス

- 早期にリソースの存在確認を行う
- エラーメッセージには問題の内容だけでなく、考えられる原因と解決策も含める
- カラーコードを活用して視認性を向上させる
- デバッグ情報を適切に提供する

### 7.3 AWS CLIの使い方のヒント

- `--query`オプションはJMESPathクエリ言語を使用している
- 複雑なフィルタリングや配列操作も可能
- `--output`オプションで出力形式を指定（text, json, yaml, table）
- クエリ結果が存在しない場合は`None`が返される

### 7.4 セキュリティ考慮事項

- スクリプトでの資格情報の扱いに注意
- AWS CLIのプロファイル機能を活用
- シークレット情報をハードコーディングしない
- 必要最小限の権限で実行する

## 8. トラブルシューティング

### 8.1 よくある問題と解決策

1. **クエリ結果が `None` になる場合**
   - リソースが存在しないか、クエリパスが間違っている
   - AWS CLIのバージョンによって挙動が異なる場合もある
   - 実際のリソース名や構造を確認する

2. **パラメータ展開の問題**
   - 変数が空の場合のデフォルト値を設定: `${VARIABLE:-default}`
   - 配列操作時は引用符の使用に注意

3. **パス解決の問題**
   - スクリプトの実行場所によって相対パスが変わる
   - `$(dirname "$0")`を使用して現在のスクリプトディレクトリを取得

### 8.2 デバッグ技術

- 一時的に `set -x` を追加して実行コマンドを表示
- 重要な変数の値を確認するためのデバッグ出力を追加
- AWS CLIの `--debug` オプションを使用して詳細なAPI情報を確認

## 9. まとめ

本ガイドで紹介した手法を用いることで、jqなどの外部ツールへの依存性を減らしつつ、堅牢なAWS検証スクリプトを作成できます。AWS CLIの`--query`オプションを活用し、共通ユーティリティライブラリを整備することで、環境に依存しない一貫した動作と明確なエラーメッセージを提供できるようになります。

作成したスクリプトは、AWS環境のデプロイ検証だけでなく、日常的な運用やトラブルシューティングにも役立つ重要なツールとなるでしょう。