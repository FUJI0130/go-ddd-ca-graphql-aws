#!/bin/bash
# enhanced-test-script.sh - 様々なケースでの出力問題を検証するテストスクリプト

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ヘッダー表示
header() {
  echo ""
  echo "==== $1 ===="
  echo ""
}

# 区切り線
separator() {
  echo "--------------------------------"
}

# 通常のtrim関数
trim() {
  local var="$*"
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  echo -n "$var"
}

# 強化されたtrim関数（改行と特殊文字を徹底的に削除）
enhanced_trim() {
  local var="$*"
  # まず改行とキャリッジリターンを削除
  var=$(echo -n "$var" | tr -d '\n' | tr -d '\r')
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  echo -n "$var"
}

# テスト1: 通常の変数比較
header "テスト1: 通常の変数比較（基本ケース）"
VAR1="0"
VAR2="0"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "通常のecho: VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"

# プリント文を使う場合
printf "printfを使用: VPC\t\t%s\t%s\t${STATUS_COLOR}%s${NC}\n" "$VAR1" "$VAR2" "$STATUS"

# テスト2: 改行を含む変数での比較
header "テスト2: 改行を含む変数での比較"
VAR1="0"
VAR2="0
"  # VAR2に意図的に改行を追加
echo "VAR1='$VAR1', VAR2の中身（改行あり）:"
echo "$(cat -A <<< "$VAR2")"  # 制御文字を表示
echo "VAR1の長さ: ${#VAR1}, VAR2の長さ: ${#VAR2}"

# 直接表示
echo ""
echo "直接表示:"
separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致" 
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"

# 一時変数を使用
echo ""
echo "一時変数を使用:"
separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致" 
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

# printfを使用
echo ""
echo "printfを使用:"
separator
printf "リソース\t\tAWS\tTerraform\t状態\n"
separator
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致" 
  STATUS_COLOR=$RED
fi
printf "VPC\t\t%s\t%s\t${STATUS_COLOR}%s${NC}\n" "$VAR1" "$VAR2" "$STATUS"

# テスト3: trim関数のテスト
header "テスト3: trim関数のテスト"
VAR_ORIGINAL="0
"  # 改行を含む変数
VAR_TRIMMED=$(trim "$VAR_ORIGINAL")
VAR_ENHANCED=$(enhanced_trim "$VAR_ORIGINAL")

echo "オリジナル変数の中身（改行あり）:"
echo "$(cat -A <<< "$VAR_ORIGINAL")"
echo "オリジナル変数の長さ: ${#VAR_ORIGINAL}"

echo "通常のtrim後:"
echo "$(cat -A <<< "$VAR_TRIMMED")"
echo "通常のtrim後の長さ: ${#VAR_TRIMMED}"

echo "強化版trim後:"
echo "$(cat -A <<< "$VAR_ENHANCED")"
echo "強化版trim後の長さ: ${#VAR_ENHANCED}"

# テスト4: AWS CLI出力のシミュレーション
header "テスト4: AWS CLI出力のシミュレーション"

# AWS CLIの出力を模倣（改行を含む）
AWS_OUTPUT="0
"  # 意図的に改行を含む
TF_OUTPUT="0"

echo "AWS出力の生データ:"
echo "$(cat -A <<< "$AWS_OUTPUT")"

echo "Terraform出力の生データ:"
echo "$(cat -A <<< "$TF_OUTPUT")"

# 4.1: 通常のtrim
AWS_CLEAN=$(trim "$AWS_OUTPUT")
TF_CLEAN=$(trim "$TF_OUTPUT")

echo "通常のtrim:"
echo "AWS_CLEAN='$AWS_CLEAN', TF_CLEAN='$TF_CLEAN'"
echo "AWS_CLEANの長さ: ${#AWS_CLEAN}, TF_CLEANの長さ: ${#TF_CLEAN}"

separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator
if [ "$AWS_CLEAN" = "$TF_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$AWS_CLEAN\t$TF_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

# 4.2: 強化版trim
AWS_ENHANCED=$(enhanced_trim "$AWS_OUTPUT")
TF_ENHANCED=$(enhanced_trim "$TF_OUTPUT")

echo "強化版trim:"
echo "AWS_ENHANCED='$AWS_ENHANCED', TF_ENHANCED='$TF_ENHANCED'"
echo "AWS_ENHANCEDの長さ: ${#AWS_ENHANCED}, TF_ENHANCEDの長さ: ${#TF_ENHANCED}"

separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator
if [ "$AWS_ENHANCED" = "$TF_ENHANCED" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$AWS_ENHANCED\t$TF_ENHANCED\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

# テスト5: 複数の変数を含む表示
header "テスト5: 複数の変数を含む表示"

VAR1="0
"  # 改行あり
VAR2="0
"  # 改行あり
VAR3="0"   # 改行なし

echo "変数の中身:"
echo "VAR1: $(cat -A <<< "$VAR1")"
echo "VAR2: $(cat -A <<< "$VAR2")"
echo "VAR3: $(cat -A <<< "$VAR3")"

# 5.1: 直接表示
echo "直接表示:"
separator
echo -e "リソース\t\tAWS\tTerraform\t追加\t状態"
separator
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VAR1\t$VAR2\t$VAR3\t${STATUS_COLOR}$STATUS${NC}"

# 5.2: 一時変数を使用
echo "一時変数を使用:"
separator
echo -e "リソース\t\tAWS\tTerraform\t追加\t状態"
separator
VAR1_CLEAN=$(enhanced_trim "$VAR1")
VAR2_CLEAN=$(enhanced_trim "$VAR2")
VAR3_CLEAN=$(enhanced_trim "$VAR3")

if [ "$VAR1_CLEAN" = "$VAR2_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VAR1_CLEAN\t$VAR2_CLEAN\t$VAR3_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

# テスト6: 実際のケースの模倣
header "テスト6: 実際のケースの模倣"

# トリム関数
trim_for_comparison() {
  local var="$*"
  var=$(echo -n "$var" | tr -d '\n' | tr -d '\r')
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  echo -n "$var"
}

# 実際のAWS CLIの出力と同様の形式
VPC_COUNT=$(echo -e "0\n")  # 改行を含む出力
TF_VPC_COUNT=$(echo "0")    # 改行なし

# トリム処理
VPC_COUNT_CLEAN=$(trim_for_comparison "$VPC_COUNT")
TF_VPC_COUNT_CLEAN=$(trim_for_comparison "$TF_VPC_COUNT")

echo "変数の中身:"
echo "VPC_COUNT（改行あり）: $(cat -A <<< "$VPC_COUNT")"
echo "TF_VPC_COUNT: $(cat -A <<< "$TF_VPC_COUNT")"
echo "VPC_COUNT_CLEAN: $(cat -A <<< "$VPC_COUNT_CLEAN")"
echo "TF_VPC_COUNT_CLEAN: $(cat -A <<< "$TF_VPC_COUNT_CLEAN")"

separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator

# 6.1 直接表示法
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo "直接表示法:"
echo -e "VPC\t\t$VPC_COUNT_CLEAN\t$TF_VPC_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"

# 6.2 一時変数を使用
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VPC_COUNT_CLEAN\t$TF_VPC_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo "一時変数を使用:"
echo -e "$DISPLAY_LINE"

# 6.3 printf使用
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo "printf使用:"
printf "VPC\t\t%s\t%s\t${STATUS_COLOR}%s${NC}\n" "$VPC_COUNT_CLEAN" "$TF_VPC_COUNT_CLEAN" "$STATUS"

# テスト7: 最終的な解決策のテスト
header "テスト7: 推奨する最終的な解決策のテスト"

# 改行を含む入力テスト
API_SERVICE_COUNT=$(echo -e "0\n")    # 意図的に改行を含む
TF_API_SERVICE_COUNT=$(echo "0")       # 改行なし

# サニタイズ
API_SERVICE_COUNT_CLEAN=$(enhanced_trim "$API_SERVICE_COUNT")
TF_API_SERVICE_COUNT_CLEAN=$(enhanced_trim "$TF_API_SERVICE_COUNT")

# 処理結果の確認
echo "サニタイズ前:"
echo "API_SERVICE_COUNT: $(cat -A <<< "$API_SERVICE_COUNT")"
echo "TF_API_SERVICE_COUNT: $(cat -A <<< "$TF_API_SERVICE_COUNT")"

echo "サニタイズ後:"
echo "API_SERVICE_COUNT_CLEAN: $(cat -A <<< "$API_SERVICE_COUNT_CLEAN")"
echo "TF_API_SERVICE_COUNT_CLEAN: $(cat -A <<< "$TF_API_SERVICE_COUNT_CLEAN")"

# 比較と表示
echo "推奨される最終的な表示方法:"
separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator

if [ "$API_SERVICE_COUNT_CLEAN" = "$TF_API_SERVICE_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi

# 安全な表示方法（推奨）
DISPLAY_LINE="APIサービス\t$API_SERVICE_COUNT_CLEAN\t$TF_API_SERVICE_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

echo ""
echo "結論: enhanced_trim関数と一時変数を組み合わせた方法が最も安全です"