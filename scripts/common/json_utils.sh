# scripts/common/json_utils.sh
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