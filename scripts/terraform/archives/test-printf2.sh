#!/bin/bash
# error-simulation-test.sh - エラーケースを意図的に再現するテスト

# 色の設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=== エラー再現テスト ==="

# 1. 通常の比較（正常ケース）
VAR1="0"
VAR2="0"
echo "1. 通常の比較（VAR1=0, VAR2=0）"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"

# 2. 改行を含む変数での比較
echo ""
echo "2. 改行を含む変数での比較"
VAR1="0"
VAR2="0
"  # VAR2に意図的に改行を追加
echo "VAR1='$VAR1', VAR2='$VAR2'"
echo "VAR1の長さ: ${#VAR1}, VAR2の長さ: ${#VAR2}"

if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致" 
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"

# 3. 現在のaws-terraform-verify.shで使用されている表示方法の模倣
echo ""
echo "3. aws-terraform-verify.shの表示方法の模倣"
VAR1="0"
VAR2="0"
echo "3.1 通常のケース（VAR1=0, VAR2=0）"
echo "--------------------------------"
echo -e "リソース\t\tAWS\tTerraform\t状態"
echo "--------------------------------"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"

echo ""
echo "3.2 改行を含むケース（VAR1=0, VAR2=0+改行）"
VAR1="0"
VAR2="0
"
echo "--------------------------------"
echo -e "リソース\t\tAWS\tTerraform\t状態"
echo "--------------------------------"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"

# 4. aws-terraform-verify.shと同様の変数処理を模倣
echo ""
echo "4. aws-terraform-verify.sh の変数処理を模倣"

# トリム関数
trim() {
  local var="$*"
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  echo -n "$var"
}

# コマンド出力をシミュレート
AWS_OUTPUT="0
"  # 意図的に改行を含む出力を模倣
TF_OUTPUT="0"

# トリム処理
AWS_CLEAN=$(trim "$AWS_OUTPUT")
TF_CLEAN=$(trim "$TF_OUTPUT")

echo "AWS_OUTPUT='$AWS_OUTPUT', TF_OUTPUT='$TF_OUTPUT'"
echo "AWS_CLEAN='$AWS_CLEAN', TF_CLEAN='$TF_CLEAN'"
echo "AWS_CLEANの長さ: ${#AWS_CLEAN}, TF_CLEANの長さ: ${#TF_CLEAN}"

echo "--------------------------------"
echo -e "リソース\t\tAWS\tTerraform\t状態"
echo "--------------------------------"
if [ "$AWS_CLEAN" = "$TF_CLEAN" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi
echo -e "VPC\t\t$AWS_CLEAN\t$TF_CLEAN\t${STATUS_COLOR}$STATUS${NC}"

# 5. 一時変数を使った表示方法のテスト
echo ""
echo "5. 一時変数を使った表示方法のテスト"
VAR1="0"
VAR2="0
"  # 改行を含む
echo "--------------------------------"
echo -e "リソース\t\tAWS\tTerraform\t状態"
echo "--------------------------------"
if [ "$VAR1" = "$VAR2" ]; then
  STATUS="一致"
  STATUS_COLOR=$GREEN
else
  STATUS="不一致"
  STATUS_COLOR=$RED
fi

# 一時変数に格納してから表示
DISPLAY_LINE="VPC\t\t$VAR1\t$VAR2\t${STATUS_COLOR}$STATUS${NC}"
echo -e "$DISPLAY_LINE"

# 6. さまざまな組み合わせのテスト
echo ""
echo "6. さまざまな組み合わせのテスト"

# トリム関数の改良版
enhanced_trim() {
  local var="$*"
  # 改行を削除
  var=$(echo -n "$var" | tr -d '\n' | tr -d '\r')
  # 先頭と末尾の空白を削除
  var="${var#"${var%%[![:space:]]*}"}"
  var="${var%"${var##*[![:space:]]}"}"
  echo -n "$var"
}

# テストケース
VAR1="0
"  # 改行あり
VAR2="0"

# 通常のtrim
VAR1_CLEAN=$(trim "$VAR1")
VAR2_CLEAN=$(trim "$VAR2")

# 改良版trim
VAR1_ENHANCED=$(enhanced_trim "$VAR1")
VAR2_ENHANCED=$(enhanced_trim "$VAR2")

echo "6.1 通常のtrim関数"
echo "VAR1_CLEAN='$VAR1_CLEAN', VAR2_CLEAN='$VAR2_CLEAN'"
echo "VAR1_CLEANの長さ: ${#VAR1_CLEAN}, VAR2_CLEANの長さ: ${#VAR2_CLEAN}"

echo "6.2 改良版trim関数"
echo "VAR1_ENHANCED='$VAR1_ENHANCED', VAR2_ENHANCED='$VAR2_ENHANCED'"
echo "VAR1_ENHANCEDの長さ: ${#VAR1_ENHANCED}, VAR2_ENHANCEDの長さ: ${#VAR2_ENHANCED}"