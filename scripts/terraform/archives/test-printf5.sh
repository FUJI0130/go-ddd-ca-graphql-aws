#!/bin/bash
# advanced-verification-test.sh - より実践的な環境を想定した検証テスト

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ヘッダー表示
header() {
  echo ""
  echo "====================================="
  echo "🔍 $1"
  echo "====================================="
  echo ""
}

# 区切り線
separator() {
  echo "-------------------------------------"
}

# 通常のtrim関数
trim() {
  local var="$*"
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  echo -n "$var"
}

# 強化版trim関数
enhanced_trim() {
  local var="$*"
  # まず改行とキャリッジリターンを削除
  var=$(echo -n "$var" | tr -d '\n' | tr -d '\r')
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  echo -n "$var"
}

# 超強化版trim関数
super_trim() {
  local var="$*"
  # すべての空白文字と制御文字を処理
  var=$(echo -n "$var" | tr -d '\n\r\t\v\f')
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  # 念のためプリントfでフォーマット
  printf "%s" "$var"
}

# 変数の詳細情報表示
debug_var() {
  local name="$1"
  local value="$2"
  echo "📊 $name:"
  echo "  内容 (cat -A): $(cat -A <<< "$value")"
  echo "  長さ: ${#value}"
  echo "  16進ダンプ: $(hexdump -C <<< "$value" | head -1)"
}

# AWS CLIの出力をシミュレートする関数
simulate_aws_output() {
  local count=$1
  # AWS CLIの実際の出力パターンをシミュレート
  # 末尾に改行を追加（AWS CLIの典型的な出力）
  echo "$count"
}

# Terraformの出力をシミュレートする関数
simulate_tf_output() {
  local count=$1
  # Terraformの出力は必ずしも改行で終わらない
  echo -n "$count"
}

header "テスト1: AWS CLI出力シミュレーション - 実際の出力パターン"

# AWS CLIをシミュレート - VPC数（0）
VPC_COUNT=$(simulate_aws_output "0")
TF_VPC_COUNT=$(simulate_tf_output "0")

debug_var "AWS CLI出力 (VPC_COUNT)" "$VPC_COUNT"
debug_var "Terraform出力 (TF_VPC_COUNT)" "$TF_VPC_COUNT"

# 各種関数でクリーニング
VPC_COUNT_TRIM=$(trim "$VPC_COUNT")
TF_VPC_COUNT_TRIM=$(trim "$TF_VPC_COUNT")

VPC_COUNT_ENHANCED=$(enhanced_trim "$VPC_COUNT")
TF_VPC_COUNT_ENHANCED=$(enhanced_trim "$TF_VPC_COUNT")

VPC_COUNT_SUPER=$(super_trim "$VPC_COUNT")
TF_VPC_COUNT_SUPER=$(super_trim "$TF_VPC_COUNT")

debug_var "AWS CLI trim後" "$VPC_COUNT_TRIM"
debug_var "Terraform trim後" "$TF_VPC_COUNT_TRIM"
debug_var "AWS CLI enhanced_trim後" "$VPC_COUNT_ENHANCED"
debug_var "Terraform enhanced_trim後" "$TF_VPC_COUNT_ENHANCED"
debug_var "AWS CLI super_trim後" "$VPC_COUNT_SUPER"
debug_var "Terraform super_trim後" "$TF_VPC_COUNT_SUPER"

header "テスト2: 実際のスクリプトフローシミュレーション"

echo "🔹 通常の変数比較（trim後）"
separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator

# 方法1: 修正前 - 直接表示
if [ "$VPC_COUNT_TRIM" = "$TF_VPC_COUNT_TRIM" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo "方法1: 直接echo (trim後):"
echo -e "VPC\t\t$VPC_COUNT_TRIM\t$TF_VPC_COUNT_TRIM\t${STATUS_COLOR}$STATUS${NC}"

# 方法2: 修正案 - 一時変数
if [ "$VPC_COUNT_TRIM" = "$TF_VPC_COUNT_TRIM" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VPC_COUNT_TRIM\t$TF_VPC_COUNT_TRIM\t${STATUS_COLOR}$STATUS${NC}"
echo "方法2: 一時変数 (trim後):"
echo -e "$DISPLAY_LINE"

# 方法3: enhanced_trim
if [ "$VPC_COUNT_ENHANCED" = "$TF_VPC_COUNT_ENHANCED" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VPC_COUNT_ENHANCED\t$TF_VPC_COUNT_ENHANCED\t${STATUS_COLOR}$STATUS${NC}"
echo "方法3: enhanced_trim + 一時変数:"
echo -e "$DISPLAY_LINE"

# 方法4: super_trim
if [ "$VPC_COUNT_SUPER" = "$TF_VPC_COUNT_SUPER" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VPC_COUNT_SUPER\t$TF_VPC_COUNT_SUPER\t${STATUS_COLOR}$STATUS${NC}"
echo "方法4: super_trim + 一時変数:"
echo -e "$DISPLAY_LINE"

header "テスト3: 極端なケース - 複数行の出力"

# 複数行データシミュレーション
MULTI_AWS=$(echo -e "0\nNo resources found")
MULTI_TF="0"

debug_var "複数行 AWS CLI出力" "$MULTI_AWS"
debug_var "単一行 Terraform出力" "$MULTI_TF"

# 各種クリーニング
MULTI_AWS_TRIM=$(trim "$MULTI_AWS")
MULTI_TF_TRIM=$(trim "$MULTI_TF")

MULTI_AWS_ENHANCED=$(enhanced_trim "$MULTI_AWS")
MULTI_TF_ENHANCED=$(enhanced_trim "$MULTI_TF")

MULTI_AWS_SUPER=$(super_trim "$MULTI_AWS")
MULTI_TF_SUPER=$(super_trim "$MULTI_TF")

debug_var "複数行 AWS trim後" "$MULTI_AWS_TRIM"
debug_var "単一行 TF trim後" "$MULTI_TF_TRIM"
debug_var "複数行 AWS enhanced_trim後" "$MULTI_AWS_ENHANCED"
debug_var "単一行 TF enhanced_trim後" "$MULTI_TF_ENHANCED"
debug_var "複数行 AWS super_trim後" "$MULTI_AWS_SUPER"
debug_var "単一行 TF super_trim後" "$MULTI_TF_SUPER"

# 表示テスト
echo "🔹 複数行出力の比較"
separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator

# 方法1: 修正前 - 直接表示
if [ "$MULTI_AWS_TRIM" = "$MULTI_TF_TRIM" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo "方法1: 直接echo (trim後):"
echo -e "VPC\t\t$MULTI_AWS_TRIM\t$MULTI_TF_TRIM\t${STATUS_COLOR}$STATUS${NC}"

# 方法2: 修正案 - 一時変数
if [ "$MULTI_AWS_TRIM" = "$MULTI_TF_TRIM" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$MULTI_AWS_TRIM\t$MULTI_TF_TRIM\t${STATUS_COLOR}$STATUS${NC}"
echo "方法2: 一時変数 (trim後):"
echo -e "$DISPLAY_LINE"

# 方法3: enhanced_trim
if [ "$MULTI_AWS_ENHANCED" = "$MULTI_TF_ENHANCED" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$MULTI_AWS_ENHANCED\t$MULTI_TF_ENHANCED\t${STATUS_COLOR}$STATUS${NC}"
echo "方法3: enhanced_trim + 一時変数:"
echo -e "$DISPLAY_LINE"

# 方法4: super_trim
if [ "$MULTI_AWS_SUPER" = "$MULTI_TF_SUPER" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$MULTI_AWS_SUPER\t$MULTI_TF_SUPER\t${STATUS_COLOR}$STATUS${NC}"
echo "方法4: super_trim + 一時変数:"
echo -e "$DISPLAY_LINE"

header "テスト4: 完全な解決策テスト - フル処理アプローチ"

# ケース1: シンプルな出力 (0 vs 0)
AWS_SIMPLE=$(simulate_aws_output "0")
TF_SIMPLE=$(simulate_tf_output "0")

# ケース2: 複雑な出力 (複数行 vs 単純)
AWS_COMPLEX=$(echo -e "0\nNo resources found")
TF_COMPLEX="0"

# ケース3: 極端な出力 (空文字列 vs 0)
AWS_EXTREME=""
TF_EXTREME="0"

# 表示テスト関数
test_display() {
  local test_name="$1"
  local aws_var="$2"
  local tf_var="$3"
  
  echo "🔹 $test_name"
  
  # 超強化版トリム処理
  local aws_clean=$(super_trim "$aws_var")
  local tf_clean=$(super_trim "$tf_var")
  
  # 比較
  if [ "$aws_clean" = "$tf_clean" ]; then
    STATUS="一致"
    STATUS_COLOR=$GREEN
  else
    STATUS="不一致"
    STATUS_COLOR=$RED
  fi
  
  # 最終的な表示方法: 全て処理済みの変数を使用し、一時変数に格納してから表示
  DISPLAY_LINE="VPC\t\t$aws_clean\t$tf_clean\t${STATUS_COLOR}$STATUS${NC}"
  echo -e "$DISPLAY_LINE"
  
  # 念のため printf でも表示
  printf "Printf版: VPC\t\t%s\t%s\t${STATUS_COLOR}%s${NC}\n" "$aws_clean" "$tf_clean" "$STATUS"
}

# 各ケースをテスト
test_display "シンプルケース" "$AWS_SIMPLE" "$TF_SIMPLE"
test_display "複雑なケース" "$AWS_COMPLEX" "$TF_COMPLEX"
test_display "極端なケース" "$AWS_EXTREME" "$TF_EXTREME"

header "テスト5: aws-terraform-verify.sh 修正検証"

echo "🔹 最終的な修正アプローチの検証"

# オリジナルの処理に最も近いシミュレーション
simulate_aws_verify() {
  # 入力変数
  local aws_raw="$1"
  local tf_raw="$2"
  
  # トリム処理
  local aws_clean=$(super_trim "$aws_raw")
  local tf_clean=$(super_trim "$tf_raw")
  
  # デバッグ情報
  debug_var "AWS 処理前" "$aws_raw"
  debug_var "TF 処理前" "$tf_raw"
  debug_var "AWS 処理後" "$aws_clean"
  debug_var "TF 処理後" "$tf_clean"
  
  # 比較処理
  if [ "$aws_clean" = "$tf_clean" ]; then
    local status="一致"
    local status_color=$GREEN
  else
    local status="不一致"
    local status_color=$RED
  fi
  
  # 修正前の表示方法
  echo "修正前:"
  echo -e "VPC\t\t$aws_clean\t$tf_clean\t${status_color}$status${NC}"
  
  # 修正後の表示方法
  echo "修正後:"
  local display_line="VPC\t\t$aws_clean\t$tf_clean\t${status_color}$status${NC}"
  echo -e "$display_line"
}

# 3つのケースでスクリプト全体の流れをシミュレート
echo "ケース1: 通常出力"
simulate_aws_verify "$(simulate_aws_output "0")" "$(simulate_tf_output "0")"

echo ""
echo "ケース2: 複雑出力"
simulate_aws_verify "$(echo -e "0\nNo resources found")" "0"

echo ""
echo "ケース3: 極端なケース"
simulate_aws_verify "" "0"

header "結論: 最も堅牢な解決策"

echo "🔹 テスト結果の総括"
echo "1. 問題の本質: AWS CLI出力とTerraform状態出力の改行と表示の扱いの違い"
echo "2. 最適解: super_trim関数 + 一時変数を使用した表示方法"
echo "3. スクリプト修正方針: 全ての比較部分で一時変数を使用し、強化されたtrim関数を適用"

echo ""
echo "🔹 推奨される修正アプローチ"
echo "1. super_trim関数の追加:"
echo '  super_trim() {'
echo '    local var="$*"'
echo '    var=$(echo -n "$var" | tr -d "\n\r\t\v\f")'
echo '    var="${var#"${var%%[![:space:]]*}"}"'
echo '    var="${var%"${var##*[![:space:]]}"}"'
echo '    printf "%s" "$var"'
echo '  }'

echo ""
echo "2. 全ての比較部分の修正例:"
echo '  VPC_COUNT_CLEAN=$(super_trim "$VPC_COUNT")'
echo '  TF_VPC_COUNT_CLEAN=$(super_trim "$TF_VPC_COUNT")'
echo '  if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then'
echo '    STATUS="一致"'
echo '    STATUS_COLOR=$GREEN'
echo '  else'
echo '    STATUS="不一致"'
echo '    STATUS_COLOR=$RED'
echo '  fi'
echo '  DISPLAY_LINE="VPC\t\t$VPC_COUNT_CLEAN\t$TF_VPC_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"'
echo '  echo -e "$DISPLAY_LINE"'

echo ""
echo "3. 安全対策: 極端なケース（空文字列など）のハンドリングも考慮"