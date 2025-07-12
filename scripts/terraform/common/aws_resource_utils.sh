#!/bin/bash
# ===================================================================
# ファイル名: aws_resource_utils.sh
# 説明: AWS環境リソース管理のための共通ユーティリティ関数
# 
# 用途:
#  - AWS環境リソースの検出、作成、削除などの標準的な操作
#  - 一貫したエラーハンドリングと結果検証
#  - 複数のスクリプトから再利用可能な共通関数の提供
# 
# 使用方法:
#  source ./scripts/common/aws_resource_utils.sh
# ===================================================================

# デバッグ設定
AWS_RESOURCE_DEBUG=${AWS_RESOURCE_DEBUG:-false}
AWS_RESOURCE_DEBUG_FILE=${AWS_RESOURCE_DEBUG_FILE:-"/tmp/aws-debug.log"}
AWS_RESOURCE_DEBUG_VERBOSE=${AWS_RESOURCE_DEBUG_VERBOSE:-false}
AWS_RESOURCE_RETRY_MAX=${AWS_RESOURCE_RETRY_MAX:-3}
AWS_RESOURCE_RETRY_WAIT=${AWS_RESOURCE_RETRY_WAIT:-15}
AWS_RESOURCE_DELETE_TIMEOUT=${AWS_RESOURCE_DELETE_TIMEOUT:-180}

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# jqコマンドの存在確認
check_jq_command() {
  if command -v jq &> /dev/null; then
    return 0
  else
    return 1
  fi
}

# 代替JSONパース関数（jqがない場合）
# 完全に再実装して堅牢性を向上
parse_json_length() {
  local json_str="$1"
  
  # 空文字列、null、undefined、または空の配列/オブジェクトの場合は0を返す
  if [ -z "$json_str" ] || [ "$json_str" = "null" ] || [ "$json_str" = "undefined" ] || 
     [ "$json_str" = "[]" ] || [ "$json_str" = "{}" ]; then
    echo "0"
    return
  fi
  
  # JSONが配列かどうかを確認
  if [[ "$json_str" =~ ^\[.*\]$ ]]; then
    # 空の配列の場合
    if [[ "$json_str" = "[]" ]]; then
      echo "0"
      return
    fi
    
    # 要素の数を数える (カンマの数 + 1)
    # 文字列内のカンマは考慮しないシンプルな実装
    local comma_count=$(echo "$json_str" | grep -o ',' | wc -l)
    echo $((comma_count + 1))
  else
    # 配列でない場合（オブジェクトなど）
    # オブジェクトの場合は1を返す
    if [[ "$json_str" =~ ^\{.*\}$ ]]; then
      echo "1"
    else
      # 不明なフォーマットの場合は0を返す
      echo "0"
    fi
  fi
}

# JSON文字列から値を抽出（jqがない場合の代替）
extract_json_value() {
  local json_str="$1"
  local key="$2"
  local default_value="${3:-}"
  
  # 空文字列またはJSONでない場合はデフォルト値を返す
  if [ -z "$json_str" ] || [[ ! "$json_str" =~ [\{\[] ]]; then
    echo "$default_value"
    return
  fi
  
  # 簡易的なキー値抽出
  # "key":"value" または "key": "value" のパターンを検索
  local pattern="\"$key\"[[:space:]]*:[[:space:]]*\"?([^\",}]*)\"?"
  if [[ "$json_str" =~ $pattern ]]; then
    echo "${BASH_REMATCH[1]}"
  else
    # 見つからない場合はデフォルト値を返す
    echo "$default_value"
  fi
}

# デバッグログ関数
aws_resource_debug() {
  local message="$1"
  local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
  
  if [ "$AWS_RESOURCE_DEBUG" = "true" ]; then
    echo -e "${BLUE}[DEBUG] $timestamp - $message${NC}" >&2
    echo "[DEBUG] $timestamp - $message" >> "$AWS_RESOURCE_DEBUG_FILE"
    
    # 詳細モードの場合は呼び出し元の情報も記録
    if [ "$AWS_RESOURCE_DEBUG_VERBOSE" = "true" ]; then
      local caller_info=$(caller)
      echo -e "${BLUE}[DEBUG] Called from: $caller_info${NC}" >&2
      echo "[DEBUG] Called from: $caller_info" >> "$AWS_RESOURCE_DEBUG_FILE"
    fi
  fi
}

# エラー出力関数
aws_resource_error() {
  local message="$1"
  local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
  echo -e "${RED}[ERROR] $timestamp - $message${NC}" >&2
  echo "[ERROR] $timestamp - $message" >> "$AWS_RESOURCE_DEBUG_FILE"
}

# 警告出力関数
aws_resource_warning() {
  local message="$1"
  local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
  echo -e "${YELLOW}[WARNING] $timestamp - $message${NC}" >&2
  echo "[WARNING] $timestamp - $message" >> "$AWS_RESOURCE_DEBUG_FILE"
}

# 成功出力関数
aws_resource_success() {
  local message="$1"
  local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
  echo -e "${GREEN}[SUCCESS] $timestamp - $message${NC}" >&2
  echo "[SUCCESS] $timestamp - $message" >> "$AWS_RESOURCE_DEBUG_FILE"
}

# 情報出力関数
aws_resource_info() {
  local message="$1"
  local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
  echo -e "${BLUE}[INFO] $timestamp - $message${NC}" >&2
  echo "[INFO] $timestamp - $message" >> "$AWS_RESOURCE_DEBUG_FILE"
}

# AWS CLIコマンド実行のラッパー関数
aws_cli_exec() {
  local command="$1"
  shift
  local args=("$@")
  
  aws_resource_debug "Executing: aws $command ${args[*]}"
  
  local output
  if output=$(aws "$command" "${args[@]}" 2>&1); then
    aws_resource_debug "Command succeeded"
    if [ "$AWS_RESOURCE_DEBUG_VERBOSE" = "true" ]; then
      aws_resource_debug "Output: $output"
    fi
    echo "$output"
    return 0
  else
    local exit_code=$?
    aws_resource_debug "Command failed with exit code $exit_code"
    aws_resource_debug "Error output: $output"
    echo "$output" >&2
    return $exit_code
  fi
}

# 環境変数のデバッグダンプ
aws_resource_dump_env() {
  aws_resource_debug "Environment variables dump:"
  aws_resource_debug "  AWS_RESOURCE_DEBUG=$AWS_RESOURCE_DEBUG"
  aws_resource_debug "  AWS_RESOURCE_DEBUG_FILE=$AWS_RESOURCE_DEBUG_FILE"
  aws_resource_debug "  AWS_RESOURCE_DEBUG_VERBOSE=$AWS_RESOURCE_DEBUG_VERBOSE"
  aws_resource_debug "  AWS_RESOURCE_RETRY_MAX=$AWS_RESOURCE_RETRY_MAX"
  aws_resource_debug "  AWS_RESOURCE_RETRY_WAIT=$AWS_RESOURCE_RETRY_WAIT"
  aws_resource_debug "  AWS_RESOURCE_DELETE_TIMEOUT=$AWS_RESOURCE_DELETE_TIMEOUT"
  aws_resource_debug "  PATH=$PATH"
  aws_resource_debug "  PWD=$PWD"
  aws_resource_debug "  ENV=$ENV"
  aws_resource_debug "  TF_ENV=$TF_ENV"
  
  # AWS認証情報の有無（値は表示しない）
  if [ -n "$AWS_ACCESS_KEY_ID" ]; then
    aws_resource_debug "  AWS_ACCESS_KEY_ID=***SET***"
  else
    aws_resource_debug "  AWS_ACCESS_KEY_ID=***NOT_SET***"
  fi
  
  if [ -n "$AWS_SECRET_ACCESS_KEY" ]; then
    aws_resource_debug "  AWS_SECRET_ACCESS_KEY=***SET***"
  else
    aws_resource_debug "  AWS_SECRET_ACCESS_KEY=***NOT_SET***"
  fi
  
  if [ -n "$AWS_SESSION_TOKEN" ]; then
    aws_resource_debug "  AWS_SESSION_TOKEN=***SET***"
  else
    aws_resource_debug "  AWS_SESSION_TOKEN=***NOT_SET***"
  fi
  
  if [ -n "$AWS_PROFILE" ]; then
    aws_resource_debug "  AWS_PROFILE=$AWS_PROFILE"
  else
    aws_resource_debug "  AWS_PROFILE=***NOT_SET***"
  fi
  
  if [ -n "$AWS_REGION" ]; then
    aws_resource_debug "  AWS_REGION=$AWS_REGION"
  else
    aws_resource_debug "  AWS_REGION=***NOT_SET***"
  fi
}

# ========== ECSクラスター関連関数 ==========

# ECSクラスターの存在確認（デバッグ強化版 - jq依存なし）
# 引数:
#   $1: クラスター名
# 戻り値:
#   0: クラスターが存在する
#   1: クラスターが存在しない
ecs_cluster_exists() {
  local cluster_name="$1"
  
  aws_resource_debug "Checking if ECS cluster exists: '$cluster_name'"
  
  # 環境変数の確認
  aws_resource_debug "Current ENV='$ENV', calling from script: $(basename "$0")"
  
  # クラスターの存在確認 - text出力形式を使用
  local clusters_text
  clusters_text=$(aws ecs describe-clusters --clusters "$cluster_name" --query "clusters[?status=='ACTIVE'].clusterName" --output text 2>/dev/null || echo "")
  
  if [ -z "$clusters_text" ] || [ "$clusters_text" = "None" ]; then
    aws_resource_debug "Cluster '$cluster_name' does not exist (text output check)"
    return 1
  fi
  
  # クラスター名の完全一致を確認
  if echo "$clusters_text" | grep -q "^$cluster_name$"; then
    aws_resource_debug "Cluster '$cluster_name' exists (direct name match)"
    return 0
  fi
  
  # 最終確認としてJSONアプローチを試す
  aws_resource_debug "Trying JSON approach as fallback"
  
  local cluster_json
  cluster_json=$(aws ecs describe-clusters --clusters "$cluster_name" --query "clusters" --output json 2>/dev/null || echo "[]")
  
  local cluster_count=0
  # jqが利用可能な場合はjqを使用
  if check_jq_command; then
    cluster_count=$(echo "$cluster_json" | jq 'length')
    aws_resource_debug "Cluster count from jq: $cluster_count"
  else
    # jqが利用できない場合は独自パース関数を使用
    cluster_count=$(parse_json_length "$cluster_json")
    aws_resource_debug "Cluster count from simple parser: $cluster_count"
  fi
  
  # 数値変換の確認
  if [[ ! "$cluster_count" =~ ^[0-9]+$ ]]; then
    aws_resource_warning "Failed to parse cluster count: '$cluster_count', assuming 0"
    cluster_count=0
  fi
  
  if [ "$cluster_count" -gt 0 ]; then
    aws_resource_debug "Cluster '$cluster_name' exists (JSON count check)"
    return 0
  else
    aws_resource_debug "Cluster '$cluster_name' does not exist (JSON count check)"
    return 1
  fi
}

# ECSクラスターのステータス取得
# 引数:
#   $1: クラスター名
# 出力:
#   クラスターのステータス（文字列）または "NOT_FOUND"
ecs_cluster_status() {
  local cluster_name="$1"
  
  if ecs_cluster_exists "$cluster_name"; then
    local status
    status=$(aws ecs describe-clusters --clusters "$cluster_name" --query "clusters[0].status" --output text 2>/dev/null || echo "UNKNOWN")
    aws_resource_debug "ECSクラスター '$cluster_name' のステータス: $status"
    echo "$status"
  else
    aws_resource_debug "ECSクラスター '$cluster_name' は存在しないため、ステータスは 'NOT_FOUND'"
    echo "NOT_FOUND"
  fi
}

# ECSクラスター内のサービス一覧取得
# 引数:
#   $1: クラスター名
# 出力:
#   サービスARNのリスト（スペース区切り）
ecs_list_services() {
  local cluster_name="$1"
  
  if ecs_cluster_exists "$cluster_name"; then
    local services
    services=$(aws ecs list-services --cluster "$cluster_name" --query "serviceArns[*]" --output text 2>/dev/null || echo "")
    
    if [ -z "$services" ] || [ "$services" = "None" ]; then
      aws_resource_debug "ECSクラスター '$cluster_name' 内にサービスはありません"
      echo ""
    else
      aws_resource_debug "ECSクラスター '$cluster_name' 内のサービス: $services"
      echo "$services"
    fi
  else
    aws_resource_debug "ECSクラスター '$cluster_name' は存在しないため、サービスはありません"
    echo ""
  fi
}

# ECSサービスの削除
# 引数:
#   $1: クラスター名
#   $2: サービス名/ARN
# 戻り値:
#   0: 削除成功または既に存在しない
#   1: 削除失敗
ecs_delete_service() {
  local cluster_name="$1"
  local service_name="$2"
  
  # サービス名からARNを抽出（必要な場合）
  local service_short_name=$(basename "$service_name")
  
  aws_resource_info "ECSサービス '$service_short_name' を削除しています..."
  
  # まずdesired countを0に設定
  aws_cli_exec ecs update-service --cluster "$cluster_name" --service "$service_short_name" --desired-count 0 || true
  
  # サービスを削除
  if aws_cli_exec ecs delete-service --cluster "$cluster_name" --service "$service_short_name" --force; then
    aws_resource_success "ECSサービス '$service_short_name' の削除に成功しました"
    return 0
  else
    aws_resource_error "ECSサービス '$service_short_name' の削除に失敗しました"
    return 1
  fi
}

# ECSクラスターの削除（強化版 - ポーリング戦略実装）
# 引数:
#   $1: クラスター名
#   $2: (オプション) 強制削除フラグ (true/false, デフォルトはfalse)
#   $3: (オプション) 最大リトライ回数 (デフォルトは環境変数または3)
#   $4: (オプション) リトライ間の待機時間(秒) (デフォルトは環境変数または15)
# 戻り値:
#   0: 削除成功または既に存在しない
#   1: 削除失敗
ecs_delete_cluster() {
  local cluster_name="$1"
  local force=${2:-false}
  local max_retries=${3:-$AWS_RESOURCE_RETRY_MAX}
  local wait_time=${4:-$AWS_RESOURCE_RETRY_WAIT}
  local timeout=${AWS_RESOURCE_DELETE_TIMEOUT:-180}
  local start_time=$(date +%s)
  
  aws_resource_debug "Attempting to delete ECS cluster: '$cluster_name' (force=$force, max_retries=$max_retries, wait_time=$wait_time)"
  
  # 環境変数の確認
  aws_resource_debug "Current ENV='$ENV', TF_ENV='$TF_ENV', cluster_name='$cluster_name'"
  
  # クラスター存在確認
  local exists_before_delete
  if ecs_cluster_exists "$cluster_name"; then
    exists_before_delete=true
    aws_resource_info "Cluster '$cluster_name' exists and will be deleted"
  else
    exists_before_delete=false
    aws_resource_info "Cluster '$cluster_name' does not exist"
    if [ "$force" != "true" ]; then
      aws_resource_success "No need to delete non-existent cluster"
      return 0
    fi
  fi
  
  # 強制フラグがtrueか、クラスターが存在する場合に削除を実行
  if [ "$exists_before_delete" = "true" ] || [ "$force" = "true" ]; then
    # サービスの確認と削除
    if [ "$exists_before_delete" = "true" ]; then
      aws_resource_debug "Checking for services in cluster"
      
      # サービス一覧の取得
      local services
      services=$(aws ecs list-services --cluster "$cluster_name" --output text 2>/dev/null || echo "")
      
      # サービスが存在する場合は削除
      if [ ! -z "$services" ] && [ "$services" != "None" ]; then
        aws_resource_debug "Services to delete: $services"
        
        # スペース区切りで各サービスを処理
        for service_arn in $services; do
          local service_name
          service_name=$(basename "$service_arn")
          aws_resource_info "Deleting service: $service_name"
          
          # サービスのdesired countを0に設定
          aws ecs update-service --cluster "$cluster_name" --service "$service_name" --desired-count 0 >/dev/null 2>&1 || true
          
          # サービスを削除
          aws ecs delete-service --cluster "$cluster_name" --service "$service_name" --force >/dev/null 2>&1 || true
          
          aws_resource_debug "Service deletion initiated: $service_name"
        done
        
        # サービス削除の完了を待機
        aws_resource_info "Waiting for services to be deleted ($wait_time seconds)..."
        sleep $wait_time
      else
        aws_resource_debug "No services found in cluster"
      fi
    fi
    
    # クラスターの削除
    aws_resource_info "Deleting cluster: $cluster_name"
    
    # 削除コマンドの実行
    aws ecs delete-cluster --cluster "$cluster_name" >/dev/null 2>&1 || true
    
    # 削除確認（リトライあり）
    for attempt in $(seq 1 $max_retries); do
      aws_resource_debug "Verification attempt $attempt/$max_retries"
      sleep $wait_time
      
      # タイムアウトチェック
      local current_time=$(date +%s)
      local elapsed_time=$((current_time - start_time))
      if [ $elapsed_time -gt $timeout ]; then
        aws_resource_warning "Timeout reached after ${elapsed_time}s (limit: ${timeout}s)"
        break
      fi
      
      if ! ecs_cluster_exists "$cluster_name"; then
        aws_resource_success "Cluster '$cluster_name' successfully deleted"
        return 0
      else
        aws_resource_warning "Cluster still exists after deletion attempt $attempt"
        
        # 2回目以降のリトライでは強制削除を試みる
        if [ "$attempt" -ge 2 ]; then
          aws_resource_debug "Retry $attempt: Trying with more aggressive approach"
          # 強制的に削除を再試行
          aws ecs delete-cluster --cluster "$cluster_name" --force >/dev/null 2>&1 || true
          # 待機時間を少し長くする
          sleep $((wait_time + 5))
        fi
      fi
    done
    
    # 最終リトライ - 待機時間を長めに取り、明示的にクラスター名を指定
    aws_resource_warning "Final attempt: trying direct deletion"
    aws ecs delete-cluster --cluster "$cluster_name" >/dev/null 2>&1 || true
    sleep $((wait_time * 2))
    
    # 最終確認
    if ! ecs_cluster_exists "$cluster_name"; then
      aws_resource_success "Cluster deleted after final attempt"
      return 0
    else
      aws_resource_error "Failed to delete cluster even after all attempts"
      aws_resource_error "Manual intervention may be required"
      aws_resource_error "Try: aws ecs delete-cluster --cluster \"$cluster_name\""
      return 1
    fi
  else
    aws_resource_success "Cluster '$cluster_name' already does not exist"
    return 0
  fi
}

# ライブラリ読み込み完了メッセージ
aws_resource_debug "AWS Resource Utilities library loaded successfully"
aws_resource_dump_env