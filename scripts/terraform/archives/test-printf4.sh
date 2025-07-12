#!/bin/bash
# verify-display-methods.sh - 変数表示問題の詳細検証スクリプト

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

# 強化版trim関数（改行と特殊文字を徹底的に削除）
enhanced_trim() {
  local var="$*"
  # まず改行とキャリッジリターンを削除
  var=$(echo -n "$var" | tr -d '\n' | tr -d '\r')
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  echo -n "$var"
}

# 変数の詳細情報を表示する関数
debug_var() {
  local name="$1"
  local value="$2"
  echo "📊 $name:"
  echo "  内容 (cat -A): $(cat -A <<< "$value")"
  echo "  長さ: ${#value}"
  echo "  16進ダンプ: $(hexdump -C <<< "$value" | head -1)"
}

header "仮説1: テスト2とテスト5の違いを検証"

echo "🔹 テスト2シナリオ（単一の改行含む変数）の詳細検証"
VAR1="0"
VAR2="0
"  # VAR2に意図的に改行を追加

debug_var "VAR1" "$VAR1"
debug_var "VAR2" "$VAR2"

separator
echo -e "リソース\t\tAWS\tTerraform\t状態"
separator

echo "🔸 直接表示:"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致" 
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"

echo "🔸 一時変数使用:"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致" 
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

echo "🔸 printf使用:"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致" 
  STATUS_COLOR=$RED
fi
printf "VPC\t\t%s\t%s\t${STATUS_COLOR}%s${NC}\n" "$VAR1" "$VAR2" "$STATUS"

echo "🔹 テスト5シナリオ（複数の改行含む変数）の詳細検証"
VAR1="0
"  # 改行あり
VAR2="0
"  # 改行あり
VAR3="0"   # 改行なし

debug_var "VAR1" "$VAR1"
debug_var "VAR2" "$VAR2"
debug_var "VAR3" "$VAR3"

separator
echo -e "リソース\t\tAWS\tTerraform\t追加\t状態"
separator

echo "🔸 直接表示:"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VAR1\t$VAR2\t$VAR3\t${STATUS_COLOR}$STATUS${NC}"

echo "🔸 一時変数使用:"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VAR1\t$VAR2\t$VAR3\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

echo "🔸 printf使用:"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
printf "VPC\t\t%s\t%s\t%s\t${STATUS_COLOR}%s${NC}\n" "$VAR1" "$VAR2" "$VAR3" "$STATUS"

header "仮説2: 変数のクリーニング段階別検証"

echo "🔹 改行含む変数の各処理段階確認"
ORIG_VAR="0
"  # 改行あり

debug_var "元の変数" "$ORIG_VAR"

# 基本的なtrim
TRIMMED_VAR=$(trim "$ORIG_VAR")
debug_var "基本trim後" "$TRIMMED_VAR"

# 強化版trim
ENHANCED_VAR=$(enhanced_trim "$ORIG_VAR")
debug_var "強化版trim後" "$ENHANCED_VAR"

# echo -n で処理
ECHO_VAR=$(echo -n "$ORIG_VAR")
debug_var "echo -n で処理" "$ECHO_VAR"

# tr で改行削除
TR_VAR=$(echo "$ORIG_VAR" | tr -d '\n')
debug_var "tr で改行削除" "$TR_VAR"

# 複合処理
COMPOUND_VAR=$(echo -n "$(trim "$ORIG_VAR")")
debug_var "trim + echo -n" "$COMPOUND_VAR"

header "仮説3: 表示方法の違いによる影響"

echo "🔹 同じ変数での表示方法比較"

TEST_VAR="0
"  # 改行あり
CLEAN_VAR=$(enhanced_trim "$TEST_VAR")

debug_var "元の変数" "$TEST_VAR"
debug_var "クリーン変数" "$CLEAN_VAR"

separator
echo "🔸 元の変数の表示方法比較:"
separator

echo "1) echo:"
echo "$TEST_VAR"

echo "2) echo -n:"
echo -n "$TEST_VAR"
echo ""  # 改行追加

echo "3) echo -e:"
echo -e "$TEST_VAR"

echo "4) printf:"
printf "%s\n" "$TEST_VAR"

echo "5) cat:"
cat <<< "$TEST_VAR"

separator
echo "🔸 クリーン変数の表示方法比較:"
separator

echo "1) echo:"
echo "$CLEAN_VAR"

echo "2) echo -n:"
echo -n "$CLEAN_VAR"
echo ""  # 改行追加

echo "3) echo -e:"
echo -e "$CLEAN_VAR"

echo "4) printf:"
printf "%s\n" "$CLEAN_VAR"

echo "5) cat:"
cat <<< "$CLEAN_VAR"

header "仮説4: タブ文字と色コードの影響"

echo "🔹 タブと色コードの組み合わせテスト"

VAR_WITH_NL="0
"  # 改行あり
CLEAN_VAR=$(enhanced_trim "$VAR_WITH_NL")

echo "🔸 タブなし、色なし:"
echo "VPC $CLEAN_VAR $CLEAN_VAR 一致"

echo "🔸 タブあり、色なし:"
echo -e "VPC\t\t$CLEAN_VAR\t$CLEAN_VAR\t一致"

echo "🔸 タブなし、色あり:"
echo "VPC $CLEAN_VAR $CLEAN_VAR ${GREEN}一致${NC}"

echo "🔸 タブあり、色あり:"
echo -e "VPC\t\t$CLEAN_VAR\t$CLEAN_VAR\t${GREEN}一致${NC}"

echo "🔸 タブあり、色あり、一時変数使用:"
DISPLAY_LINE="VPC\t\t$CLEAN_VAR\t$CLEAN_VAR\t${GREEN}一致${NC}"
echo -e "$DISPLAY_LINE"

echo "🔸 タブあり、色あり、printf使用:"
printf "VPC\t\t%s\t%s\t${GREEN}%s${NC}\n" "$CLEAN_VAR" "$CLEAN_VAR" "一致"

header "仮説5: 変数展開の順序の影響"

echo "🔹 変数展開の順序テスト"

VAR_NL="0
"  # 改行あり
CLEAN_VAR=$(enhanced_trim "$VAR_NL")

echo "🔸 通常の順序（変数→echo）:"
echo -e "VPC\t\t$VAR_NL\t$CLEAN_VAR\t一致"

echo "🔸 先に文字列を構成:"
VAR_STR="VPC\t\t$VAR_NL\t$CLEAN_VAR\t一致"
echo -e "$VAR_STR"

echo "🔸 先に文字列を構成（クリーンな変数のみ）:"
VAR_STR="VPC\t\t$CLEAN_VAR\t$CLEAN_VAR\t一致"
echo -e "$VAR_STR"

header "結論: 最も安全な方法の検証"

echo "🔹 実際のaws-terraform-verify.shでの最適な方法"

# AWSコマンド出力をシミュレート
VPC_COUNT=$(echo -e "0\n")  # 改行含む
TF_VPC_COUNT=$(echo "0")    # 改行なし

# クリーニング処理
VPC_COUNT_CLEAN=$(enhanced_trim "$VPC_COUNT")
TF_VPC_COUNT_CLEAN=$(enhanced_trim "$TF_VPC_COUNT")

debug_var "VPC_COUNT" "$VPC_COUNT"
debug_var "TF_VPC_COUNT" "$TF_VPC_COUNT"
debug_var "VPC_COUNT_CLEAN" "$VPC_COUNT_CLEAN"
debug_var "TF_VPC_COUNT_CLEAN" "$TF_VPC_COUNT_CLEAN"

echo "🔸 方法1: 直接echoで表示:"
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VPC_COUNT_CLEAN\t$TF_VPC_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"

echo "🔸 方法2: 一時変数使用:"
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VPC_COUNT_CLEAN\t$TF_VPC_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

echo "🔸 方法3: printf使用:"
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
printf "VPC\t\t%s\t%s\t${STATUS_COLOR}%s${NC}\n" "$VPC_COUNT_CLEAN" "$TF_VPC_COUNT_CLEAN" "$STATUS"

echo "🔸 方法4: 複合的なアプローチ（全てのステップを分離）:"
# 1. 変数の取得と処理を分離
RAW_VAR1=$(echo -e "0\n")  # AWSコマンドの出力を模倣
RAW_VAR2="0"                # Terraformの変数を模倣

# 2. 徹底的なクリーニング
CLEAN_VAR1=$(enhanced_trim "$RAW_VAR1")
CLEAN_VAR2=$(enhanced_trim "$RAW_VAR2")

# 3. 比較処理を分離
if [ "$CLEAN_VAR1" = "$CLEAN_VAR2" ]; then
  CMP_STATUS="一致"
  CMP_COLOR=$GREEN
else
  CMP_STATUS="不一致"
  CMP_COLOR=$RED
fi

# 4. 表示処理も分離（一時変数使用）
DISPLAY_STR="VPC\t\t$CLEAN_VAR1\t$CLEAN_VAR2\t${CMP_COLOR}$CMP_STATUS${NC}"
echo -e "$DISPLAY_STR"

header "実際のスクリプト修正案のテスト"

echo "🔹 aws-terraform-verify.sh向けの最終修正案テスト"

# 実際のスクリプト中の処理に近い流れでテスト
VPC_COUNT=$(echo -e "0\n")  # 改行を含む模擬AWS CLI出力
TF_VPC_COUNT="0"           # Terraform出力（改行なし）

VPC_COUNT_CLEAN=$(trim "$VPC_COUNT")
TF_VPC_COUNT_CLEAN=$(trim "$TF_VPC_COUNT")

debug_var "VPC_COUNT" "$VPC_COUNT"
debug_var "TF_VPC_COUNT" "$TF_VPC_COUNT"
debug_var "VPC_COUNT_CLEAN" "$VPC_COUNT_CLEAN"
debug_var "TF_VPC_COUNT_CLEAN" "$TF_VPC_COUNT_CLEAN"

# 修正前
echo "🔸 修正前（直接echo）:"
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VPC_COUNT_CLEAN\t$TF_VPC_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"

# 修正案
echo "🔸 修正案（一時変数使用）:"
if [ "$VPC_COUNT_CLEAN" = "$TF_VPC_COUNT_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
DISPLAY_LINE="VPC\t\t$VPC_COUNT_CLEAN\t$TF_VPC_COUNT_CLEAN\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

echo ""
echo "📝 このテスト結果を分析することで、最も確実な解決策を特定できます"